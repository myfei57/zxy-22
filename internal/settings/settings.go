// Package settings carries the runtime configuration for the envmonitor
// service: where data lives and which address the HTTP console listens on.
package settings

import (
	"os"
	"strconv"
	"time"
)

// Settings is the resolved runtime configuration.
type Settings struct {
	// DataDir is the root directory that all stores persist under.
	DataDir string
	// Addr is the listen address of the HTTP console, for example
	// "127.0.0.1:18080".
	Addr string
	// StationInterval is the default sampling interval used by demo samplers.
	StationInterval time.Duration
}

// Default returns the settings used when no environment override is present.
func Default() Settings {
	return Settings{
		DataDir:         ".envmonitor-data",
		Addr:            "127.0.0.1:18080",
		StationInterval: 60 * time.Second,
	}
}

// FromEnv resolves settings from ENVMONITOR_DATA_DIR, ENVMONITOR_ADDR and
// ENVMONITOR_INTERVAL_SECONDS, falling back to Default for missing values.
func FromEnv() Settings {
	cfg := Default()
	if value := os.Getenv("ENVMONITOR_DATA_DIR"); value != "" {
		cfg.DataDir = value
	}
	if value := os.Getenv("ENVMONITOR_ADDR"); value != "" {
		cfg.Addr = value
	}
	if value := os.Getenv("ENVMONITOR_INTERVAL_SECONDS"); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
			cfg.StationInterval = time.Duration(seconds) * time.Second
		}
	}
	return cfg
}
