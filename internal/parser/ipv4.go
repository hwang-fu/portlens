package parser

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	IPv4MinHeaderSize = 20 // in bytes
	IPv4MaxHeaderSize = 60 // in bytes

	ProtocolTCP = 6
	ProtocolUDP = 17
)

// IPv4Packet represents a parsed IPv4 header.
type IPv4Packet struct {
	Version  uint8
	IHL      uint8 // Header length in 32-bit words
	TotalLen uint16
	TTL      uint8
	Protocol uint8
	SrcIP    net.IP
	DstIP    net.IP
	Payload  []byte
}

// ParseIPv4 parses raw bytes into an IPv4Packet.
func ParseIPv4(data []byte) (*IPv4Packet, error) {
	if len(data) < IPv4MinHeaderSize {
		return nil, fmt.Errorf("packet too short: %d bytes", len(data))
	}
	if len(data) > IPv4MaxHeaderSize {
		return nil, fmt.Errorf("packet too large: %d bytes", len(data))
	}

	versionIHL := data[0]
	version := versionIHL >> 4
	ihl := versionIHL & 0x0F

	if version != 4 {
		return nil, fmt.Errorf("not IPv4: version %d", version)
	}

	headerLen := int(ihl) * 4
	if headerLen < IPv4MinHeaderSize {
		return nil, fmt.Errorf("invalid IHL: %d", ihl)
	}
	if len(data) < headerLen {
		return nil, fmt.Errorf("packet too short for header: %d < %d", len(data), headerLen)
	}

	return &IPv4Packet{
		Version:  version,
		IHL:      ihl,
		TotalLen: binary.BigEndian.Uint16(data[2:4]),
		TTL:      data[8],
		Protocol: data[9],
		SrcIP:    net.IP(data[12:16]),
		DstIP:    net.IP(data[16:20]),
		Payload:  data[headerLen:],
	}, nil
}
