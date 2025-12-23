package parser

import "testing"

func TestParseTCP(t *testing.T) {
	// TCP header (20 bytes, no options) + 4 bytes payload
	data := []byte{
		0x1F, 0x90, // Source port: 8080
		0x00, 0x50, // Dest port: 80
		0x00, 0x00, 0x00, 0x01, // Sequence number: 1
		0x00, 0x00, 0x00, 0x02, // Ack number: 2
		0x50,       // Data offset: 5 (20 bytes), reserved: 0
		0x12,       // Flags: SYN + ACK (0x02 | 0x10)
		0x72, 0x10, // Window: 29200
		0x00, 0x00, // Checksum (ignored)
		0x00, 0x00, // Urgent pointer (ignored)
		0xde, 0xad, 0xbe, 0xef, // Payload
	}

	seg, err := ParseTCP(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if seg.SrcPort != 8080 {
		t.Errorf("SrcPort = %d, want 8080", seg.SrcPort)
	}

	if seg.DstPort != 80 {
		t.Errorf("DstPort = %d, want 80", seg.DstPort)
	}

	if seg.SeqNum != 1 {
		t.Errorf("SeqNum = %d, want 1", seg.SeqNum)
	}

	if seg.AckNum != 2 {
		t.Errorf("AckNum = %d, want 2", seg.AckNum)
	}

	if seg.DataOffset != 5 {
		t.Errorf("DataOffset = %d, want 5", seg.DataOffset)
	}

	if seg.Flags != (TCPFlagSYN | TCPFlagACK) {
		t.Errorf("Flags = 0x%02x, want 0x%02x", seg.Flags, TCPFlagSYN|TCPFlagACK)
	}

	if seg.Window != 29200 {
		t.Errorf("Window = %d, want 29200", seg.Window)
	}

	if len(seg.Payload) != 4 {
		t.Errorf("Payload length = %d, want 4", len(seg.Payload))
	}
}

func TestParseTCPTooShort(t *testing.T) {
	data := []byte{0x1F, 0x90, 0x00} // Only 3 bytes

	_, err := ParseTCP(data)
	if err == nil {
		t.Error("expected error for short segment, got nil")
	}
}
