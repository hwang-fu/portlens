package parser

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
