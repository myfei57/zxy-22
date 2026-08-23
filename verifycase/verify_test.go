package verifycase

import (
	"path/filepath"
	"testing"
	"time"

	"envmonitor/internal/alert"
	"envmonitor/internal/audit"
	"envmonitor/internal/reading"
	"envmonitor/internal/report"
	"envmonitor/internal/rule"
)

func TestExceedRetrySkipsAckedAlert(t *testing.T) {
	dir := t.TempDir()
	readings, err := reading.NewStore(filepath.Join(dir, "readings"))
	if err != nil {
		t.Fatal(err)
	}
	exceeds, err := reading.NewExceedStore(filepath.Join(dir, "exceeds"))
	if err != nil {
		t.Fatal(err)
	}
	alerts, err := alert.NewStore(filepath.Join(dir, "alerts"))
	if err != nil {
		t.Fatal(err)
	}
	recorder, err := audit.NewRecorder(filepath.Join(dir, "audit"))
	if err != nil {
		t.Fatal(err)
	}
	centre, err := report.NewCenter(filepath.Join(dir, "centre"))
	if err != nil {
		t.Fatal(err)
	}
	rd := reading.NewReading("st-01", "nh3", 2.0, time.Date(2026, 8, 22, 7, 0, 0, 0, time.UTC))
	if err := reading.Append(readings, rd); err != nil {
		t.Fatal(err)
	}
	verdict := rule.Verdict{ReadingID: rd.ID, MetricID: "nh3", Value: 2.0, Threshold: 1.5, Exceed: true}
	raised, err := alert.Raise(alerts, exceeds, readings, recorder, centre, rd, verdict)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := alert.Ack(alerts, raised.ID); err != nil {
		t.Fatal(err)
	}
	reported, err := alert.Retry(alerts, centre)
	if err != nil {
		t.Fatal(err)
	}
	if reported != 0 {
		t.Fatalf("retry must not re-report an acknowledged alert, reported %d", reported)
	}
	events, err := centre.List()
	if err != nil {
		t.Fatal(err)
	}
	duplicates := 0
	for _, event := range events {
		if event.Key == raised.ID {
			duplicates++
		}
	}
	if duplicates != 1 {
		t.Fatalf("acknowledged alert must be reported exactly once, got %d", duplicates)
	}
}
