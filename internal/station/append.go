// Package station: sample append orchestration with quota gate and audit.
package station

import (
	"fmt"

	"envmonitor/internal/audit"
	"envmonitor/internal/quota"
	"envmonitor/internal/reading"
)

// AppendSample applies the station capacity gate before durably storing the
// reading, then records the audit success event only after the reading is
// durable. Checking capacity first guarantees that a full station rejects the
// sample before any reading value is written, so exhausted capacity never lets
// new readings occupy storage.
func AppendSample(quotas *quota.State, readings *reading.Store, recorder *audit.Recorder, stationID string, rd reading.Reading) error {
	if err := quota.Check(quotas, stationID); err != nil {
		return err
	}
	if err := reading.Append(readings, rd); err != nil {
		return err
	}
	event := audit.NewEvent(
		audit.TypeSampleSuccess,
		stationID,
		fmt.Sprintf("%s=%.2f", rd.MetricID, rd.Value),
		rd.Timestamp,
	)
	if err := recorder.Record(event); err != nil {
		return fmt.Errorf("station: audit sample success: %w", err)
	}
	return nil
}
