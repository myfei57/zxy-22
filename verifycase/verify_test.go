package verifycase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"envmonitor/internal/reading"
	"envmonitor/internal/report"
)

func TestMonthlyReportAfterFileDurable(t *testing.T) {
	dir := t.TempDir()
	readings, err := reading.NewStore(filepath.Join(dir, "readings"))
	if err != nil {
		t.Fatal(err)
	}
	windows, err := report.NewWindowManager(filepath.Join(dir, "windows"))
	if err != nil {
		t.Fatal(err)
	}
	summaries, err := reading.NewSummaryStore(filepath.Join(dir, "summaries"))
	if err != nil {
		t.Fatal(err)
	}
	files, err := report.NewFileWriter(filepath.Join(dir, "report-files"))
	if err != nil {
		t.Fatal(err)
	}
	registry, err := report.NewRegistry(filepath.Join(dir, "reports"))
	if err != nil {
		t.Fatal(err)
	}
	rd := reading.NewReading("st-01", "nh3", 1.0, time.Date(2026, 8, 5, 8, 0, 0, 0, time.UTC))
	if err := reading.Append(readings, rd); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, "report-files")); err != nil {
		t.Fatal(err)
	}
	builder := report.NewBuilder(readings, windows, summaries, files, registry)
	if _, err := builder.Build("st-01", "2026-08"); err == nil {
		t.Fatal("build must fail when the report file is not durable")
	}
	for _, rep := range registry.List() {
		if rep.Status == report.StatusComplete {
			t.Fatal("report must not be marked complete when the file is not durable")
		}
	}
}
