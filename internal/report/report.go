// Package report owns the monthly report lifecycle and the centre outbox that
// forwards readings and alerts to the downstream data centre.
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"envmonitor/internal/state"
)

// Report status values.
const StatusComplete = "complete"

// Report is one monthly aggregation result.
type Report struct {
	ID        string    `json:"id"`
	StationID string    `json:"station_id"`
	Month     string    `json:"month"`
	Status    string    `json:"status"`
	FilePath  string    `json:"file_path"`
	BuiltAt   time.Time `json:"built_at"`
}

// Validate checks the mandatory report fields.
func (r Report) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("report: id is required")
	}
	if strings.TrimSpace(r.StationID) == "" {
		return fmt.Errorf("report: station id is required")
	}
	if strings.TrimSpace(r.Month) == "" {
		return fmt.Errorf("report: month is required")
	}
	return nil
}

// Registry persists reports as one JSON file per report.
type Registry struct {
	dir    string
	mu     sync.RWMutex
	reports map[string]Report
}

// NewRegistry opens the report registry rooted at dir.
func NewRegistry(dir string) (*Registry, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("report: create registry %s: %w", dir, err)
	}
	registry := &Registry{dir: dir, reports: map[string]Report{}}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("report: read registry: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var report Report
		if err := json.Unmarshal(data, &report); err != nil {
			return nil, fmt.Errorf("report: decode %s: %w", entry.Name(), err)
		}
		registry.reports[report.ID] = report
	}
	return registry, nil
}

// Save durably stores a report record.
func (r *Registry) Save(report Report) error {
	if err := report.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("report: encode report: %w", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := state.WriteFile(filepath.Join(r.dir, report.ID+".json"), data); err != nil {
		return fmt.Errorf("report: persist %s: %w", report.ID, err)
	}
	r.reports[report.ID] = report
	return nil
}

// List returns all reports ordered by build time.
func (r *Registry) List() []Report {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reports := make([]Report, 0, len(r.reports))
	for _, report := range r.reports {
		reports = append(reports, report)
	}
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].BuiltAt.Before(reports[j].BuiltAt)
	})
	return reports
}
