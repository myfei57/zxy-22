// Package reading: durable value store for sample readings.
package reading

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"envmonitor/internal/state"
)

// Store persists readings as one JSON file per reading.
type Store struct {
	dir string
	mu  sync.RWMutex
}

// NewStore opens the reading store rooted at dir.
func NewStore(dir string) (*Store, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("reading: create store %s: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

func (s *Store) path(id string) string {
	return filepath.Join(s.dir, id+".json")
}

// List returns all readings of a station ordered by timestamp.
func (s *Store) List(stationID string) ([]Reading, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("reading: list store: %w", err)
	}
	var rows []Reading
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var rd Reading
		if err := json.Unmarshal(data, &rd); err != nil {
			return nil, fmt.Errorf("reading: decode %s: %w", entry.Name(), err)
		}
		if rd.StationID == stationID {
			rows = append(rows, rd)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Timestamp.Before(rows[j].Timestamp)
	})
	return rows, nil
}

// Count returns how many readings a station has stored.
func (s *Store) Count(stationID string) (int, error) {
	rows, err := s.List(stationID)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// SetOverThreshold durably updates the over-threshold flag of one reading.
func (s *Store) SetOverThreshold(id string, flag bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := state.ReadFile(s.path(id))
	if err != nil {
		return fmt.Errorf("reading: update flag %s: %w", id, err)
	}
	var rd Reading
	if err := json.Unmarshal(data, &rd); err != nil {
		return fmt.Errorf("reading: decode flag target %s: %w", id, err)
	}
	rd.OverThreshold = flag
	encoded, err := json.Marshal(rd)
	if err != nil {
		return fmt.Errorf("reading: encode flag target %s: %w", id, err)
	}
	if err := state.WriteFile(s.path(id), encoded); err != nil {
		return fmt.Errorf("reading: persist flag %s: %w", id, err)
	}
	return nil
}
