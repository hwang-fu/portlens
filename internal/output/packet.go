package output

import "time"

// PacketRecord represents a captured packet in JSON-serializable format.
type PacketRecord struct {
	Timestamp string `json:"timestamp"`
	Protocol  string `json:"protocol"`
	SrcIP     string `json:"src_ip"`
	SrcPort   uint16 `json:"src_port"`
	DstIP     string `json:"dst_ip"`
	DstPort   uint16 `json:"dst_port"`
	Direction string `json:"direction"` // "in", "out", or "unknown"

	// Process info (may be empty if not found)
	PID         int    `json:"pid,omitempty"`
	ProcessName string `json:"process,omitempty"`

	// Protocol-specific fields (only one will be set)
	TCP *TCPInfo `json:"tcp,omitempty"`
	UDP *UDPInfo `json:"udp,omitempty"`
}

// TCPInfo contains TCP-specific fields.
type TCPInfo struct {
	Seq   uint32 `json:"seq"`
	Ack   uint32 `json:"ack"`
	Flags string `json:"flags"`
}

// UDPInfo contains UDP-specific fields.
type UDPInfo struct {
	Length uint16 `json:"length"`
}

// PayloadInfo contains payload preview for verbose output.
type PayloadInfo struct {
	Size int    `json:"size"`           // Total payload size in bytes
	Head string `json:"head,omitempty"` // First 64 bytes as hex
	Tail string `json:"tail,omitempty"` // Last 64 bytes as hex (if different from head)
}

// Now returns the current time formatted as ISO 8601 with milliseconds.
func Now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}
