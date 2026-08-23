// Package reading: durable exceed records.
package reading

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"envmonitor/internal/state"
)

// ExceedRecord is the persisted evidence that a reading crossed its threshold.
type ExceedRecord struct {
	ID         string    `json:"id"`
	ReadingID  string    `json:"reading_id"`
	StationID  string    `json:"station_id"`
	MetricID   string    `json:"metric_id"`
	Value      float64   `json:"value"`
	Threshold  float64   `json:"threshold"`
	RecordedAt time.Time `json:"recorded_at"`
}

// ExceedStore persists exceed records as one JSON file per record.
type ExceedStore struct {
	dir string
	mu  sync.Mutex
}

// NewExceedStore opens the exceed record store rooted at dir.
func NewExceedStore(dir string) (*ExceedStore, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("reading: create exceed store %s: %w", dir, err)
	}
	return &ExceedStore{dir: dir}, nil
}

// RecordExceed durably stores one exceed record.
func RecordExceed(st *ExceedStore, record ExceedRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("reading: encode exceed record: %w", err)
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	if err := state.WriteFile(filepath.Join(st.dir, record.ID+".json"), data); err != nil {
		return fmt.Errorf("reading: persist exceed record %s: %w", record.ID, err)
	}
	return nil
}

// ListExceeds returns every exceed record in the store.
func ListExceeds(st *ExceedStore) ([]ExceedRecord, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	entries, err := os.ReadDir(st.dir)
	if err != nil {
		return nil, fmt.Errorf("reading: list exceed store: %w", err)
	}
	var records []ExceedRecord
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(st.dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var record ExceedRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("reading: decode exceed record %s: %w", entry.Name(), err)
		}
		records = append(records, record)
	}
	return records, nil
}
