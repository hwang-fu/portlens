package parser

const (
	UDPHeaderSize = 8
)

// UDPDatagram represents a parsed UDP header.
// UDP is a connectionless, stateless protocol - each datagram is independent.
type UDPDatagram struct {
	SrcPort uint16
	DstPort uint16
	Length  uint16
	Payload []byte
}
