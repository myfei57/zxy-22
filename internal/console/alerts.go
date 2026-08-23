// Package console: alert handlers.
package console

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"envmonitor/internal/alert"
	"envmonitor/internal/audit"
)

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.deps.Alerts.List())
}

func (s *Server) handleAckAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	alertRecord, err := alert.Ack(s.deps.Alerts, id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	_ = s.deps.Audit.Record(audit.NewEvent(
		audit.TypeAlertAcked,
		alertRecord.StationID,
		id,
		time.Now().UTC(),
	))
	writeJSON(w, http.StatusOK, alertRecord)
}

func (s *Server) handleRetryAlerts(w http.ResponseWriter, r *http.Request) {
	count, err := alert.Retry(s.deps.Alerts, s.deps.Centre)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"reported": count})
}
