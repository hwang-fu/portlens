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

// emitEvent sends an event to the events channel (non-blocking).
func (t *Tracker) emitEvent(event Event) {
	select {
	case t.events <- event:
	default:
		// Channel full, drop event (could log warning here)
	}
}

// ActiveConnections returns the number of currently tracked connections.
func (t *Tracker) ActiveConnections() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.connections)
}

// Close closes the events channel.
func (t *Tracker) Close() {
	close(t.events)
}

// ProcessTCPPacket processes a TCP packet and updates connection state.
// Returns the connection and any state change event.
func (t *Tracker) ProcessTCPPacket(
	srcIP string, srcPort uint16,
	dstIP string, dstPort uint16,
	flags uint8,
	payloadLen int,
	isOutbound bool,
) *Connection {
	// TCP flag constants (should match parser package)
	const (
		FlagFIN = 0x01
		FlagSYN = 0x02
		FlagRST = 0x04
		FlagACK = 0x10
	)

	key := ConnKey{
		SrcIP:    srcIP,
		SrcPort:  srcPort,
		DstIP:    dstIP,
		DstPort:  dstPort,
		Protocol: "TCP",
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	conn, isNew := t.getOrCreateConnection(key)
	oldState := conn.State

	// Update statistics
	if isOutbound {
		conn.PacketsSent++
		conn.BytesSent += uint64(payloadLen)
	} else {
		conn.PacketsReceived++
		conn.BytesReceived += uint64(payloadLen)
	}

	// Handle RST - immediate close
	if flags&FlagRST != 0 {
		conn.State = StateClosed
		conn.EndTime = time.Now()
		t.emitEvent(Event{
			Type:       "closed",
			Connection: conn,
			OldState:   oldState,
			Timestamp:  time.Now(),
		})
		t.removeConnection(key)
		return conn
	}

	// State machine transitions
	switch conn.State {
	case StateClosed:
		if flags&FlagSYN != 0 && flags&FlagACK == 0 {
			// SYN only - connection initiation
			conn.State = StateSynSent
			if isNew {
				t.emitEvent(Event{
					Type:       "opened",
					Connection: conn,
					Timestamp:  time.Now(),
				})
			}
		}

	case StateSynSent:
		if flags&FlagSYN != 0 && flags&FlagACK != 0 {
			// SYN+ACK - server responding
			conn.State = StateSynReceived
		}

	case StateSynReceived:
		if flags&FlagACK != 0 && flags&FlagSYN == 0 {
			// ACK only - handshake complete
			conn.State = StateEstablished
			t.emitEvent(Event{
				Type:       "state_change",
				Connection: conn,
				OldState:   oldState,
				Timestamp:  time.Now(),
			})
		}

	case StateEstablished:
		if flags&FlagFIN != 0 {
			// FIN received - start closing
			conn.State = StateFinWait1
		}

	case StateFinWait1:
		if flags&FlagACK != 0 && flags&FlagFIN == 0 {
			conn.State = StateFinWait2
		} else if flags&FlagFIN != 0 {
			conn.State = StateLastAck
		}

	case StateFinWait2:
		if flags&FlagFIN != 0 {
			conn.State = StateTimeWait
			conn.EndTime = time.Now()
			t.emitEvent(Event{
				Type:       "closed",
				Connection: conn,
				OldState:   oldState,
				Timestamp:  time.Now(),
			})
			// In real implementation, would wait for TIME_WAIT timeout
			t.removeConnection(key)
		}

	case StateLastAck:
		if flags&FlagACK != 0 {
			conn.State = StateClosed
			conn.EndTime = time.Now()
			t.emitEvent(Event{
				Type:       "closed",
				Connection: conn,
				OldState:   oldState,
				Timestamp:  time.Now(),
			})
			t.removeConnection(key)
		}
	}

	return conn
}

// NormalizeKey returns a normalized key where the "lower" endpoint comes first.
// This ensures both directions of a connection map to the same key.
func NormalizeKey(srcIP string, srcPort uint16, dstIP string, dstPort uint16, protocol string) ConnKey {
	// Compare endpoints: first by IP, then by port
	srcFirst := srcIP < dstIP || (srcIP == dstIP && srcPort < dstPort)

	if srcFirst {
		return ConnKey{
			SrcIP:    srcIP,
			SrcPort:  srcPort,
			DstIP:    dstIP,
			DstPort:  dstPort,
			Protocol: protocol,
		}
	}
	return ConnKey{
		SrcIP:    dstIP,
		SrcPort:  dstPort,
		DstIP:    srcIP,
		DstPort:  srcPort,
		Protocol: protocol,
	}
}
