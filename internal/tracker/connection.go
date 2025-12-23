package tracker

import (
	"fmt"
	"time"
)

// ConnKey uniquely identifies a connection.
//
// IMPORTANT: This is a normalized key, NOT the actual packet direction.
// Both directions of a connection map to the same key. For example:
//   - Packet A:1000 → B:80  → key = {A:1000, B:80}
//   - Packet B:80 → A:1000  → key = {A:1000, B:80} (same key!)
//
// Use NormalizeKey() to create a ConnKey from packet src/dst.
// The SrcIP/SrcPort fields represent the "lower" endpoint, not the sender.
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

// String returns the state name.
func (s TCPState) String() string {
	names := []string{
		"CLOSED",
		"SYN_SENT",
		"SYN_RECEIVED",
		"ESTABLISHED",
		"FIN_WAIT_1",
		"FIN_WAIT_2",
		"CLOSE_WAIT",
		"LAST_ACK",
		"TIME_WAIT",
	}
	if int(s) >= 0 && int(s) < len(names) {
		return names[s]
	}
	return "UNKNOWN"
}

// Connection tracks the state and statistics of a single connection.
type Connection struct {
	Key       ConnKey
	State     TCPState
	StartTime time.Time
	EndTime   time.Time

	// Statistics
	PacketsSent     uint64
	PacketsReceived uint64
	BytesSent       uint64
	BytesReceived   uint64
}

// Duration returns how long the connection has been active.
func (c *Connection) Duration() time.Duration {
	if c.EndTime.IsZero() {
		return time.Since(c.StartTime)
	}
	return c.EndTime.Sub(c.StartTime)
}
