package parser

import "testing"

func TestParseEthernet(t *testing.T) {
	// Sample Ethernet frame: dst MAC, src MAC, EtherType (IPv4)
	data := []byte{
		0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Dest MAC
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66, // Src MAC
		0x08, 0x00, // EtherType: IPv4
		0xde, 0xad, 0xbe, 0xef, // Payload (dummy)
	}

	frame, err := ParseEthernet(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if frame.DestMAC.String() != "aa:bb:cc:d:ee:ff" {
		t.Errorf("DestMAC = %s, want aa:bb:cc:dd:ee:ff", frame.DestMAC)
	}

	if frame.SrcMAC.String() != "11:22:33:44:55:66" {
		t.Errorf("SrcMAC = %s, want 11:22:33:44:55:66", frame.SrcMAC)
	}

	if frame.EtherType != EtherTypeIPv4 {
		t.Errorf("EtherType = 0x%04x, want 0x0800", frame.EtherType)
	}

	if len(frame.Payload) != 4 {
		t.Errorf("Payload length = %d, want 4", len(frame.Payload))
	}
}

func TestParseEthernetTooShort(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02} // Only 3 bytes

	_, err := ParseEthernet(data)
	if err == nil {
		t.Error("expected error for short packet, got nil")
	}
}
