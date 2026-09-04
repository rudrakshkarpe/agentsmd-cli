// Package reflect defines the task-boundary reflection interface.
package reflect

import (
	"context"

	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

type Verdict string

const (
	MissingRule Verdict = "missing_rule"
	WrongRule   Verdict = "wrong_rule"
	StaleRule   Verdict = "stale_rule"
	NotRelevant Verdict = "not_an_agentsmd_problem"
)

type Result struct {
	Verdict    Verdict       `json:"verdict"`
	Rule       string        `json:"rule,omitempty"`
	Confidence float64       `json:"confidence"`
	Origin     schema.Origin `json:"origin"`
	Rationale  string        `json:"rationale"`
}

type Reflector interface {
	Reflect(context.Context, schema.Trajectory) (Result, error)
}
