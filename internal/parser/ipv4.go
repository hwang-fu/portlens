package parser

import "net"

const (
	IPv4MinHeaderSize = 20

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
