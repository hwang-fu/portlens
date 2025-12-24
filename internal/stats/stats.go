package stats

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

// StatsRecorder tracks packet capture statistics.
type StatsRecorder struct {
	mu        sync.Mutex
	startTime time.Time

	// Counters
	PacketsCaptured uint64
	BytesProcessed  uint64
}

// NewRecorder creates a new stats recorder.
func NewRecorder() *StatsRecorder {
	return &StatsRecorder{
		startTime: time.Now(),
	}
}
