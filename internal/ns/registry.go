// Package ns: persisted basin registry.
package ns

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

// Registry persists basins as one JSON file per basin.
type Registry struct {
	dir    string
	mu     sync.RWMutex
	basins map[string]Basin
}

// NewRegistry opens the registry rooted at dir, loading every basin file.
func NewRegistry(dir string) (*Registry, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, err
	}
	registry := &Registry{dir: dir, basins: map[string]Basin{}}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("ns: read registry %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var basin Basin
		if err := json.Unmarshal(data, &basin); err != nil {
			return nil, fmt.Errorf("ns: decode basin %s: %w", entry.Name(), err)
		}
		registry.basins[basin.ID] = basin
	}
	return registry, nil
}

// Register stores a new basin. Registering an existing ID overwrites it.
func (r *Registry) Register(basin Basin) error {
	if err := basin.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := json.Marshal(basin)
	if err != nil {
		return fmt.Errorf("ns: encode basin: %w", err)
	}
	path := filepath.Join(r.dir, basin.ID+".json")
	if err := state.WriteFile(path, data); err != nil {
		return fmt.Errorf("ns: persist basin %s: %w", basin.ID, err)
	}
	r.basins[basin.ID] = basin
	return nil
}

// Get returns the basin with the given ID.
func (r *Registry) Get(id string) (Basin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	basin, ok := r.basins[id]
	if !ok {
		return Basin{}, fmt.Errorf("ns: basin %s not found", id)
	}
	return basin, nil
}

// List returns all basins ordered by name.
func (r *Registry) List() []Basin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	basins := make([]Basin, 0, len(r.basins))
	for _, basin := range r.basins {
		basins = append(basins, basin)
	}
	sort.Slice(basins, func(i, j int) bool {
		return basins[i].Name < basins[j].Name
	})
	return basins
}

// Remove deletes the basin file and the in-memory entry.
func (r *Registry) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.basins[id]; !ok {
		return fmt.Errorf("ns: basin %s not found", id)
	}
	if err := os.Remove(filepath.Join(r.dir, id+".json")); err != nil {
		return fmt.Errorf("ns: remove basin %s: %w", id, err)
	}
	delete(r.basins, id)
	return nil
}

// Count returns the number of registered basins.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.basins)
}
