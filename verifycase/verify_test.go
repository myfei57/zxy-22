package verifycase

import (
	"path/filepath"
	"testing"
	"time"

	"envmonitor/internal/reading"
	"envmonitor/internal/rule"
)

func TestThresholdUsesCurrentVersion(t *testing.T) {
	dir := t.TempDir()
	first, err := rule.NewVersionManager(filepath.Join(dir, "rules"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Publish("nh3", 1.5); err != nil {
		t.Fatal(err)
	}
	manager, err := rule.NewVersionManager(filepath.Join(dir, "rules"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Publish("nh3", 0.5); err != nil {
		t.Fatal(err)
	}
	rd := reading.NewReading("st-01", "nh3", 0.9, time.Date(2026, 8, 22, 7, 0, 0, 0, time.UTC))
	verdicts, err := rule.Check(manager, []reading.Reading{rd})
	if err != nil {
		t.Fatal(err)
	}
	if len(verdicts) != 1 {
		t.Fatalf("expected one verdict, got %d", len(verdicts))
	}
	if !verdicts[0].Exceed {
		t.Fatalf("reading 0.9 must exceed the current threshold 0.5, verdict used %v", verdicts[0].Threshold)
	}
}
