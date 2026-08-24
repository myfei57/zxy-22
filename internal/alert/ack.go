// Package alert: centre acknowledgement of an alert.
package alert

import "time"

// Ack records that the centre has acknowledged the alert.
func Ack(store *Store, id string) (Alert, error) {
	alert, err := store.Get(id)
	if err != nil {
		return Alert{}, err
	}
	alert.Status = StatusAcked
	alert.AckedAt = time.Now().UTC()
	if err := store.Save(alert); err != nil {
		return Alert{}, err
	}
	return alert, nil
}
