// Package reading owns the water-quality reading values and their durable
// file stores: sample values, exceed records, centre acknowledgements and
// monthly summaries.
package reading

import (
	"time"

	"github.com/google/uuid"
)

// Metric identifiers used by the monitoring platform.
const (
	MetricPH  = "ph"
	MetricDO  = "do"
	MetricNH3 = "nh3"
)

// Reading is one sampled value of a metric at a station.
type Reading struct {
	ID            string    `json:"id"`
	StationID     string    `json:"station_id"`
	MetricID      string    `json:"metric_id"`
	Value         float64   `json:"value"`
	Timestamp     time.Time `json:"timestamp"`
	OverThreshold bool      `json:"over_threshold"`
	CreatedAt     time.Time `json:"created_at"`
}

// NewReading creates a reading with a fresh identifier.
func NewReading(stationID, metricID string, value float64, timestamp time.Time) Reading {
	return Reading{
		ID:        uuid.NewString(),
		StationID: stationID,
		MetricID:  metricID,
		Value:     value,
		Timestamp: timestamp.UTC(),
		CreatedAt: time.Now().UTC(),
	}
}
