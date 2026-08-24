// Package quota tracks the sampling capacity of every station.
package quota

import (
	"fmt"
)

// Quota is the capacity ledger of one station.
type Quota struct {
	StationID string `json:"station_id"`
	Capacity  int    `json:"capacity"`
	Used      int    `json:"used"`
}

// Validate checks the quota fields.
func (q Quota) Validate() error {
	if q.StationID == "" {
		return fmt.Errorf("quota: station id is required")
	}
	if q.Capacity < 0 {
		return fmt.Errorf("quota: capacity must not be negative")
	}
	if q.Used < 0 {
		return fmt.Errorf("quota: used must not be negative")
	}
	return nil
}
