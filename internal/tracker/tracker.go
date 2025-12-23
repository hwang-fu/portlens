package tracker

import "time"

// Event represents a connection state change event.
type Event struct {
	Type       string // "opened", "closed", "state_change"
	Connection *Connection
	OldState   TCPState // Only for state_change events
	Timestamp  time.Time
}
