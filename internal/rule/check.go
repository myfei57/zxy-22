// Package rule: batch threshold judgement.
package rule

import (
	"fmt"

	"envmonitor/internal/reading"
)

// Verdict is the judgement of one reading against the published threshold.
type Verdict struct {
	ReadingID string  `json:"reading_id"`
	MetricID  string  `json:"metric_id"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Exceed    bool    `json:"exceed"`
}

// Check evaluates every reading against the currently published threshold of
// its metric.
func Check(manager *VersionManager, readings []reading.Reading) ([]Verdict, error) {
	verdicts := make([]Verdict, 0, len(readings))
	for _, rd := range readings {
		threshold, err := manager.Current(rd.MetricID)
		if err != nil {
			return nil, fmt.Errorf("rule: check reading %s: %w", rd.ID, err)
		}
		verdicts = append(verdicts, Verdict{
			ReadingID: rd.ID,
			MetricID:  rd.MetricID,
			Value:     rd.Value,
			Threshold: threshold.Value,
			Exceed:    reading.Evaluate(rd, threshold.Value),
		})
	}
	return verdicts, nil
}
