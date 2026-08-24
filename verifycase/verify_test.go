package verifycase

import (
	"path/filepath"
	"testing"
	"time"

	"envmonitor/internal/alert"
	"envmonitor/internal/audit"
	"envmonitor/internal/quota"
	"envmonitor/internal/reading"
	"envmonitor/internal/report"
	"envmonitor/internal/rule"
	"envmonitor/internal/station"
)

func TestStationQuotaRejectsBeforeSample(t *testing.T) {
	dir := t.TempDir()
	readings, err := reading.NewStore(filepath.Join(dir, "readings"))
	if err != nil {
		t.Fatal(err)
	}
	quotas, err := quota.NewState(filepath.Join(dir, "quota"))
	if err != nil {
		t.Fatal(err)
	}
	recorder, err := audit.NewRecorder(filepath.Join(dir, "audit"))
	if err != nil {
		t.Fatal(err)
	}
	rules, err := rule.NewVersionManager(filepath.Join(dir, "rules"))
	if err != nil {
		t.Fatal(err)
	}
	alerts, err := alert.NewStore(filepath.Join(dir, "alerts"))
	if err != nil {
		t.Fatal(err)
	}
	exceeds, err := reading.NewExceedStore(filepath.Join(dir, "exceeds"))
	if err != nil {
		t.Fatal(err)
	}
	centre, err := report.NewCenter(filepath.Join(dir, "centre"))
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := station.NewCursor(filepath.Join(dir, "cursors", "st.cursor"))
	if err != nil {
		t.Fatal(err)
	}
	stationID := "st-01"
	if err := quotas.SetCapacity(stationID, 1); err != nil {
		t.Fatal(err)
	}
	sampler := station.NewSampler(stationID, "nh3", quotas, readings, recorder, rules, alerts, exceeds, centre, cursor)
	sampler.SetProvider(func() float64 { return 1.0 })
	first := time.Date(2026, 8, 22, 7, 0, 0, 0, time.UTC)
	if err := sampler.Sample(first); err != nil {
		t.Fatal(err)
	}
	second := time.Date(2026, 8, 22, 8, 0, 0, 0, time.UTC)
	if err := sampler.Sample(second); err == nil {
		t.Fatal("over-quota sample must be rejected")
	}
	count, err := readings.Count(stationID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("over-quota reading must not be stored, got %d", count)
	}
}
