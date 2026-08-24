// Package station: durable sampling cursor.
package station

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"envmonitor/internal/reading"
	"envmonitor/internal/state"
)

// Cursor tracks the sampling position of one station. It only advances after
// the durable step of the owning flow has completed.
type Cursor struct {
	path      string
	mu        sync.Mutex
	value     time.Time
	loaded    bool
}

// NewCursor opens the cursor file for a station.
func NewCursor(path string) (*Cursor, error) {
	if err := state.EnsureDir(filepath.Dir(path)); err != nil {
		return nil, fmt.Errorf("station: create cursor dir: %w", err)
	}
	cursor := &Cursor{path: path}
	if state.FileExists(path) {
		data, err := state.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var stored time.Time
		if err := json.Unmarshal(data, &stored); err != nil {
			return nil, fmt.Errorf("station: decode cursor: %w", err)
		}
		cursor.value = stored
		cursor.loaded = true
	}
	return cursor, nil
}

// Advance moves the cursor forward to the given timestamp.
func (c *Cursor) Advance(timestamp time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saveLocked(timestamp)
}

// AdvanceAfterAck advances the cursor only after the centre acknowledgement
// is durable.
func (c *Cursor) AdvanceAfterAck(acks *reading.AckStore, record reading.AckRecord) error {
	if err := reading.Ack(acks, record); err != nil {
		return err
	}
	return c.Advance(record.Timestamp)
}

// Value returns the current cursor position.
func (c *Cursor) Value() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Loaded reports whether a cursor file existed on disk.
func (c *Cursor) Loaded() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loaded
}

func (c *Cursor) saveLocked(timestamp time.Time) error {
	data, err := json.Marshal(timestamp)
	if err != nil {
		return fmt.Errorf("station: encode cursor: %w", err)
	}
	if err := state.WriteFile(c.path, data); err != nil {
		return fmt.Errorf("station: persist cursor: %w", err)
	}
	c.value = timestamp
	c.loaded = true
	return nil
}
