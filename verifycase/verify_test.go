package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"envmonitor/internal/reading"
	"envmonitor/internal/report"
)

func TestReportWindowAfterSummaryDurable(t *testing.T) {
	dir := t.TempDir()
	windows, err := report.NewWindowManager(filepath.Join(dir, "windows"))
	if err != nil {
		t.Fatal(err)
	}
	summaries, err := reading.NewSummaryStore(filepath.Join(dir, "summaries"))
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.Open("st-01", "2026-08"); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, "summaries")); err != nil {
		t.Fatal(err)
	}
	summary := reading.Summary{Month: "2026-08", StationID: "st-01", SampleCount: 1, AvgValue: 1.0, MaxValue: 1.0}
	if err := windows.Close("st-01", "2026-08", summaries, summary); err == nil {
		t.Fatal("close must fail when the summary is not durable")
	}
	state, err := windows.State("st-01", "2026-08")
	if err != nil {
		t.Fatal(err)
	}
	if state != report.WindowOpen {
		t.Fatalf("window must stay open when the summary is not durable, got %s", state)
	}
}
