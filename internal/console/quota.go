// Package console: quota handlers.
package console

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListQuota(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.deps.Quota.List())
}

func (s *Server) handleSetQuota(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "station")
	capacity, err := strconv.Atoi(r.URL.Query().Get("capacity"))
	if err != nil || capacity <= 0 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("capacity must be a positive integer"))
		return
	}
	if err := s.deps.Quota.SetCapacity(stationID, capacity); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"station":  stationID,
		"capacity": capacity,
		"used":     s.deps.Quota.Usage(stationID),
	})
}
