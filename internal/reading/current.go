// Package reading: current reading lookup.
package reading

import (
	"errors"
	"fmt"
)

// ErrNoReadings is returned when a station has no stored reading yet.
var ErrNoReadings = errors.New("reading: station has no readings")

// Current returns the most recent reading of a station. It fails when the
// station has no stored reading yet.
func Current(st *Store, stationID string) (Reading, error) {
	rows, err := st.List(stationID)
	if err != nil {
		return Reading{}, err
	}
	if len(rows) == 0 {
		return Reading{}, fmt.Errorf("%w: station %s", ErrNoReadings, stationID)
	}
	return rows[len(rows)-1], nil
}
