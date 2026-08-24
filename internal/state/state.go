// Package state provides the durable file primitives used by every
// envmonitor store: atomic writes, plain reads and existence checks.
//
// WriteFile never creates parent directories on purpose: a missing parent is
// treated as a durability failure so callers can detect incomplete storage.
package state

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir creates the directory and all missing parents.
func EnsureDir(path string) error {
	if path == "" {
		return fmt.Errorf("state: empty directory path")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("state: create directory %s: %w", path, err)
	}
	return nil
}

// WriteFile atomically writes data to path. The parent directory must exist;
// otherwise the write fails and the caller can treat the value as not durable.
func WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("state: create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("state: write temp file %s: %w", tmpName, err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("state: sync temp file %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("state: close temp file %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("state: rename temp file to %s: %w", path, err)
	}
	return nil
}

// ReadFile reads and returns the content at path.
func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("state: read %s: %w", path, err)
	}
	return data, nil
}

// FileExists reports whether path exists as a regular file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
