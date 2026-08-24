// Package alert: raise flow from a verdict to a durable alert.
package alert

import (
	"time"

	"envmonitor/internal/audit"
	"envmonitor/internal/reading"
	"envmonitor/internal/report"
	"envmonitor/internal/rule"
)

// Raise records the exceed evidence, persists the alert, writes the audit
// event and reports the alert to the centre outbox.
func Raise(store *Store, exceeds *reading.ExceedStore, readings *reading.Store, recorder *audit.Recorder, centre *report.Center, rd reading.Reading, verdict rule.Verdict) (Alert, error) {
	if err := Mark(readings, exceeds, rd, verdict.Threshold); err != nil {
		return Alert{}, err
	}
	alert, err := NewAlert(rd, verdict)
	if err != nil {
		return Alert{}, err
	}
	if err := store.Save(alert); err != nil {
		return Alert{}, err
	}
	if err := centre.Enqueue(report.CenterEvent{
		Type:      report.EventAlertReport,
		Key:       alert.ID,
		Detail:    alert.StationID + "/" + alert.MetricID,
		Timestamp: alert.RaisedAt,
	}); err != nil {
		return Alert{}, err
	}
	if err := recorder.Record(audit.NewEvent(
		audit.TypeAlertRaised,
		alert.StationID,
		alert.MetricID,
		time.Now().UTC(),
	)); err != nil {
		return Alert{}, err
	}
	return alert, nil
}
