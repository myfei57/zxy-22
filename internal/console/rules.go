// Package console: threshold rule handlers.
package console

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"envmonitor/internal/audit"
	"envmonitor/internal/rule"
)

func (s *Server) handleListRules(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.deps.Rules.List())
}

func (s *Server) handlePublishRule(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	value, err := strconv.ParseFloat(r.URL.Query().Get("value"), 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("value must be a number"))
		return
	}
	published, err := s.deps.Rules.Publish(metric, value)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	_ = s.deps.Audit.Record(audit.NewEvent(
		audit.TypeThresholdPublish,
		"",
		fmt.Sprintf("%s@v%d=%.2f", metric, published.Version, published.Value),
		time.Now().UTC(),
	))
	writeJSON(w, http.StatusCreated, published)
}

func (s *Server) handleRuleSnapshot(w http.ResponseWriter, r *http.Request) {
	metric := chi.URLParam(r, "metric")
	snapshot, err := rule.Snapshot(s.deps.Rules, metric)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}
