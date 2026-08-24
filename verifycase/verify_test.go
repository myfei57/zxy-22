package verifycase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"envmonitor/internal/alert"
	"envmonitor/internal/reading"
)

func TestExceedFlagAfterRecordDurable(t *testing.T) {
	dir := t.TempDir()
	readings, err := reading.NewStore(filepath.Join(dir, "readings"))
	if err != nil {
		t.Fatal(err)
	}
	exceeds, err := reading.NewExceedStore(filepath.Join(dir, "exceeds"))
	if err != nil {
		t.Fatal(err)
	}
	rd := reading.NewReading("st-01", "nh3", 2.0, time.Date(2026, 8, 22, 7, 0, 0, 0, time.UTC))
	if err := reading.Append(readings, rd); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, "exceeds")); err != nil {
		t.Fatal(err)
	}
	if err := alert.Mark(readings, exceeds, rd, 1.5); err == nil {
		t.Fatal("mark must fail when the exceed record is not durable")
	}
	rows, err := readings.List("st-01")
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].OverThreshold {
		t.Fatal("reading must not be flagged when the exceed record is not durable")
	}
}
