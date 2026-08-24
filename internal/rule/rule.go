// Package rule owns the threshold rules and their versioned lifecycle.
package rule

import (
	"fmt"
	"strings"
	"time"
)

// Threshold is one published value for a metric, carrying a monotonic version.
type Threshold struct {
	MetricID    string    `json:"metric_id"`
	Version     int       `json:"version"`
	Value       float64   `json:"value"`
	PublishedAt time.Time `json:"published_at"`
}

// Validate checks that the threshold targets a known metric with a sane value.
func (t Threshold) Validate() error {
	if strings.TrimSpace(t.MetricID) == "" {
		return fmt.Errorf("rule: metric id is required")
	}
	if t.Value < 0 {
		return fmt.Errorf("rule: threshold value must not be negative")
	}
	return nil
}
