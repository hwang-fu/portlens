package parser

import "testing"

func TestParseUDP(t *testing.T) {
	// UDP header (8 bytes) + payload (4 bytes)
	data := []byte{
		0x1F, 0x90, // Source port: 8080
		0x00, 0x50, // Dest port: 80
		0x00, 0x0C, // Length: 12 (8 header + 4 payload)
		0x00, 0x00, // Checksum (ignored)
		0xde, 0xad, 0xbe, 0xef, // Payload
	}

	datagram, err := ParseUDP(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if datagram.SrcPort != 8080 {
		t.Errorf("SrcPort = %d, want 8080", datagram.SrcPort)
	}

	if datagram.DstPort != 80 {
		t.Errorf("DstPort = %d, want 80", datagram.DstPort)
	}

	if datagram.Length != 12 {
		t.Errorf("Length = %d, want 12", datagram.Length)
	}

	if len(datagram.Payload) != 4 {
		t.Errorf("Payload length = %d, want 4", len(datagram.Payload))
	}
}

func TestParseUDPTooShort(t *testing.T) {
	data := []byte{0x1F, 0x90, 0x00} // Only 3 bytes

	_, err := ParseUDP(data)
	if err == nil {
		t.Error("expected error for short packet, got nil")
	}
}
