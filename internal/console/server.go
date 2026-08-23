// Package console exposes the envmonitor HTTP API and the four embedded
// operator pages backed by chi.
package console

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"envmonitor/internal/alert"
	"envmonitor/internal/audit"
	"envmonitor/internal/ns"
	"envmonitor/internal/quota"
	"envmonitor/internal/reading"
	"envmonitor/internal/report"
	"envmonitor/internal/rule"
	"envmonitor/internal/station"
)

// Deps carries every store and service the console handlers need.
type Deps struct {
	Basins    *ns.Registry
	Stations  *station.Registry
	Readings  *reading.Store
	Exceeds   *reading.ExceedStore
	Acks      *reading.AckStore
	Rules     *rule.VersionManager
	Alerts    *alert.Store
	Reports   *report.Registry
	Windows   *report.WindowManager
	Centre    *report.Center
	Quota     *quota.State
	Audit     *audit.Recorder
	Builder   *report.Builder
	DataDir   string
}

// Server owns the router and the per-station runtime objects.
type Server struct {
	deps     Deps
	dataDir  string
	samplers map[string]*station.Sampler
	senders  map[string]*station.Sender
	cursors  map[string]*station.Cursor
	mux      *chi.Mux
	viewer   *station.Viewer
}

// NewServer builds the runtime for every registered station and the router.
func NewServer(deps Deps) (*Server, error) {
	server := &Server{
		deps:     deps,
		dataDir:  deps.DataDir,
		samplers: map[string]*station.Sampler{},
		senders:  map[string]*station.Sender{},
		cursors:  map[string]*station.Cursor{},
	}
	for _, registered := range deps.Stations.List() {
		if err := server.registerStationRuntime(registered); err != nil {
			return nil, fmt.Errorf("console: prepare station %s: %w", registered.ID, err)
		}
	}
	server.viewer = station.NewViewer(deps.Stations, deps.Readings, deps.Quota, server.cursors)
	server.mux = chi.NewRouter()
	server.routes()
	return server, nil
}

// Handler returns the HTTP handler of the console.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.Use(middleware.Recoverer)
	s.mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/stations", http.StatusFound)
	})
	s.mux.Get("/stations", s.page("stations"))
	s.mux.Get("/readings", s.page("readings"))
	s.mux.Get("/alerts", s.page("alerts"))
	s.mux.Get("/audit", s.page("audit"))

	s.mux.Get("/api/health", s.handleHealth)
	s.mux.Get("/api/basins", s.handleListBasins)
	s.mux.Post("/api/basins", s.handleRegisterBasin)
	s.mux.Delete("/api/basins/{id}", s.handleDeleteBasin)
	s.mux.Get("/api/stations", s.handleListStations)
	s.mux.Post("/api/stations", s.handleRegisterStation)
	s.mux.Route("/api/stations/{id}", func(r chi.Router) {
		r.Get("/", s.handleStationView)
		r.Post("/sample", s.handleSampleStation)
		r.Post("/send", s.handleSendStation)
	})
	s.mux.Get("/api/readings", s.handleListReadings)
	s.mux.Get("/api/exceeds", s.handleListExceeds)
	s.mux.Get("/api/alerts", s.handleListAlerts)
	s.mux.Post("/api/alerts/{id}/ack", s.handleAckAlert)
	s.mux.Post("/api/alerts/retry", s.handleRetryAlerts)
	s.mux.Get("/api/reports", s.handleListReports)
	s.mux.Get("/api/reports/windows", s.handleListWindows)
	s.mux.Post("/api/reports/build", s.handleBuildReport)
	s.mux.Get("/api/reports/window/{station}/{month}", s.handleWindowState)
	s.mux.Get("/api/rules", s.handleListRules)
	s.mux.Post("/api/rules/publish", s.handlePublishRule)
	s.mux.Get("/api/rules/snapshot/{metric}", s.handleRuleSnapshot)
	s.mux.Get("/api/quota", s.handleListQuota)
	s.mux.Post("/api/quota/{station}", s.handleSetQuota)
	s.mux.Get("/api/audit", s.handleListAudit)
	s.mux.Get("/api/centre", s.handleListCentre)
}

func (s *Server) registerStationRuntime(st station.Station) error {
	for _, metric := range st.Metrics {
		cursor, err := station.NewCursor(filepath.Join(s.dataDir, "cursors", st.ID+".cursor"))
		if err != nil {
			return err
		}
		s.cursors[st.ID] = cursor
		key := st.ID + "/" + metric
		if _, ok := s.samplers[key]; !ok {
			s.samplers[key] = station.NewSampler(
				st.ID,
				metric,
				s.deps.Quota,
				s.deps.Readings,
				s.deps.Audit,
				s.deps.Rules,
				s.deps.Alerts,
				s.deps.Exceeds,
				s.deps.Centre,
				cursor,
			)
		}
	}
	if _, ok := s.senders[st.ID]; !ok {
		cursor := s.cursors[st.ID]
		if cursor == nil {
			return fmt.Errorf("console: no cursor prepared for %s", st.ID)
		}
		s.senders[st.ID] = station.NewSender(st.ID, s.deps.Centre, s.deps.Acks, cursor)
	}
	return nil
}
