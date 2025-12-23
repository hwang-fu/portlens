package parser

import (
	"encoding/binary"
	"fmt"
)

const (
	TCPMinHeaderSize = 20 // in bytes
	TCPMaxHeaderSize = 60 // in bytes (with options)
)

// TCP flag bitmasks (byte 13 of TCP header)
const (
	TCPFlagFIN = 0x01 // Finish - end of data
	TCPFlagSYN = 0x02 // Synchronize - initiate connection
	TCPFlagRST = 0x04 // Reset - abort connection
	TCPFlagPSH = 0x08 // Push - send data immediately
	TCPFlagACK = 0x10 // Acknowledgment
	TCPFlagURG = 0x20 // Urgent (rarely used)
)

// TCPSegment represents a parsed TCP header.
type TCPSegment struct {
	SrcPort    uint16
	DstPort    uint16
	SeqNum     uint32 // Sequence number
	AckNum     uint32 // Acknowledgment number
	DataOffset uint8  // Header length in 32-bit words (like IHL)
	Flags      uint8  // TCP flags (SYN, ACK, FIN, RST, etc.)
	Window     uint16 // Flow control window size
	Payload    []byte
}

// ParseTCP parses raw bytes into a TCPSegment.
// Returns an error if the data is too short or has invalid header length.
func ParseTCP(data []byte) (*TCPSegment, error) {
	if len(data) < TCPMinHeaderSize {
		return nil, fmt.Errorf("TCP segment too short: %d bytes", len(data))
	}

	// Data offset is the high 4 bits of byte 12
	// Similar to IPv4's IHL - gives header length in 32-bit words
	dataOffset := data[12] >> 4

	headerLen := int(dataOffset) * 4
	if headerLen < TCPMinHeaderSize {
		return nil, fmt.Errorf("invalid data offset, too small: %d", dataOffset)
	}
	if headerLen > TCPMaxHeaderSize {
		return nil, fmt.Errorf("invalid data offset, too large: %d", dataOffset)
	}
	if len(data) < headerLen {
		return nil, fmt.Errorf("TCP segment too short for header: %d < %d", len(data), headerLen)
	}

	return &TCPSegment{
		SrcPort:    binary.BigEndian.Uint16(data[0:2]),
		DstPort:    binary.BigEndian.Uint16(data[2:4]),
		SeqNum:     binary.BigEndian.Uint32(data[4:8]),
		AckNum:     binary.BigEndian.Uint32(data[8:12]),
		DataOffset: dataOffset,
		Flags:      data[13],
		Window:     binary.BigEndian.Uint16(data[14:16]),
		Payload:    data[headerLen:],
	}, nil
}
