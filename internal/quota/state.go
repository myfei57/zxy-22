// Package quota: persisted capacity ledger.
package quota

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"envmonitor/internal/state"
)

// State persists one quota file per station.
type State struct {
	dir    string
	mu     sync.RWMutex
	quotas map[string]Quota
}

// NewState opens the quota ledger rooted at dir.
func NewState(dir string) (*State, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("quota: create state %s: %w", dir, err)
	}
	ledger := &State{dir: dir, quotas: map[string]Quota{}}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("quota: read state: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var quota Quota
		if err := json.Unmarshal(data, &quota); err != nil {
			return nil, fmt.Errorf("quota: decode %s: %w", entry.Name(), err)
		}
		ledger.quotas[quota.StationID] = quota
	}
	return ledger, nil
}

// SetCapacity sets the capacity of a station.
func (s *State) SetCapacity(stationID string, capacity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.quotas[stationID]
	if !ok {
		current = Quota{StationID: stationID}
	}
	current.Capacity = capacity
	if err := current.Validate(); err != nil {
		return err
	}
	if err := s.saveLocked(current); err != nil {
		return err
	}
	s.quotas[stationID] = current
	return nil
}

// Consume charges one unit of capacity after a reading is stored.
func (s *State) Consume(stationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.quotas[stationID]
	if !ok {
		return fmt.Errorf("quota: unknown station %s", stationID)
	}
	if current.Used >= current.Capacity {
		return fmt.Errorf("quota: station %s already full", stationID)
	}
	current.Used++
	if err := s.saveLocked(current); err != nil {
		return err
	}
	s.quotas[stationID] = current
	return nil
}

// Remaining returns the unused capacity of a station.
func (s *State) Remaining(stationID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	quota, ok := s.quotas[stationID]
	if !ok {
		return 0
	}
	if quota.Used >= quota.Capacity {
		return 0
	}
	return quota.Capacity - quota.Used
}

// Usage returns the used capacity of a station.
func (s *State) Usage(stationID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	quota, ok := s.quotas[stationID]
	if !ok {
		return 0
	}
	return quota.Used
}

// List returns every quota entry ordered by station id.
func (s *State) List() []Quota {
	s.mu.RLock()
	defer s.mu.RUnlock()
	quotas := make([]Quota, 0, len(s.quotas))
	for _, quota := range s.quotas {
		quotas = append(quotas, quota)
	}
	return quotas
}

func (s *State) saveLocked(quota Quota) error {
	data, err := json.Marshal(quota)
	if err != nil {
		return fmt.Errorf("quota: encode quota: %w", err)
	}
	path := filepath.Join(s.dir, quota.StationID+".json")
	if err := state.WriteFile(path, data); err != nil {
		return fmt.Errorf("quota: persist %s: %w", quota.StationID, err)
	}
	return nil
}
