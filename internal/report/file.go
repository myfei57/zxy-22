// Package report: monthly report file writer.
package report

import (
	"fmt"
	"path/filepath"

	"envmonitor/internal/state"
)

// FileWriter writes monthly report files into a directory.
type FileWriter struct {
	dir string
}

// NewFileWriter opens the report file directory.
func NewFileWriter(dir string) (*FileWriter, error) {
	if err := state.EnsureDir(dir); err != nil {
		return nil, fmt.Errorf("report: create file dir %s: %w", dir, err)
	}
	return &FileWriter{dir: dir}, nil
}

// Write durably stores the report content and returns its path.
func (w *FileWriter) Write(stationID, month string, content []byte) (string, error) {
	name := stationID + "-" + month + ".txt"
	path := filepath.Join(w.dir, name)
	if err := state.WriteFile(path, content); err != nil {
		return "", fmt.Errorf("report: write file %s: %w", path, err)
	}
	return path, nil
}
