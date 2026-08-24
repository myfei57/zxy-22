package verifycase

import (
	"os"
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

func TestStationCursorAfterReadingDurable(t *testing.T) {
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
	if err := quotas.SetCapacity(stationID, 10); err != nil {
		t.Fatal(err)
	}
	sampler := station.NewSampler(stationID, "nh3", quotas, readings, recorder, rules, alerts, exceeds, centre, cursor)
	if err := os.RemoveAll(filepath.Join(dir, "readings")); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 8, 22, 7, 0, 0, 0, time.UTC)
	sampler.SetProvider(func() float64 { return 2.0 })
	if err := sampler.Sample(at); err == nil {
		t.Fatal("sample must fail when the reading value is not durable")
	}
	if !cursor.Value().IsZero() {
		t.Fatalf("sampling cursor must not advance when the reading is not durable, got %v", cursor.Value())
	}
}
