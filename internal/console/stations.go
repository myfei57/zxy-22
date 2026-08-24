// Package console: station and basin handlers.
package console

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"envmonitor/internal/audit"
	"envmonitor/internal/ns"
	"envmonitor/internal/quota"
	"envmonitor/internal/station"
)

func (s *Server) handleListBasins(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.deps.Basins.List())
}

type basinRequest struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

func (s *Server) handleRegisterBasin(w http.ResponseWriter, r *http.Request) {
	var request basinRequest
	if err := decodeBody(r, &request); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	basin, err := ns.NewBasin(request.Name, request.Region)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := s.deps.Basins.Register(basin); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, basin)
}

func (s *Server) handleDeleteBasin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.deps.Basins.Remove(id); err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"removed": id})
}

func (s *Server) handleListStations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.deps.Stations.List())
}

type stationRequest struct {
	Name     string   `json:"name"`
	BasinID  string   `json:"basin_id"`
	Metrics  []string `json:"metrics"`
	Capacity int      `json:"capacity"`
}

func (s *Server) handleRegisterStation(w http.ResponseWriter, r *http.Request) {
	var request stationRequest
	if err := decodeBody(r, &request); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if _, err := s.deps.Basins.Get(request.BasinID); err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("unknown basin %s", request.BasinID))
		return
	}
	registered, err := station.NewStation(request.Name, request.BasinID, request.Metrics)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := s.deps.Stations.Register(registered); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	capacity := request.Capacity
	if capacity <= 0 {
		capacity = 100
	}
	if err := s.deps.Quota.SetCapacity(registered.ID, capacity); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	_ = s.deps.Audit.Record(audit.NewEvent(
		audit.TypeStationRegistered,
		registered.ID,
		registered.Name,
		time.Now().UTC(),
	))
	if err := s.registerStationRuntime(registered); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, registered)
}

func (s *Server) handleStationView(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	view, err := s.viewer.View(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleSampleStation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "nh3"
	}
	value, err := strconv.ParseFloat(r.URL.Query().Get("value"), 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("value must be a number"))
		return
	}
	key := id + "/" + metric
	sampler, ok := s.samplers[key]
	if !ok {
		writeErr(w, http.StatusNotFound, fmt.Errorf("no sampler for %s", key))
		return
	}
	sampler.SetProvider(func() float64 { return value })
	at := time.Now().UTC()
	if err := sampler.Sample(at); err != nil {
		if errors.Is(err, quota.ErrQuotaExceeded) {
			_ = s.deps.Audit.Record(audit.NewEvent(audit.TypeQuotaRejected, id, key, at))
		}
		_ = s.deps.Audit.Record(audit.NewEvent(audit.TypeSampleFailed, id, key, at))
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sampled": true,
		"station": id,
		"metric":  metric,
		"at":      at,
	})
}

func (s *Server) handleSendStation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	segment := r.URL.Query().Get("segment")
	if segment == "" {
		segment = fmt.Sprintf("seg-%d", time.Now().Unix())
	}
	sender, ok := s.senders[id]
	if !ok {
		writeErr(w, http.StatusNotFound, fmt.Errorf("no sender for %s", id))
		return
	}
	at := time.Now().UTC()
	if err := sender.Send(segment, at); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	_ = s.deps.Audit.Record(audit.NewEvent(audit.TypeSegmentSent, id, segment, at))
	writeJSON(w, http.StatusOK, map[string]any{"sent": true, "segment": segment})
}
