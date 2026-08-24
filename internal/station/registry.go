// Package station: persisted station registry.
package station

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

// Registry persists stations as one JSON file per station.
type Registry struct {
	dir      string
	mu       sync.RWMutex
	stations map[string]Station
}

// NewRegistry opens the station registry rooted at dir.
func NewRegistry(dir string) (*Registry, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("station: create registry %s: %w", dir, err)
	}
	registry := &Registry{dir: dir, stations: map[string]Station{}}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("station: read registry: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var station Station
		if err := json.Unmarshal(data, &station); err != nil {
			return nil, fmt.Errorf("station: decode %s: %w", entry.Name(), err)
		}
		registry.stations[station.ID] = station
	}
	return registry, nil
}

// Register stores a station record.
func (r *Registry) Register(station Station) error {
	if err := station.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(station)
	if err != nil {
		return fmt.Errorf("station: encode station: %w", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := state.WriteFile(filepath.Join(r.dir, station.ID+".json"), data); err != nil {
		return fmt.Errorf("station: persist %s: %w", station.ID, err)
	}
	r.stations[station.ID] = station
	return nil
}

// Get returns the station with the given identifier.
func (r *Registry) Get(id string) (Station, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	station, ok := r.stations[id]
	if !ok {
		return Station{}, fmt.Errorf("station: %s not found", id)
	}
	return station, nil
}

// List returns all stations ordered by name.
func (r *Registry) List() []Station {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stations := make([]Station, 0, len(r.stations))
	for _, station := range r.stations {
		stations = append(stations, station)
	}
	sort.Slice(stations, func(i, j int) bool {
		return stations[i].Name < stations[j].Name
	})
	return stations
}

// Count returns the number of registered stations.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.stations)
}
