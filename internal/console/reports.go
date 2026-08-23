// Package console: report, window and centre handlers.
package console

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"envmonitor/internal/audit"
	"envmonitor/internal/report"
)

func (s *Server) handleListReports(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.deps.Reports.List())
}

func (s *Server) handleListWindows(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.deps.Windows.List())
}

func (s *Server) handleBuildReport(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station")
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().UTC().Format("2006-01")
	}
	if stationID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("station query parameter is required"))
		return
	}
	built, err := s.deps.Builder.Build(stationID, month)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	_ = s.deps.Audit.Record(audit.NewEvent(
		audit.TypeReportBuilt,
		stationID,
		month,
		time.Now().UTC(),
	))
	writeJSON(w, http.StatusCreated, built)
}

func (s *Server) handleWindowState(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "station")
	month := chi.URLParam(r, "month")
	state, err := s.deps.Windows.State(stationID, month)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"station": stationID, "month": month, "state": state})
}

func (s *Server) handleListCentre(w http.ResponseWriter, r *http.Request) {
	events, err := s.deps.Centre.List()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"count":  len(events),
		"types":  centreTypes(events),
	})
}

func centreTypes(events []report.CenterEvent) map[string]int {
	counts := map[string]int{}
	for _, event := range events {
		counts[string(event.Type)]++
	}
	return counts
}
