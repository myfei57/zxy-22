// Package alert owns exceed handling: marking readings, alert records,
// reporting to the centre and retrying failed reports.
package alert

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"envmonitor/internal/reading"
	"envmonitor/internal/rule"
)

// Alert lifecycle states.
const (
	StatusRaised   = "raised"
	StatusReported = "reported"
	StatusAcked    = "acked"
)

// Alert is one exceed event tracked until the centre acknowledges it.
type Alert struct {
	ID        string    `json:"id"`
	ReadingID string    `json:"reading_id"`
	StationID string    `json:"station_id"`
	MetricID  string    `json:"metric_id"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Status    string    `json:"status"`
	RaisedAt  time.Time `json:"raised_at"`
	AckedAt   time.Time `json:"acked_at,omitempty"`
}

// NewAlert builds an alert from an exceed verdict.
func NewAlert(rd reading.Reading, verdict rule.Verdict) (Alert, error) {
	alert := Alert{
		ID:        uuid.NewString(),
		ReadingID: rd.ID,
		StationID: rd.StationID,
		MetricID:  rd.MetricID,
		Value:     rd.Value,
		Threshold: verdict.Threshold,
		Status:    StatusRaised,
		RaisedAt:  time.Now().UTC(),
	}
	if err := alert.Validate(); err != nil {
		return Alert{}, err
	}
	return alert, nil
}

// Validate checks the mandatory alert fields.
func (a Alert) Validate() error {
	if strings.TrimSpace(a.ID) == "" {
		return fmt.Errorf("alert: id is required")
	}
	if strings.TrimSpace(a.StationID) == "" {
		return fmt.Errorf("alert: station id is required")
	}
	if strings.TrimSpace(a.MetricID) == "" {
		return fmt.Errorf("alert: metric id is required")
	}
	return nil
}
