// Package report: centre outbox for events forwarded downstream.
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"envmonitor/internal/state"
)

// EventType discriminates outbox entries.
type EventType string

// Outbox event types.
const (
	EventReadingSegment EventType = "reading_segment"
	EventAlertReport    EventType = "alert_report"
)

// CenterEvent is one outbound event waiting at the centre outbox.
type CenterEvent struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Key       string    `json:"key"`
	Detail    string    `json:"detail"`
	Timestamp time.Time `json:"timestamp"`
}

// Center is the durable outbox that the downstream data centre drains.
type Center struct {
	dir string
	mu  sync.Mutex
}

// NewCenter opens the outbox rooted at dir.
func NewCenter(dir string) (*Center, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("report: create centre dir %s: %w", dir, err)
	}
	return &Center{dir: dir}, nil
}

// Enqueue durably appends one event to the outbox.
func (c *Center) Enqueue(event CenterEvent) error {
	if strings.TrimSpace(event.ID) == "" {
		event.ID = fmt.Sprintf("%s-%d", event.Key, time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("report: encode centre event: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := state.WriteFile(filepath.Join(c.dir, event.ID+".json"), data); err != nil {
		return fmt.Errorf("report: persist centre event %s: %w", event.ID, err)
	}
	return nil
}

// List returns every event in the outbox ordered by timestamp.
func (c *Center) List() ([]CenterEvent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return nil, fmt.Errorf("report: read centre: %w", err)
	}
	var events []CenterEvent
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(c.dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var event CenterEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, fmt.Errorf("report: decode centre event %s: %w", entry.Name(), err)
		}
		events = append(events, event)
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})
	return events, nil
}
