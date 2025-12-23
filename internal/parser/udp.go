package parser

const (
	UDPHeaderSize = 8
)

type UDPDatagram struct {
	SrcPort uint16
	DstPort uint16
	Length  uint16
	Payload []byte
}
