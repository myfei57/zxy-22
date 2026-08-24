// Package reading: durable monthly summaries.
package reading

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"envmonitor/internal/state"
)

// Summary is the aggregated view of one station for one month.
type Summary struct {
	Month       string  `json:"month"`
	StationID   string  `json:"station_id"`
	SampleCount int     `json:"sample_count"`
	AvgValue    float64 `json:"avg_value"`
	MaxValue    float64 `json:"max_value"`
}

// SummaryStore persists summaries as one JSON file per month key.
type SummaryStore struct {
	dir string
	mu  sync.Mutex
}

// NewSummaryStore opens the summary store rooted at dir.
func NewSummaryStore(dir string) (*SummaryStore, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("reading: create summary store %s: %w", dir, err)
	}
	return &SummaryStore{dir: dir}, nil
}

// WriteSummary durably stores one monthly summary.
func WriteSummary(st *SummaryStore, summary Summary) error {
	data, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("reading: encode summary: %w", err)
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	path := filepath.Join(st.dir, summaryKey(summary.StationID, summary.Month)+".json")
	if err := state.WriteFile(path, data); err != nil {
		return fmt.Errorf("reading: persist summary: %w", err)
	}
	return nil
}

func summaryKey(stationID, month string) string {
	return stationID + "-" + month
}
