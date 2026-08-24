// Package quota: capacity gate evaluated before a reading is stored.
package quota

import (
	"errors"
	"fmt"
)

// ErrQuotaExceeded is returned when a station has no remaining capacity.
var ErrQuotaExceeded = errors.New("quota: station capacity exhausted")

// Check verifies that the station still has capacity before storage.
func Check(state *State, stationID string) error {
	state.mu.RLock()
	defer state.mu.RUnlock()
	quota, ok := state.quotas[stationID]
	if !ok {
		return fmt.Errorf("quota: unknown station %s", stationID)
	}
	if quota.Used >= quota.Capacity {
		return ErrQuotaExceeded
	}
	return nil
}
