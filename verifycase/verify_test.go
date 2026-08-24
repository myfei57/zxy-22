package verifycase

import (
	"path/filepath"
	"testing"
	"time"

	"envmonitor/internal/quota"
	"envmonitor/internal/reading"
	"envmonitor/internal/station"
)

func TestStationViewUsesCurrentReading(t *testing.T) {
	dir := t.TempDir()
	stations, err := station.NewRegistry(filepath.Join(dir, "stations"))
	if err != nil {
		t.Fatal(err)
	}
	readings, err := reading.NewStore(filepath.Join(dir, "readings"))
	if err != nil {
		t.Fatal(err)
	}
	quotas, err := quota.NewState(filepath.Join(dir, "quota"))
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := station.NewCursor(filepath.Join(dir, "cursors", "st.cursor"))
	if err != nil {
		t.Fatal(err)
	}
	registered, err := station.NewStation("东港站", "basin-1", []string{"nh3"})
	if err != nil {
		t.Fatal(err)
	}
	if err := stations.Register(registered); err != nil {
		t.Fatal(err)
	}
	if err := quotas.SetCapacity(registered.ID, 10); err != nil {
		t.Fatal(err)
	}
	earlier := reading.NewReading(registered.ID, "nh3", 1.0, time.Date(2026, 8, 22, 6, 0, 0, 0, time.UTC))
	if err := reading.Append(readings, earlier); err != nil {
		t.Fatal(err)
	}
	viewer := station.NewViewer(stations, readings, quotas, map[string]*station.Cursor{registered.ID: cursor})
	latest := reading.NewReading(registered.ID, "nh3", 3.1, time.Date(2026, 8, 22, 7, 0, 0, 0, time.UTC))
	if err := reading.Append(readings, latest); err != nil {
		t.Fatal(err)
	}
	view, err := viewer.View(registered.ID)
	if err != nil {
		t.Fatal(err)
	}
	if view.Reading == nil {
		t.Fatal("station view must expose the current reading")
	}
	if view.Reading.Value != 3.1 || view.Reading.Timestamp != latest.Timestamp {
		t.Fatalf("station view must read the current reading 3.1, got %v", view.Reading)
	}
}
