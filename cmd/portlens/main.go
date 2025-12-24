package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hwang-fu/portlens/internal/capture"
	"github.com/hwang-fu/portlens/internal/parser"
	"github.com/hwang-fu/portlens/internal/stats"
)

var version = "dev"

var (
	cfg     config
	jsonOut *jsonWriter
)

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
	jsonOut = &jsonWriter{w: outWriter}

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

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		if cfg.graceful && statsRecorder != nil {
			fmt.Fprintln(os.Stderr, "\n--- Shutdown Summary ---")
			statsRecorder.WriteJSON(os.Stderr)
		}
		os.Exit(0)
	}()

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
