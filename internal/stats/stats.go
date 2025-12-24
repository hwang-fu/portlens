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

// RecordPacket records a captured packet.
func (s *StatsRecorder) RecordPacket(size int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PacketsCaptured++
	s.BytesProcessed += uint64(size)
}

// Snapshot returns current stats as a JSON-serializable struct.
func (s *StatsRecorder) Snapshot() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	elapsed := time.Since(s.startTime).Seconds()
	packetsPerSec := float64(0)
	bytesPerSec := float64(0)
	if elapsed > 0 {
		packetsPerSec = float64(s.PacketsCaptured) / elapsed
		bytesPerSec = float64(s.BytesProcessed) / elapsed
	}

	return map[string]any{
		"type":             "stats",
		"timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"elapsed_seconds":  elapsed,
		"packets_captured": s.PacketsCaptured,
		"bytes_processed":  s.BytesProcessed,
		"packets_per_sec":  packetsPerSec,
		"bytes_per_sec":    bytesPerSec,
	}
}

// WriteJSON writes the current stats as JSON to the given writer.
func (s *StatsRecorder) WriteJSON(w io.Writer) {
	json.NewEncoder(w).Encode(s.Snapshot())
}
