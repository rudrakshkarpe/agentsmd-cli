// Package capture defines vendor-neutral trajectory adapters.
package capture

import (
	"context"

	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

type Capabilities struct {
	Hooks      bool
	Transcript string
	Events     []string
}

type Adapter interface {
	Name() string
	Capabilities() Capabilities
	Latest(context.Context) (*schema.Trajectory, error)
}
