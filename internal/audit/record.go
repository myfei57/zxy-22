// Package audit: durable event recorder.
package audit

import (
	"encoding/json"
	"fmt"

	"envmonitor/internal/state"
)

// Recorder appends audit events to a line journal.
type Recorder struct {
	journal *state.Journal
}

// NewRecorder opens the audit journal rooted at dir.
func NewRecorder(dir string) (*Recorder, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("audit: create dir %s: %w", dir, err)
	}
	return &Recorder{journal: state.NewJournal(dir + "/audit.jsonl")}, nil
}

// Record durably appends one audit event.
func (r *Recorder) Record(event Event) error {
	if err := r.journal.Append(event); err != nil {
		return fmt.Errorf("audit: record %s: %w", event.Type, err)
	}
	return nil
}

// List returns every audit event in stored order.
func (r *Recorder) List() ([]Event, error) {
	records, err := r.journal.Read()
	if err != nil {
		return nil, err
	}
	events := make([]Event, 0, len(records))
	for _, record := range records {
		var event Event
		if err := json.Unmarshal(record, &event); err != nil {
			return nil, fmt.Errorf("audit: decode event: %w", err)
		}
		events = append(events, event)
	}
	return events, nil
}
