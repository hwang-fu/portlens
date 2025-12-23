package parser

import (
	"encoding/binary"
	"fmt"
)

const (
	UDPHeaderSize = 8 // in bytes
)

// UDPDatagram represents a parsed UDP header.
// UDP is a connectionless, stateless protocol - each datagram is independent.
type UDPDatagram struct {
	SrcPort uint16
	DstPort uint16
	Length  uint16 // Total length: header (8 bytes) + payload
	Payload []byte // Data after the UDP header
}

// ParseUDP parses raw bytes into a UDPDatagram.
// Returns an error if the data is too short to contain a valid UDP header.
func ParseUDP(data []byte) (*UDPDatagram, error) {
	if len(data) < UDPHeaderSize {
		return nil, fmt.Errorf("UDP packet too short: %d bytes", len(data))
	}

	// binary.BigEndian.Uint16(data) decodes big-endian bytes into a Go uint16.
	// binary.BigEndian.Uint16([0x1F, 0x90])
	// = 0x1F * 256 + 0x90
	// = 31 * 256 + 144
	// = 8080 <- the port number
	return &UDPDatagram{
		SrcPort: binary.BigEndian.Uint16(data[0:2]),
		DstPort: binary.BigEndian.Uint16(data[2:4]),
		Length:  binary.BigEndian.Uint16(data[4:6]),
		Payload: data[UDPHeaderSize:],
	}, nil
}
