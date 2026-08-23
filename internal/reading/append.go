// Package reading: durable append of one sample value.
package reading

import (
	"encoding/json"
	"fmt"

	"envmonitor/internal/state"
)

// Append durably stores the reading value. The write is atomic: the value is
// only considered stored once the JSON file has been renamed into place.
func Append(st *Store, rd Reading) error {
	data, err := json.Marshal(rd)
	if err != nil {
		return fmt.Errorf("reading: encode reading: %w", err)
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	if err := state.WriteFile(st.path(rd.ID), data); err != nil {
		return fmt.Errorf("reading: append value %s: %w", rd.ID, err)
	}
	return nil
}
