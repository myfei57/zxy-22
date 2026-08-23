// Package state: append-only JSON line journal used by audit and registries.
package state

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// Journal is an append-only file of JSON lines.
type Journal struct {
	path string
}

// NewJournal returns a journal writing to path. The parent directory must
// exist before the first Append call.
func NewJournal(path string) *Journal {
	return &Journal{path: path}
}

// Append marshals value and appends it as one JSON line.
func (j *Journal) Append(value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("state: marshal journal record: %w", err)
	}
	file, err := os.OpenFile(j.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("state: open journal %s: %w", j.path, err)
	}
	defer file.Close()
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("state: append journal %s: %w", j.path, err)
	}
	return nil
}

// Read returns every record currently stored in the journal.
func (j *Journal) Read() ([]json.RawMessage, error) {
	file, err := os.Open(j.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("state: open journal %s: %w", j.path, err)
	}
	defer file.Close()
	var records []json.RawMessage
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		record := make([]byte, len(line))
		copy(record, line)
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("state: scan journal %s: %w", j.path, err)
	}
	return records, nil
}
