// Package station: forwarding sampling segments to the data centre.
package station

import (
	"time"

	"envmonitor/internal/reading"
	"envmonitor/internal/report"
)

// Sender forwards sampling segments and only then advances the send cursor.
type Sender struct {
	stationID string
	centre    *report.Center
	acks      *reading.AckStore
	cursor    *Cursor
}

// NewSender wires the send flow for one station.
func NewSender(stationID string, centre *report.Center, acks *reading.AckStore, cursor *Cursor) *Sender {
	return &Sender{
		stationID: stationID,
		centre:    centre,
		acks:      acks,
		cursor:    cursor,
	}
}

// Send delivers a sampling segment to the centre outbox and records the
// acknowledgement before moving the cursor.
func (s *Sender) Send(segmentID string, at time.Time) error {
	if err := s.centre.Enqueue(report.CenterEvent{
		Type:      report.EventReadingSegment,
		Key:       s.stationID,
		Detail:    segmentID,
		Timestamp: at,
	}); err != nil {
		return err
	}
	return s.cursor.AdvanceAfterAck(s.acks, reading.AckRecord{
		ID:        segmentID + "-ack",
		SegmentID: segmentID,
		StationID: s.stationID,
		Timestamp: at,
	})
}
