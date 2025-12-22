package parser

import (
	"testing"
)

func TestParseIPv4(t *testing.T) {
	// Minimal valid IPv4 header (20 bytes) + 4 bytes payload
	data := []byte{
		0x45,       // Version (4) + IHL (5)
		0x00,       // ToS
		0x00, 0x18, // Total Length (24 bytes)
		0x00, 0x00, // Identification
		0x00, 0x00, // Flags + Fragment Offset
		0x40,       // TTL (64)
		0x06,       // Protocol (TCP)
		0x00, 0x00, // Header Checksum
		0xc0, 0xa8, 0x01, 0x01, // Src IP: 192.168.1.1
		0xc0, 0xa8, 0x01, 0x02, // Dst IP: 192.168.1.2
		0xde, 0xad, 0xbe, 0xef, // Payload
	}

	pkt, err := ParseIPv4(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pkt.Version != 4 {
		t.Errorf("Version = %d, want 4", pkt.Version)
	}

	if pkt.IHL != 5 {
		t.Errorf("IHL = %d, want 5", pkt.IHL)
	}

	if pkt.TTL != 64 {
		t.Errorf("TTL = %d, want 64", pkt.TTL)
	}

	if pkt.Protocol != ProtocolTCP {
		t.Errorf("Protocol = %d, want %d", pkt.Protocol, ProtocolTCP)
	}

	if pkt.SrcIP.String() != "192.168.1.1" {
		t.Errorf("SrcIP = %s, want 192.168.1.1", pkt.SrcIP)
	}

	if pkt.DstIP.String() != "192.168.1.2" {
		t.Errorf("DstIP = %s, want 192.168.1.2", pkt.DstIP)
	}

	if len(pkt.Payload) != 4 {
		t.Errorf("Payload length = %d, want 4", len(pkt.Payload))
	}
}
