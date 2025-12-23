package tracker

// ConnKey uniquely identifies a TCP connection (5-tuple).
// We normalize the key so that both directions map to the same connection.
type ConnKey struct {
	SrcIP    string
	SrcPort  uint16
	DstIP    string
	DstPort  uint16
	Protocol string // "TCP" or "UDP"
}
