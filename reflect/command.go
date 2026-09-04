package reflect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

// Command delegates reflection to any local program that accepts a trajectory
// as JSON on stdin and returns a Result as JSON on stdout.
type Command struct {
	Argv    []string
	Timeout time.Duration
}

func (c Command) Reflect(ctx context.Context, trajectory schema.Trajectory) (Result, error) {
	if len(c.Argv) == 0 {
		return Result{}, fmt.Errorf("reflect command cannot be empty")
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	input, err := json.Marshal(trajectory)
	if err != nil {
		return Result{}, err
	}
	command := exec.CommandContext(ctx, c.Argv[0], c.Argv[1:]...)
	command.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return Result{}, fmt.Errorf("reflect command timed out: %w", ctx.Err())
		}
		return Result{}, fmt.Errorf("reflect command failed: %w: %s", err, stderr.String())
	}
	var result Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return Result{}, fmt.Errorf("decode reflect result: %w", err)
	}
	if err := Validate(result); err != nil {
		return Result{}, err
	}
	return result, nil
}

func Validate(result Result) error {
	switch result.Verdict {
	case MissingRule, WrongRule, StaleRule:
		if result.Rule == "" {
			return fmt.Errorf("verdict %s requires one targeted rule", result.Verdict)
		}
	case NotRelevant:
		if result.Rule != "" {
			return fmt.Errorf("not-an-AGENTS.md-problem verdict cannot include a rule")
		}
	default:
		return fmt.Errorf("unknown reflection verdict %q", result.Verdict)
	}
	if result.Confidence < 0 || result.Confidence > 1 {
		return fmt.Errorf("reflection confidence must be between 0 and 1")
	}
	return nil
}
