// Package alert: persisted alert record store.
package alert

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

// Store persists alerts as one JSON file per alert.
type Store struct {
	dir    string
	mu     sync.RWMutex
	alerts map[string]Alert
}

// NewStore opens the alert store rooted at dir.
func NewStore(dir string) (*Store, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("alert: create store %s: %w", dir, err)
	}
	store := &Store{dir: dir, alerts: map[string]Alert{}}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("alert: read store: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var alert Alert
		if err := json.Unmarshal(data, &alert); err != nil {
			return nil, fmt.Errorf("alert: decode %s: %w", entry.Name(), err)
		}
		store.alerts[alert.ID] = alert
	}
	return store, nil
}

// Save durably stores an alert record.
func (s *Store) Save(alert Alert) error {
	if err := alert.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("alert: encode alert: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := state.WriteFile(filepath.Join(s.dir, alert.ID+".json"), data); err != nil {
		return fmt.Errorf("alert: persist %s: %w", alert.ID, err)
	}
	s.alerts[alert.ID] = alert
	return nil
}

// Get returns the alert with the given identifier.
func (s *Store) Get(id string) (Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	alert, ok := s.alerts[id]
	if !ok {
		return Alert{}, fmt.Errorf("alert: %s not found", id)
	}
	return alert, nil
}

// List returns all alerts ordered by raised time.
func (s *Store) List() []Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	alerts := make([]Alert, 0, len(s.alerts))
	for _, alert := range s.alerts {
		alerts = append(alerts, alert)
	}
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].RaisedAt.Before(alerts[j].RaisedAt)
	})
	return alerts
}

// SetStatus durably updates the status of one alert.
func (s *Store) SetStatus(id, status string) error {
	s.mu.RLock()
	alert, ok := s.alerts[id]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("alert: %s not found", id)
	}
	alert.Status = status
	return s.Save(alert)
}
