package main

import (
	"encoding/json"
	"io"
	"log"

	"github.com/hwang-fu/portlens/internal/output"
	"github.com/hwang-fu/portlens/internal/parser"
	"github.com/hwang-fu/portlens/internal/tracker"
)

// jsonWriter handles JSON output with pretty-printing.
type jsonWriter struct {
	w io.Writer
}

// Encode writes a value as pretty-printed JSON.
func (jw *jsonWriter) Encode(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = jw.w.Write(append(data, '\n'))
	return err
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
