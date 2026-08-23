// Package rule: versioned threshold manager.
package rule

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"envmonitor/internal/state"
)

// ErrNotPublished is returned when no threshold has been published for a
// metric yet.
var ErrNotPublished = errors.New("rule: no threshold published")

// VersionManager keeps the published thresholds. The active map is the live
// view used for judgement; the session map is the snapshot captured when the
// manager was loaded and is exposed to the console for inspection.
type VersionManager struct {
	dir     string
	mu      sync.RWMutex
	active  map[string]Threshold
	session map[string]Threshold
}

// NewVersionManager loads every published threshold from dir.
func NewVersionManager(dir string) (*VersionManager, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("rule: create version dir %s: %w", dir, err)
	}
	manager := &VersionManager{
		dir:     dir,
		active:  map[string]Threshold{},
		session: map[string]Threshold{},
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("rule: read version dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := state.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var threshold Threshold
		if err := json.Unmarshal(data, &threshold); err != nil {
			return nil, fmt.Errorf("rule: decode threshold %s: %w", entry.Name(), err)
		}
		manager.active[threshold.MetricID] = threshold
		manager.session[threshold.MetricID] = threshold
	}
	return manager, nil
}

// Publish stores a new threshold version and makes it the active one.
func (m *VersionManager) Publish(metricID string, value float64) (Threshold, error) {
	next := Threshold{
		MetricID:    metricID,
		Value:       value,
		PublishedAt: time.Now().UTC(),
	}
	if err := next.Validate(); err != nil {
		return Threshold{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if current, ok := m.active[metricID]; ok {
		next.Version = current.Version + 1
	} else {
		next.Version = 1
	}
	data, err := json.Marshal(next)
	if err != nil {
		return Threshold{}, fmt.Errorf("rule: encode threshold: %w", err)
	}
	if err := state.WriteFile(m.versionFile(metricID), data); err != nil {
		return Threshold{}, fmt.Errorf("rule: persist threshold %s: %w", metricID, err)
	}
	m.active[metricID] = next
	return next, nil
}

// Current returns the live published threshold for a metric.
func (m *VersionManager) Current(metricID string) (Threshold, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	threshold, ok := m.active[metricID]
	if !ok {
		return Threshold{}, fmt.Errorf("%w for %s", ErrNotPublished, metricID)
	}
	return threshold, nil
}

// Session returns the snapshot threshold captured when the manager loaded.
func (m *VersionManager) Session(metricID string) (Threshold, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	threshold, ok := m.session[metricID]
	if !ok {
		return Threshold{}, fmt.Errorf("%w for %s", ErrNotPublished, metricID)
	}
	return threshold, nil
}

// List returns every active threshold ordered by metric id.
func (m *VersionManager) List() []Threshold {
	m.mu.RLock()
	defer m.mu.RUnlock()
	thresholds := make([]Threshold, 0, len(m.active))
	for _, threshold := range m.active {
		thresholds = append(thresholds, threshold)
	}
	sort.Slice(thresholds, func(i, j int) bool {
		return thresholds[i].MetricID < thresholds[j].MetricID
	})
	return thresholds
}

func (m *VersionManager) versionFile(metricID string) string {
	return filepath.Join(m.dir, metricID+".json")
}
