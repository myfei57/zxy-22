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
	Station        Station           `json:"station"`
	Reading        *reading.Reading `json:"reading"`
	CursorAt       time.Time         `json:"cursor_at"`
	CursorLoaded   bool              `json:"cursor_loaded"`
	QuotaRemaining int               `json:"quota_remaining"`
	ReadingCount   int               `json:"reading_count"`
}

// Viewer builds station views from the live stores.
type Viewer struct {
	registry *Registry
	readings *reading.Store
	quotas   *quota.State
	cursors  map[string]*Cursor
}

// NewViewer wires the view flow for the console.
func NewViewer(registry *Registry, readings *reading.Store, quotas *quota.State, cursors map[string]*Cursor) *Viewer {
	return &Viewer{
		registry: registry,
		readings: readings,
		quotas:   quotas,
		cursors:  cursors,
	}
}

// View renders the live station view.
func (v *Viewer) View(stationID string) (View, error) {
	station, err := v.registry.Get(stationID)
	if err != nil {
		return View{}, err
	}
	// Read the station's current reading straight from the store so the
	// page always reflects the most recently sampled, durable value. Do
	// not cache it: a cached snapshot would lag behind newly written
	// samples and mislead operators into thinking sampling stalled.
	current, currentErr := reading.Current(v.readings, stationID)
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
