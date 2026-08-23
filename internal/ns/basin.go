// Package ns models the watershed namespace: basins that stations belong to.
package ns

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Basin is a watershed namespace that groups monitoring stations.
type Basin struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
}

// NewBasin builds a basin with a fresh identifier.
func NewBasin(name, region string) (Basin, error) {
	basin := Basin{
		ID:        uuid.NewString(),
		Name:      strings.TrimSpace(name),
		Region:    strings.TrimSpace(region),
		CreatedAt: time.Now().UTC(),
	}
	if err := basin.Validate(); err != nil {
		return Basin{}, err
	}
	return basin, nil
}

// Validate checks that the basin carries a name and a region.
func (b Basin) Validate() error {
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("ns: basin name is required")
	}
	if strings.TrimSpace(b.Region) == "" {
		return fmt.Errorf("ns: basin region is required")
	}
	return nil
}
