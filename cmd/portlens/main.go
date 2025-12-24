package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/hwang-fu/portlens/internal/capture"
	yamlconfig "github.com/hwang-fu/portlens/internal/config"
	"github.com/hwang-fu/portlens/internal/output"
	"github.com/hwang-fu/portlens/internal/parser"
	"github.com/hwang-fu/portlens/internal/procfs"
	"github.com/hwang-fu/portlens/internal/stats"
	"github.com/hwang-fu/portlens/internal/tracker"
)

var version = "dev"

// config holds all runtime configuration from flags.
type config struct {
	interfaceName string
	protocol      string
	port          int
	ip            string
	direction     string
	process       string
	pid           int
	stateful      bool
	verbosity     int    // 0=minimal, 1=normal, 2=detailed, 3=verbose
	outputFile    string // output file path (empty = stdout)
	debug         bool   // enable debug logging
	logFile       string // log file path (empty = stderr)
	configFile    string // config file path
	stats         bool   // show performance statistics
	graceful      bool   // enable graceful shutdown with summary
}

var (
	cfg     config
	jsonOut *json.Encoder
)

func parseFlags() {
	// Check for config file flag first (manual parse)
	configPath := yamlconfig.DefaultPath()
	for i, arg := range os.Args[1:] {
		if arg == "-c" || arg == "--config" {
			if i+1 < len(os.Args)-1 {
				configPath = os.Args[i+2]
			}
		}
	}

	// Load config file (if exists)
	fileCfg, err := yamlconfig.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Set defaults from file config
	cfg.configFile = configPath
	cfg.interfaceName = fileCfg.Interface
	cfg.protocol = fileCfg.Protocol
	cfg.port = fileCfg.Port
	cfg.ip = fileCfg.IP
	cfg.direction = fileCfg.Direction
	cfg.process = fileCfg.Process
	cfg.pid = fileCfg.PID
	cfg.stateful = fileCfg.Stateful
	cfg.verbosity = fileCfg.Verbosity
	cfg.outputFile = fileCfg.Output
	cfg.debug = fileCfg.Debug
	cfg.logFile = fileCfg.LogFile
	cfg.stats = fileCfg.Stats
	cfg.graceful = fileCfg.Graceful

	// Default verbosity if not set
	if cfg.verbosity == 0 {
		cfg.verbosity = 2
	}
	// Default protocol if not set
	if cfg.protocol == "" {
		cfg.protocol = "all"
	}
	// Default direction if not set
	if cfg.direction == "" {
		cfg.direction = "all"
	}

	flag.StringVar(&cfg.interfaceName, "interface", cfg.interfaceName, "network interface to capture on")
	flag.StringVar(&cfg.interfaceName, "i", cfg.interfaceName, "network interface (shorthand)")
	flag.StringVar(&cfg.protocol, "protocol", cfg.protocol, "protocol to capture: tcp, udp, or all")
	flag.IntVar(&cfg.port, "port", cfg.port, "filter by port number (0 = all ports)")
	flag.IntVar(&cfg.port, "p", cfg.port, "filter by port (shorthand)")
	flag.StringVar(&cfg.ip, "ip", cfg.ip, "filter by IP address (empty = all IPs)")
	flag.StringVar(&cfg.direction, "direction", cfg.direction, "filter by direction: in, out, or all")
	flag.StringVar(&cfg.process, "process", cfg.process, "filter by process name")
	flag.IntVar(&cfg.pid, "pid", cfg.pid, "filter by process ID")
	flag.BoolVar(&cfg.stateful, "stateful", cfg.stateful, "enable connection state tracking")
	flag.IntVar(&cfg.verbosity, "verbosity", cfg.verbosity, "output verbosity: 0=minimal, 1=normal, 2=detailed, 3=verbose")
	flag.IntVar(&cfg.verbosity, "v", cfg.verbosity, "verbosity level (shorthand)")
	flag.StringVar(&cfg.outputFile, "output", cfg.outputFile, "write output to file (default: stdout)")
	flag.StringVar(&cfg.outputFile, "o", cfg.outputFile, "output file (shorthand)")
	flag.BoolVar(&cfg.debug, "debug", cfg.debug, "enable debug logging")
	flag.StringVar(&cfg.logFile, "log-file", cfg.logFile, "write logs to file (default: stderr)")
	flag.StringVar(&cfg.configFile, "config", cfg.configFile, "config file path")
	flag.StringVar(&cfg.configFile, "c", cfg.configFile, "config file (shorthand)")
	flag.BoolVar(&cfg.stats, "stats", cfg.stats, "show performance statistics")
	flag.BoolVar(&cfg.graceful, "graceful", cfg.graceful, "enable graceful shutdown with summary")

	showVersion := flag.Bool("version", false, "show version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Println("portlens", version)
		os.Exit(0)
	}

	if cfg.interfaceName == "" {
		fmt.Fprintln(os.Stderr, "error: --interface (-i) is required")
		fmt.Fprintln(os.Stderr, "usage: portlens -i <interface> [--protocol tcp|udp|all]")
		fmt.Fprintln(os.Stderr, "example: sudo portlens -i lo")
		os.Exit(1)
	}
}

// logDebug logs a message only if debug mode is enabled.
func logDebug(format string, args ...any) {
	if cfg.debug {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// getDirection returns "in", "out", or "unknown" based on src/dst IPs.
func getDirection(srcIP, dstIP string, localIPs map[string]bool) string {
	srcLocal := localIPs[srcIP]
	dstLocal := localIPs[dstIP]

	if srcLocal && !dstLocal {
		return "out"
	}
	if !srcLocal && dstLocal {
		return "in"
	}
	return "unknown"
}

// lookupProcess finds the process owning a socket.
func lookupProcess(protocol string, srcIP, dstIP net.IP, srcPort, dstPort uint16) *procfs.ProcessInfo {
	inode, err := procfs.FindSocketInode(protocol, srcIP, srcPort, dstIP, dstPort)
	if err != nil || inode == 0 {
		return nil
	}

	proc, err := procfs.FindProcessBySocket(inode)
	if err != nil {
		return nil
	}
	return proc
}

// matchesProcessFilter checks if proc matches the configured process filters.
// Returns true if the packet should be processed, false if it should be skipped.
func matchesProcessFilter(proc *procfs.ProcessInfo) bool {
	if cfg.process != "" && (proc == nil || proc.Name != cfg.process) {
		return false
	}
	if cfg.pid != 0 && (proc == nil || proc.PID != cfg.pid) {
		return false
	}
	return true
}

// setupTracker creates a connection tracker and starts its event handler.
// Returns nil if stateful mode is disabled.
func setupTracker() *tracker.Tracker {
	if !cfg.stateful {
		return nil
	}

	t := tracker.New(100)

	go func() {
		for event := range t.Events() {
			eventRecord := map[string]any{
				"event_type": event.Type,
				"timestamp":  event.Timestamp.UTC().Format("2006-01-02T15:04:05.000Z"),
				"connection": map[string]any{
					"src_ip":       event.Connection.Key.SrcIP,
					"src_port":     event.Connection.Key.SrcPort,
					"dst_ip":       event.Connection.Key.DstIP,
					"dst_port":     event.Connection.Key.DstPort,
					"protocol":     event.Connection.Key.Protocol,
					"state":        event.Connection.State.String(),
					"duration":     event.Connection.Duration().String(),
					"packets_sent": event.Connection.PacketsSent,
					"packets_recv": event.Connection.PacketsReceived,
					"bytes_sent":   event.Connection.BytesSent,
					"bytes_recv":   event.Connection.BytesReceived,
				},
			}
			jsonOut.Encode(eventRecord)
		}
	}()

	return t
}

// handleTCPPacket processes a TCP packet and outputs the record.
// Returns false if the packet was filtered out.
func handleTCPPacket(ipv4 *parser.IPv4Packet, dir string, connTracker *tracker.Tracker) bool {
	tcp, err := parser.ParseTCP(ipv4.Payload)
	if err != nil {
		log.Printf("parse TCP error: %v", err)
		return false
	}

	// Port filter
	if cfg.port != 0 && tcp.SrcPort != uint16(cfg.port) && tcp.DstPort != uint16(cfg.port) {
		return false
	}

	// Process lookup and filter
	proc := lookupProcess("tcp", ipv4.SrcIP, ipv4.DstIP, tcp.SrcPort, tcp.DstPort)
	if !matchesProcessFilter(proc) {
		return false
	}

	// Connection tracking
	if connTracker != nil {
		connTracker.ProcessTCPPacket(
			ipv4.SrcIP.String(), tcp.SrcPort,
			ipv4.DstIP.String(), tcp.DstPort,
			tcp.Flags,
			len(tcp.Payload),
			dir == "out",
		)
	}

	// Build and output record
	record := output.PacketRecord{
		Timestamp: output.Now(),
		Protocol:  "TCP",
		SrcIP:     ipv4.SrcIP.String(),
		SrcPort:   tcp.SrcPort,
		DstIP:     ipv4.DstIP.String(),
		DstPort:   tcp.DstPort,
		Direction: dir,
		TCP: &output.TCPInfo{
			Seq:   tcp.SeqNum,
			Ack:   tcp.AckNum,
			Flags: parser.FormatFlags(tcp.Flags),
		},
	}
	if proc != nil {
		record.PID = proc.PID
		record.ProcessName = proc.Name
	}

	if cfg.verbosity >= 3 {
		record.Payload = output.NewPayloadInfo(tcp.Payload)
	}

	if cfg.verbosity >= 2 {
		jsonOut.Encode(record)
	}

	return true
}

// handleUDPPacket processes a UDP packet and outputs the record.
// Returns false if the packet was filtered out.
func handleUDPPacket(ipv4 *parser.IPv4Packet, dir string) bool {
	udp, err := parser.ParseUDP(ipv4.Payload)
	if err != nil {
		log.Printf("parse UDP error: %v", err)
		return false
	}

	// Port filter
	if cfg.port != 0 && udp.SrcPort != uint16(cfg.port) && udp.DstPort != uint16(cfg.port) {
		return false
	}

	// Process lookup and filter
	proc := lookupProcess("udp", ipv4.SrcIP, ipv4.DstIP, udp.SrcPort, udp.DstPort)
	if !matchesProcessFilter(proc) {
		return false
	}

	// Build and output record
	record := output.PacketRecord{
		Timestamp: output.Now(),
		Protocol:  "UDP",
		SrcIP:     ipv4.SrcIP.String(),
		SrcPort:   udp.SrcPort,
		DstIP:     ipv4.DstIP.String(),
		DstPort:   udp.DstPort,
		Direction: dir,
		UDP: &output.UDPInfo{
			Length: udp.Length,
		},
	}
	if proc != nil {
		record.PID = proc.PID
		record.ProcessName = proc.Name
	}

	if cfg.verbosity >= 3 {
		record.Payload = output.NewPayloadInfo(udp.Payload)
	}

	if cfg.verbosity >= 2 {
		jsonOut.Encode(record)
	}
	return true
}

func main() {
	parseFlags()

	// Setup log output
	if cfg.logFile != "" {
		f, err := os.OpenFile(cfg.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// Setup output destination
	outWriter := os.Stdout
	if cfg.outputFile != "" {
		f, err := os.Create(cfg.outputFile)
		if err != nil {
			log.Fatalf("create output file: %v", err)
		}
		defer f.Close()
		outWriter = f
	}
	jsonOut = json.NewEncoder(outWriter)

	connTracker := setupTracker()
	if connTracker != nil {
		defer connTracker.Close()
	}

	localIPs, err := capture.LocalIPs()
	if err != nil {
		log.Fatalf("get local IPs: %v", err)
	}

	sock, err := capture.NewSocket()
	if err != nil {
		log.Fatalf("create socket: %v", err)
	}
	defer sock.Close()

	if err := sock.Bind(cfg.interfaceName); err != nil {
		log.Fatalf("bind: %v", err)
	}

	logDebug("config: interface=%s, protocol=%s, verbosity=%d", cfg.interfaceName, cfg.protocol, cfg.verbosity)

	fmt.Fprintf(os.Stderr, "capturing on %s...\n", cfg.interfaceName)

	// Setup stats recorder
	var statsRecorder *stats.StatsRecorder
	if cfg.stats {
		statsRecorder = stats.NewRecorder()
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				statsRecorder.WriteJSON(os.Stderr)
			}
		}()
	}

	buf := make([]byte, 65535)
	for {
		n, err := sock.ReadPacket(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		if statsRecorder != nil {
			statsRecorder.RecordPacket(n)
		}

		frame, err := parser.ParseEthernet(buf[:n])
		if err != nil {
			log.Printf("parse error: %v", err)
			continue
		}

		if frame.EtherType != parser.EtherTypeIPv4 {
			continue
		}

		ipv4, err := parser.ParseIPv4(frame.Payload)
		if err != nil {
			log.Printf("parse ipv4 error: %v", err)
			continue
		}

		// IP filter
		if cfg.ip != "" && ipv4.SrcIP.String() != cfg.ip && ipv4.DstIP.String() != cfg.ip {
			continue
		}

		// Direction filter
		dir := getDirection(ipv4.SrcIP.String(), ipv4.DstIP.String(), localIPs)
		if cfg.direction != "all" && dir != cfg.direction {
			continue
		}

		// Protocol handling
		switch ipv4.Protocol {
		case parser.ProtocolTCP:
			if cfg.protocol != "udp" {
				handleTCPPacket(ipv4, dir, connTracker)
			}
		case parser.ProtocolUDP:
			if cfg.protocol != "tcp" {
				handleUDPPacket(ipv4, dir)
			}
		}
	}
}
