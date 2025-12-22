package parser

import "net"

const (
	EthernetHeaderSize = 14 // in bytes

	EtherTypeIPv4 = uint16(0x0800)
	EtherTypeARP  = uint16(0x0806)
	EtherTypeIPv6 = uint16(0x86DD)
)

type EthernetFrame struct {
	DestMAC   net.HardwareAddr
	SrcMAC    net.HardwareAddr
	EtherType uint16
	Payload   []byte
}
