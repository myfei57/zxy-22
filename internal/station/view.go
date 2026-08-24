// Package station: console view of a station.
package station

import (
	"errors"
	"time"

	"envmonitor/internal/quota"
	"envmonitor/internal/reading"
)

// View is what the station page renders.
type View struct {
	Station        Station          `json:"station"`
	Reading        *reading.Reading `json:"reading"`
	CursorAt       time.Time        `json:"cursor_at"`
	CursorLoaded   bool             `json:"cursor_loaded"`
	QuotaRemaining int              `json:"quota_remaining"`
	ReadingCount   int              `json:"reading_count"`
}

// Viewer builds station views from the live stores.
type Viewer struct {
	registry  *Registry
	readings  *reading.Store
	quotas    *quota.State
	cursors   map[string]*Cursor
	snapshots map[string]reading.Reading
}

// NewViewer wires the view flow for the console.
func NewViewer(registry *Registry, readings *reading.Store, quotas *quota.State, cursors map[string]*Cursor) *Viewer {
	viewer := &Viewer{
		registry:  registry,
		readings:  readings,
		quotas:    quotas,
		cursors:   cursors,
		snapshots: map[string]reading.Reading{},
	}
	for _, registered := range registry.List() {
		if current, err := reading.Current(readings, registered.ID); err == nil {
			viewer.snapshots[registered.ID] = current
		}
	}
	return viewer
}

// View renders the live station view.
func (v *Viewer) View(stationID string) (View, error) {
	station, err := v.registry.Get(stationID)
	if err != nil {
		return View{}, err
	}
	snapshot, hasSnapshot := v.snapshots[stationID]
	current, currentErr := reading.Current(v.readings, stationID)
	if hasSnapshot {
		current = snapshot
		currentErr = nil
	}
	if currentErr != nil && !errors.Is(currentErr, reading.ErrNoReadings) {
		return View{}, currentErr
	}
	count, err := v.readings.Count(stationID)
	if err != nil {
		return View{}, err
	}
	var cursorAt time.Time
	cursorLoaded := false
	if cursor, ok := v.cursors[stationID]; ok {
		cursorAt = cursor.Value()
		cursorLoaded = cursor.Loaded()
	}
	var readingPtr *reading.Reading
	if currentErr == nil {
		readingPtr = &current
	}
	return View{
		Station:        station,
		Reading:        readingPtr,
		CursorAt:       cursorAt,
		CursorLoaded:   cursorLoaded,
		QuotaRemaining: v.quotas.Remaining(stationID),
		ReadingCount:   count,
	}, nil
}
