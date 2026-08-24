// Package report: monthly report builder.
package report

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"envmonitor/internal/reading"
)

// Builder assembles monthly reports from stored readings.
type Builder struct {
	readings  *reading.Store
	windows   *WindowManager
	summaries *reading.SummaryStore
	files     *FileWriter
	registry  *Registry
}

// NewBuilder wires the monthly build flow together.
func NewBuilder(readings *reading.Store, windows *WindowManager, summaries *reading.SummaryStore, files *FileWriter, registry *Registry) *Builder {
	return &Builder{
		readings:  readings,
		windows:   windows,
		summaries: summaries,
		files:     files,
		registry:  registry,
	}
}

// Build produces the monthly report for one station and month. The report
// file is durably written before the report record is marked complete.
func (b *Builder) Build(stationID, month string) (Report, error) {
	start, end, err := parseMonth(month)
	if err != nil {
		return Report{}, err
	}
	if err := b.windows.Open(stationID, month); err != nil {
		return Report{}, err
	}
	rows, err := reading.ByWindow(b.readings, stationID, start, end)
	if err != nil {
		return Report{}, err
	}
	summary := summarize(rows, stationID, month)
	if err := b.windows.Close(stationID, month, b.summaries, summary); err != nil {
		return Report{}, err
	}
	content := renderReport(rows, summary)
	path, err := b.files.Write(stationID, month, content)
	if err != nil {
		return Report{}, err
	}
	report := Report{
		ID:        uuid.NewString(),
		StationID: stationID,
		Month:     month,
		Status:    StatusComplete,
		FilePath:  path,
		BuiltAt:   time.Now().UTC(),
	}
	if err := b.registry.Save(report); err != nil {
		return Report{}, err
	}
	return report, nil
}

// parseMonth converts "2006-01" into the month's UTC window.
func parseMonth(month string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("report: invalid month %q", month)
	}
	end := start.AddDate(0, 1, 0)
	return start.UTC(), end.UTC(), nil
}

func summarize(rows []reading.Reading, stationID, month string) reading.Summary {
	summary := reading.Summary{
		Month:     month,
		StationID: stationID,
	}
	if len(rows) == 0 {
		return summary
	}
	var total float64
	for _, rd := range rows {
		total += rd.Value
		if rd.Value > summary.MaxValue {
			summary.MaxValue = rd.Value
		}
	}
	summary.SampleCount = len(rows)
	summary.AvgValue = total / float64(len(rows))
	return summary
}

func renderReport(rows []reading.Reading, summary reading.Summary) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("station,metric,value,timestamp\n")
	for _, rd := range rows {
		buffer.WriteString(rd.StationID + "," + rd.MetricID + "," + strconv.FormatFloat(rd.Value, 'f', 2, 64) + "," + rd.Timestamp.Format(time.RFC3339) + "\n")
	}
	buffer.WriteString("summary,samples,avg,max\n")
	buffer.WriteString(summary.Month + "," + strconv.Itoa(summary.SampleCount) + "," +
		strconv.FormatFloat(summary.AvgValue, 'f', 2, 64) + "," + strconv.FormatFloat(summary.MaxValue, 'f', 2, 64) + "\n")
	return buffer.Bytes()
}
