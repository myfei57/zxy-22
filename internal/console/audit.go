// Package console: audit handlers.
package console

import (
	"net/http"

	"envmonitor/internal/audit"
)

func (s *Server) handleListAudit(w http.ResponseWriter, r *http.Request) {
	events, err := s.deps.Audit.List()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"events": events,
		"stats":  audit.Stats(events),
	})
}
