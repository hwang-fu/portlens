package tracker

import (
	"sync"
	"time"
)

// Event represents a connection state change event.
type Event struct {
	Type       string   // "opened", "closed", "state_change"
	OldState   TCPState // Only for state_change events
	Connection *Connection
	Timestamp  time.Time
}

// Tracker manages TCP connection state tracking.
type Tracker struct {
	mu          sync.RWMutex
	connections map[ConnKey]*Connection
	events      chan Event
}

// New creates a new connection tracker.
// eventBufferSize determines how many events can be buffered before blocking.
func New(eventBufferSize int) *Tracker {
	return &Tracker{
		connections: make(map[ConnKey]*Connection),
		events:      make(chan Event, eventBufferSize),
	}
}

// Events returns the channel for receiving connection events.
func (t *Tracker) Events() <-chan Event {
	return t.events
}

// GetConnection returns an existing connection or nil if not found.
func (t *Tracker) GetConnection(key ConnKey) *Connection {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connections[key]
}

// getOrCreateConnection returns an existing connection or creates a new one.
// Caller must hold the write lock.
func (t *Tracker) getOrCreateConnection(key ConnKey) (*Connection, bool) {
	if conn, exists := t.connections[key]; exists {
		return conn, false
	}

	conn := &Connection{
		Key:       key,
		State:     StateClosed,
		StartTime: time.Now(),
	}
	t.connections[key] = conn
	return conn, true
}

// removeConnection removes a connection from tracking.
// Caller must hold the write lock.
func (t *Tracker) removeConnection(key ConnKey) {
	delete(t.connections, key)
}
