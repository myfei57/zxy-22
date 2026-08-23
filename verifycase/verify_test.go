package verifycase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"envmonitor/internal/reading"
	"envmonitor/internal/report"
	"envmonitor/internal/station"
)

func TestSampleCursorAfterAckDurable(t *testing.T) {
	dir := t.TempDir()
	centre, err := report.NewCenter(filepath.Join(dir, "centre"))
	if err != nil {
		t.Fatal(err)
	}
	acks, err := reading.NewAckStore(filepath.Join(dir, "acks"))
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := station.NewCursor(filepath.Join(dir, "cursors", "st.cursor"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, "acks")); err != nil {
		t.Fatal(err)
	}
	sender := station.NewSender("st-01", centre, acks, cursor)
	at := time.Date(2026, 8, 22, 7, 0, 0, 0, time.UTC)
	if err := sender.Send("seg-1", at); err == nil {
		t.Fatal("send must fail when the acknowledgement is not durable")
	}
	if !cursor.Value().IsZero() {
		t.Fatalf("send cursor must not advance when the acknowledgement is not durable, got %v", cursor.Value())
	}
}
