// Package report: aggregation window lifecycle.
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"envmonitor/internal/reading"
	"envmonitor/internal/state"
)

// Window state values.
const (
	WindowOpen   = "open"
	WindowClosed = "closed"
)

// Window tracks one station/month aggregation window.
type Window struct {
	StationID string    `json:"station_id"`
	Month     string    `json:"month"`
	State     string    `json:"state"`
	OpenedAt  time.Time `json:"opened_at"`
	ClosedAt  time.Time `json:"closed_at,omitempty"`
}

// WindowManager persists the aggregation windows.
type WindowManager struct {
	dir     string
	mu      sync.RWMutex
	windows map[string]Window
}

// NewWindowManager opens the window store rooted at dir.
func NewWindowManager(dir string) (*WindowManager, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("report: create window dir %s: %w", dir, err)
	}
	manager := &WindowManager{dir: dir, windows: map[string]Window{}}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("report: read window dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var window Window
		if err := json.Unmarshal(data, &window); err != nil {
			return nil, fmt.Errorf("report: decode window %s: %w", entry.Name(), err)
		}
		manager.windows[windowKey(window.StationID, window.Month)] = window
	}
	return manager, nil
}

// Open creates or reopens the aggregation window for a station and month.
func (m *WindowManager) Open(stationID, month string) error {
	key := windowKey(stationID, month)
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.windows[key]; ok && existing.State == WindowClosed {
		return fmt.Errorf("report: window %s is already closed", key)
	}
	window := Window{
		StationID: stationID,
		Month:     month,
		State:     WindowOpen,
		OpenedAt:  time.Now().UTC(),
	}
	if err := m.saveLocked(window); err != nil {
		return err
	}
	m.windows[key] = window
	return nil
}

// Close durably stores the monthly summary and then marks the window closed.
func (m *WindowManager) Close(stationID, month string, summaries *reading.SummaryStore, summary reading.Summary) error {
	key := windowKey(stationID, month)
	m.mu.Lock()
	defer m.mu.Unlock()
	window, ok := m.windows[key]
	if !ok {
		return fmt.Errorf("report: window %s was never opened", key)
	}
	if window.State == WindowClosed {
		return fmt.Errorf("report: window %s is already closed", key)
	}
	window.State = WindowClosed
	window.ClosedAt = time.Now().UTC()
	if err := m.saveLocked(window); err != nil {
		return err
	}
	m.windows[key] = window
	if err := reading.WriteSummary(summaries, summary); err != nil {
		return err
	}
	return nil
}

// State returns the current state of a station/month window.
func (m *WindowManager) State(stationID, month string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	window, ok := m.windows[windowKey(stationID, month)]
	if !ok {
		return "", fmt.Errorf("report: window %s/%s not found", stationID, month)
	}
	return window.State, nil
}

// List returns every window ordered by station and month.
func (m *WindowManager) List() []Window {
	m.mu.RLock()
	defer m.mu.RUnlock()
	windows := make([]Window, 0, len(m.windows))
	for _, window := range m.windows {
		windows = append(windows, window)
	}
	return windows
}

func (m *WindowManager) saveLocked(window Window) error {
	data, err := json.Marshal(window)
	if err != nil {
		return fmt.Errorf("report: encode window: %w", err)
	}
	path := filepath.Join(m.dir, windowKey(window.StationID, window.Month)+".json")
	if err := state.WriteFile(path, data); err != nil {
		return fmt.Errorf("report: persist window: %w", err)
	}
	return nil
}

func windowKey(stationID, month string) string {
	return stationID + "-" + month
}
