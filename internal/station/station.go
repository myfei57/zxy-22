// Package station owns the monitoring stations: registration, sampling,
// cursor advancement, centre sending and console views.
package station

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Station status values.
const (
	StatusIdle = "idle"
)

// Station is one monitoring site bound to a basin and a metric set.
type Station struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BasinID   string    `json:"basin_id"`
	Metrics   []string  `json:"metrics"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// NewStation builds a station with a fresh identifier.
func NewStation(name, basinID string, metrics []string) (Station, error) {
	station := Station{
		ID:        uuid.NewString(),
		Name:      strings.TrimSpace(name),
		BasinID:   basinID,
		Metrics:   metrics,
		Status:    StatusIdle,
		CreatedAt: time.Now().UTC(),
	}
	if err := station.Validate(); err != nil {
		return Station{}, err
	}
	return station, nil
}

// Validate checks the station fields.
func (s Station) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("station: name is required")
	}
	if strings.TrimSpace(s.BasinID) == "" {
		return fmt.Errorf("station: basin id is required")
	}
	if len(s.Metrics) == 0 {
		return fmt.Errorf("station: at least one metric is required")
	}
	return nil
}
