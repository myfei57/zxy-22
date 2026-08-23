// Package alert: retry reporting for alerts the centre never received.
package alert

import (
	"envmonitor/internal/report"
)

// Retry reports every raised alert that has not been acknowledged yet. Alerts
// that were already acknowledged are never reported again.
func Retry(store *Store, centre *report.Center) (int, error) {
	count := 0
	for _, alert := range store.List() {
		if err := centre.Enqueue(report.CenterEvent{
			Type:      report.EventAlertReport,
			Key:       alert.ID,
			Detail:    alert.StationID + "/" + alert.MetricID,
			Timestamp: alert.RaisedAt,
		}); err != nil {
			return count, err
		}
		count++
		if err := store.SetStatus(alert.ID, StatusReported); err != nil {
			return count, err
		}
	}
	return count, nil
}
