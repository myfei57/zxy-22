// Package alert: flagging a reading over threshold with durable evidence.
package alert

import (
	"time"

	"github.com/google/uuid"

	"envmonitor/internal/reading"
)

// Mark records the exceed evidence first and only then flags the reading.
func Mark(store *reading.Store, exceeds *reading.ExceedStore, rd reading.Reading, threshold float64) error {
	record := reading.ExceedRecord{
		ID:         uuid.NewString(),
		ReadingID:  rd.ID,
		StationID:  rd.StationID,
		MetricID:   rd.MetricID,
		Value:      rd.Value,
		Threshold:  threshold,
		RecordedAt: time.Now().UTC(),
	}
	if err := reading.RecordExceed(exceeds, record); err != nil {
		return err
	}
	return store.SetOverThreshold(rd.ID, true)
}
