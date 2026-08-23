// Package reading: windowed queries over stored readings.
package reading

import "time"

// ByWindow returns the readings of a station whose timestamps fall inside the
// half-open range [start, end).
func ByWindow(st *Store, stationID string, start, end time.Time) ([]Reading, error) {
	rows, err := st.List(stationID)
	if err != nil {
		return nil, err
	}
	var window []Reading
	for _, rd := range rows {
		if !rd.Timestamp.Before(start) && rd.Timestamp.Before(end) {
			window = append(window, rd)
		}
	}
	return window, nil
}
