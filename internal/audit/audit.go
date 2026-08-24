// Package audit records sampling and alert events for traceability.
package audit

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Audit event types.
const (
	TypeSampleSuccess     = "sample_success"
	TypeSampleFailed      = "sample_failed"
	TypeAlertRaised       = "alert_raised"
	TypeAlertAcked        = "alert_acked"
	TypeReportBuilt       = "report_built"
	TypeQuotaRejected     = "quota_rejected"
	TypeSegmentSent       = "segment_sent"
	TypeThresholdPublish  = "threshold_published"
	TypeStationRegistered = "station_registered"
)

// Event is one immutable audit record.
type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	StationID string    `json:"station_id,omitempty"`
	Detail    string    `json:"detail"`
	Timestamp time.Time `json:"timestamp"`
}

// NewEvent builds an audit event with a fresh identifier.
func NewEvent(eventType, stationID, detail string, timestamp time.Time) Event {
	return Event{
		ID:        uuid.NewString(),
		Type:      eventType,
		StationID: stationID,
		Detail:    strings.TrimSpace(detail),
		Timestamp: timestamp.UTC(),
	}
}
