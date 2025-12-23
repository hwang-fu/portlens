package tracker

import "fmt"

// ConnKey uniquely identifies a TCP connection (5-tuple).
// We normalize the key so that both directions map to the same connection.
type ConnKey struct {
	SrcIP    string
	SrcPort  uint16
	DstIP    string
	DstPort  uint16
	Protocol string // "TCP" or "UDP"
}

// String returns a human-readable representation of the connection key.
func (k ConnKey) String() string {
	return fmt.Sprintf("%s:%d -> %s:%d (%s)", k.SrcIP, k.SrcPort, k.DstIP, k.DstPort, k.Protocol)
}

// TCPState represents the state of a TCP connection.
type TCPState int

const (
	StateClosed TCPState = iota
	StateSynSent
	StateSynReceived
	StateEstablished
	StateFinWait1
	StateFinWait2
	StateCloseWait
	StateLastAck
	StateTimeWait
)
