package parser

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	EthernetHeaderSize = 14 // in bytes

	EtherTypeIPv4 = uint16(0x0800)
	EtherTypeARP  = uint16(0x0806)
	EtherTypeIPv6 = uint16(0x86DD)
)

// EthernetFrame represents a parsed Ethernet frame header.
type EthernetFrame struct {
	DestMAC   net.HardwareAddr
	SrcMAC    net.HardwareAddr
	EtherType uint16
	Payload   []byte
}

// ParseEthernet parses raw bytes into an EthernetFrame.
func ParseEthernet(data []byte) (*EthernetFrame, error) {
	if len(data) < EthernetHeaderSize {
		return nil, fmt.Errorf("packet too short: %d bytes", len(data))
	}

	return &EthernetFrame{
		DestMAC:   net.HardwareAddr(data[0:6]),
		SrcMAC:    net.HardwareAddr(data[6:12]),
		EtherType: binary.BigEndian.Uint16(data[12:14]),
		Payload:   data[EthernetHeaderSize:],
	}, nil
}
