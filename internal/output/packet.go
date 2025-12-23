package output

// PacketRecord represents a captured packet in JSON-serializable format.
type PacketRecord struct {
	Timestamp string `json:"timestamp"`
	Protocol  string `json:"protocol"`
	SrcIP     string `json:"src_ip"`
	SrcPort   uint16 `json:"src_port"`
	DstIP     string `json:"dst_ip"`
	DstPort   uint16 `json:"dst_port"`

	// Protocol-specific fields (only one will be set)
	TCP *TCPInfo `json:"tcp,omitempty"`
	UDP *UDPInfo `json:"udp,omitempty"`
}
