// Package reading: durable centre acknowledgements.
package reading

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"envmonitor/internal/state"
)

// AckRecord is the persisted confirmation that the data centre received a
// sampling segment.
type AckRecord struct {
	ID        string    `json:"id"`
	SegmentID string    `json:"segment_id"`
	StationID string    `json:"station_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AckStore persists acknowledgements as one JSON file per record.
type AckStore struct {
	dir string
	mu  sync.Mutex
}

// NewAckStore opens the acknowledgement store rooted at dir.
func NewAckStore(dir string) (*AckStore, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("reading: create ack store %s: %w", dir, err)
	}
	return &AckStore{dir: dir}, nil
}

// Ack durably stores one acknowledgement record.
func Ack(st *AckStore, record AckRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("reading: encode ack record: %w", err)
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	if err := state.WriteFile(filepath.Join(st.dir, record.ID+".json"), data); err != nil {
		return fmt.Errorf("reading: persist ack %s: %w", record.ID, err)
	}
	return nil
}
