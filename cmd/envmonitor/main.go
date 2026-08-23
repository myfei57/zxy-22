// Command envmonitor runs the water-quality monitoring console.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"envmonitor/internal/alert"
	"envmonitor/internal/audit"
	"envmonitor/internal/console"
	"envmonitor/internal/ns"
	"envmonitor/internal/quota"
	"envmonitor/internal/reading"
	"envmonitor/internal/report"
	"envmonitor/internal/rule"
	"envmonitor/internal/settings"
	"envmonitor/internal/station"
)

func main() {
	cfg := settings.FromEnv()
	if err := run(cfg); err != nil {
		log.Fatalf("envmonitor: %v", err)
	}
}

func run(cfg settings.Settings) error {
	root := cfg.DataDir
	basins, err := ns.NewRegistry(filepath.Join(root, "ns"))
	if err != nil {
		return err
	}
	stations, err := station.NewRegistry(filepath.Join(root, "stations"))
	if err != nil {
		return err
	}
	readings, err := reading.NewStore(filepath.Join(root, "readings"))
	if err != nil {
		return err
	}
	exceeds, err := reading.NewExceedStore(filepath.Join(root, "exceeds"))
	if err != nil {
		return err
	}
	acks, err := reading.NewAckStore(filepath.Join(root, "acks"))
	if err != nil {
		return err
	}
	summaries, err := reading.NewSummaryStore(filepath.Join(root, "summaries"))
	if err != nil {
		return err
	}
	reportFiles, err := report.NewFileWriter(filepath.Join(root, "report-files"))
	if err != nil {
		return err
	}
	rules, err := rule.NewVersionManager(filepath.Join(root, "rules"))
	if err != nil {
		return err
	}
	alerts, err := alert.NewStore(filepath.Join(root, "alerts"))
	if err != nil {
		return err
	}
	reports, err := report.NewRegistry(filepath.Join(root, "reports"))
	if err != nil {
		return err
	}
	windows, err := report.NewWindowManager(filepath.Join(root, "windows"))
	if err != nil {
		return err
	}
	centre, err := report.NewCenter(filepath.Join(root, "centre"))
	if err != nil {
		return err
	}
	quotas, err := quota.NewState(filepath.Join(root, "quota"))
	if err != nil {
		return err
	}
	recorder, err := audit.NewRecorder(filepath.Join(root, "audit"))
	if err != nil {
		return err
	}

	builder := report.NewBuilder(readings, windows, summaries, reportFiles, reports)

	if err := seed(basins, stations, quotas, rules, recorder); err != nil {
		return err
	}

	server, err := console.NewServer(console.Deps{
		Basins:    basins,
		Stations:  stations,
		Readings:  readings,
		Exceeds:   exceeds,
		Acks:      acks,
		Rules:     rules,
		Alerts:    alerts,
		Reports:   reports,
		Windows:   windows,
		Centre:    centre,
		Quota:     quotas,
		Audit:     recorder,
		Builder:   builder,
		DataDir:   root,
	})
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("envmonitor listening on %s (data=%s)", cfg.Addr, root)
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}

func seed(
	basins *ns.Registry,
	stations *station.Registry,
	quotas *quota.State,
	rules *rule.VersionManager,
	recorder *audit.Recorder,
) error {
	if basins.Count() == 0 {
		east, err := ns.NewBasin("东港流域", "东港市")
		if err != nil {
			return err
		}
		if err := basins.Register(east); err != nil {
			return err
		}
	}
	basinList := basins.List()
	if stations.Count() == 0 {
		first, err := station.NewStation("东港站", basinList[0].ID, []string{reading.MetricPH, reading.MetricDO, reading.MetricNH3})
		if err != nil {
			return err
		}
		second, err := station.NewStation("西港站", basinList[0].ID, []string{reading.MetricPH, reading.MetricNH3})
		if err != nil {
			return err
		}
		for _, registered := range []station.Station{first, second} {
			if err := stations.Register(registered); err != nil {
				return err
			}
			if err := quotas.SetCapacity(registered.ID, 1000); err != nil {
				return err
			}
			_ = recorder.Record(audit.NewEvent(
				audit.TypeStationRegistered,
				registered.ID,
				registered.Name,
				time.Now().UTC(),
			))
		}
	}
	_, err := rules.Publish(reading.MetricPH, 7.5)
	if err != nil {
		return err
	}
	if _, err := rules.Publish(reading.MetricDO, 5.0); err != nil {
		return err
	}
	if _, err := rules.Publish(reading.MetricNH3, 1.5); err != nil {
		return err
	}
	return nil
}
