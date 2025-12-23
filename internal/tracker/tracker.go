package tracker

import (
	"sync"
	"time"
)

// Event represents a connection state change event.
type Event struct {
	Type       string // "opened", "closed", "state_change"
	Connection *Connection
	OldState   TCPState // Only for state_change events
	Timestamp  time.Time
}

// Tracker manages TCP connection state tracking.
type Tracker struct {
	mu          sync.RWMutex
	connections map[ConnKey]*Connection
	events      chan Event
}
