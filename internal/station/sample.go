// Package station: sampling flow.
package station

import (
	"errors"
	"sync"
	"time"

	"envmonitor/internal/alert"
	"envmonitor/internal/audit"
	"envmonitor/internal/quota"
	"envmonitor/internal/reading"
	"envmonitor/internal/report"
	"envmonitor/internal/rule"
)

// Sampler drives one station/metric sampling loop.
type Sampler struct {
	stationID string
	metricID  string

	mu      sync.Mutex
	provide func() float64

	quotas   *quota.State
	readings *reading.Store
	recorder *audit.Recorder
	rules    *rule.VersionManager
	alerts   *alert.Store
	exceeds  *reading.ExceedStore
	centre   *report.Center
	cursor   *Cursor
}

// NewSampler wires one station/metric sampling loop.
func NewSampler(
	stationID, metricID string,
	quotas *quota.State,
	readings *reading.Store,
	recorder *audit.Recorder,
	rules *rule.VersionManager,
	alerts *alert.Store,
	exceeds *reading.ExceedStore,
	centre *report.Center,
	cursor *Cursor,
) *Sampler {
	return &Sampler{
		stationID: stationID,
		metricID:  metricID,
		provide:   func() float64 { return 0.8 },
		quotas:    quotas,
		readings:  readings,
		recorder:  recorder,
		rules:     rules,
		alerts:    alerts,
		exceeds:   exceeds,
		centre:    centre,
		cursor:    cursor,
	}
}

// SetProvider replaces the value source used by Sample.
func (s *Sampler) SetProvider(provider func() float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if provider != nil {
		s.provide = provider
	}
}

// Sample durably stores one reading and advances the sampling cursor only
// after the reading and its downstream effects are durable. AppendSample
// persists the reading value and records the sample-success audit, Consume
// charges the quota, and evaluate checks thresholds and raises any alerts;
// the cursor advances last. A failure at any earlier step leaves the cursor
// in place, so on restart the gap is visible (cursor behind, reading absent)
// rather than silently skipped because the cursor already passed the slot.
func (s *Sampler) Sample(at time.Time) error {
	rd := reading.NewReading(s.stationID, s.metricID, s.currentValue(), at)
	if err := AppendSample(s.quotas, s.readings, s.recorder, s.stationID, rd); err != nil {
		return err
	}
	if err := s.quotas.Consume(s.stationID); err != nil {
		return err
	}
	if err := s.evaluate(rd); err != nil {
		return err
	}
	return s.cursor.Advance(rd.Timestamp)
}

func (s *Sampler) evaluate(rd reading.Reading) error {
	verdicts, err := rule.Check(s.rules, []reading.Reading{rd})
	if err != nil {
		if errors.Is(err, rule.ErrNotPublished) {
			return nil
		}
		return err
	}
	for _, verdict := range verdicts {
		if !verdict.Exceed {
			continue
		}
		if _, err := alert.Raise(s.alerts, s.exceeds, s.readings, s.recorder, s.centre, rd, verdict); err != nil {
			return err
		}
	}
	return nil
}

func (s *Sampler) currentValue() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.provide()
}
