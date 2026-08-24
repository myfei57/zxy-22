// Package console: reading and exceed handlers.
package console

import (
	"fmt"
	"net/http"
	"time"

	"envmonitor/internal/reading"
)

func (s *Server) handleListReadings(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station")
	if stationID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("station query parameter is required"))
		return
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" && to == "" {
		rows, err := s.deps.Readings.List(stationID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, rows)
		return
	}
	var start, end time.Time
	var err error
	if from != "" {
		start, err = time.Parse(time.RFC3339, from)
		if err != nil {
			writeErr(w, http.StatusBadRequest, fmt.Errorf("from must be RFC3339"))
			return
		}
	}
	if to != "" {
		end, err = time.Parse(time.RFC3339, to)
		if err != nil {
			writeErr(w, http.StatusBadRequest, fmt.Errorf("to must be RFC3339"))
			return
		}
	}
	rows, err := reading.ByWindow(s.deps.Readings, stationID, start, end)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) handleListExceeds(w http.ResponseWriter, r *http.Request) {
	records, err := reading.ListExceeds(s.deps.Exceeds)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, records)
}
