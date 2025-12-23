package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/hwang-fu/portlens/internal/capture"
	"github.com/hwang-fu/portlens/internal/output"
	"github.com/hwang-fu/portlens/internal/parser"
	"github.com/hwang-fu/portlens/internal/procfs"
)

var version = "dev"

// getDirection returns "in", "out", or "unknown" based on src/dst IPs.
func getDirection(srcIP, dstIP string, localIPs map[string]bool) string {
	srcLocal := localIPs[srcIP]
	dstLocal := localIPs[dstIP]

	if srcLocal && !dstLocal {
		return "out" // From local to remote
	}
	if !srcLocal && dstLocal {
		return "in" // From remote to local
	}
	// Both local (loopback) or both remote (shouldn't happen)
	return "unknown"
}

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

func main() {
	// Define flags
	var (
		interfaceName = flag.String("interface", "", "network interface to capture on")
		protocol      = flag.String("protocol", "all", "protocol to capture: tcp, udp, or all")
		showVersion   = flag.Bool("version", false, "show version and exit")
		port          = flag.Int("port", 0, "filter by port number (0 = all ports)")
		ip            = flag.String("ip", "", "filter by IP address (empty = all IPs)")
		direction     = flag.String("direction", "all", "filter by direction: in, out, or all")
		process       = flag.String("process", "", "filter by process name")
		pid           = flag.Int("pid", 0, "filter by process ID")
	)

	// Short aliases
	flag.StringVar(interfaceName, "i", "", "network interface (shorthand)")
	flag.IntVar(port, "p", 0, "filter by port (shorthand)")

	flag.Parse()

	if *showVersion {
		fmt.Println("portlens", version)
		os.Exit(0)
	}

	if *interfaceName == "" {
		fmt.Fprintln(os.Stderr, "error: --interface (-i) is required")
		fmt.Fprintln(os.Stderr, "usage: portlens -i <interface> [--protocol tcp|udp|all]")
		fmt.Fprintln(os.Stderr, "example: sudo portlens -i lo")
		os.Exit(1)
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

	if err := sock.Bind(*interfaceName); err != nil {
		log.Fatalf("bind: %v", err)
	}

	fmt.Fprintf(os.Stderr, "capturing on %s...\n", *interfaceName)

	buf := make([]byte, 65535)
	for {
		n, err := sock.ReadPacket(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		frame, err := parser.ParseEthernet(buf[:n])
		if err != nil {
			log.Printf("parse error: %v", err)
			continue
		}
		// Skip non-IPv4 packets by far
		if frame.EtherType != parser.EtherTypeIPv4 {
			continue
		}

		ipv4Packet, err := parser.ParseIPv4(frame.Payload)
		if err != nil {
			log.Printf("parse ipv4 error: %v", err)
			continue
		}

		// IP filter: check if src or dst IP matches the filtering ip (if provided)
		if *ip != "" && ipv4Packet.SrcIP.String() != *ip && ipv4Packet.DstIP.String() != *ip {
			continue
		}

		// Determine packet direction
		dir := getDirection(ipv4Packet.SrcIP.String(), ipv4Packet.DstIP.String(), localIPs)
		// Direction filter
		if *direction != "all" && dir != *direction {
			continue
		}

		switch ipv4Packet.Protocol {
		case parser.ProtocolTCP:
			if *protocol == "udp" {
				continue // Skip TCP when filtering for UDP only
			}

			tcpSegment, err := parser.ParseTCP(ipv4Packet.Payload)
			if err != nil {
				log.Printf("parse TCP error: %v", err)
				continue
			}

			// Port filter: check if src or dst port matches
			if *port != 0 && tcpSegment.SrcPort != uint16(*port) && tcpSegment.DstPort != uint16(*port) {
				continue
			}

			// Lookup process info
			proc := lookupProcess("tcp", ipv4Packet.SrcIP, ipv4Packet.DstIP, tcpSegment.SrcPort, tcpSegment.DstPort)

			// Process filters
			if *process != "" && (proc == nil || proc.Name != *process) {
				continue
			}
			if *pid != 0 && (proc == nil || proc.PID != *pid) {
				continue
			}

			record := output.PacketRecord{
				Timestamp: output.Now(),
				Protocol:  "TCP",
				SrcIP:     ipv4Packet.SrcIP.String(),
				SrcPort:   tcpSegment.SrcPort,
				DstIP:     ipv4Packet.DstIP.String(),
				DstPort:   tcpSegment.DstPort,
				Direction: dir,
				TCP: &output.TCPInfo{
					Seq:   tcpSegment.SeqNum,
					Ack:   tcpSegment.AckNum,
					Flags: parser.FormatFlags(tcpSegment.Flags),
				},
			}

			// Add process info if found
			if proc != nil {
				record.PID = proc.PID
				record.ProcessName = proc.Name
			}

			json.NewEncoder(os.Stdout).Encode(record)

		case parser.ProtocolUDP:
			if *protocol == "tcp" {
				continue // Skip UDP when filtering for TCP only
			}

			udpDatagram, err := parser.ParseUDP(ipv4Packet.Payload)
			if err != nil {
				log.Printf("parse UDP error: %v", err)
				continue
			}

			// Port filter: check if src or dst port matches
			if *port != 0 && udpDatagram.SrcPort != uint16(*port) && udpDatagram.DstPort != uint16(*port) {
				continue
			}

			// Lookup process info
			proc := lookupProcess("udp", ipv4Packet.SrcIP, ipv4Packet.DstIP, udpDatagram.SrcPort, udpDatagram.DstPort)

			// Process filters
			if *process != "" && (proc == nil || proc.Name != *process) {
				continue
			}
			if *pid != 0 && (proc == nil || proc.PID != *pid) {
				continue
			}

			record := output.PacketRecord{
				Timestamp: output.Now(),
				Protocol:  "UDP",
				SrcIP:     ipv4Packet.SrcIP.String(),
				SrcPort:   udpDatagram.SrcPort,
				DstIP:     ipv4Packet.DstIP.String(),
				DstPort:   udpDatagram.DstPort,
				Direction: dir,
				UDP: &output.UDPInfo{
					Length: udpDatagram.Length,
				},
			}

			// Add process info if found (ADD THIS)
			if proc != nil {
				record.PID = proc.PID
				record.ProcessName = proc.Name
			}

			json.NewEncoder(os.Stdout).Encode(record)
		default:
			// Skip non-TCP/UDP packets
			continue
		}
	}
}
