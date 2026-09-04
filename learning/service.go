// Package learning implements the propose-review-promote loop.
package learning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/ledger"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	reflector "github.com/rudrakshkarpe/agentsmd-cli/reflect"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/rudrakshkarpe/agentsmd-cli/version"
)

type Service struct {
	Project *project.Project
	Now     func() time.Time
}

func (s *Service) Learn(ctx context.Context, trajectory schema.Trajectory, engine reflector.Reflector) (*schema.Proposal, reflector.Result, error) {
	result, err := engine.Reflect(ctx, trajectory)
	if err != nil {
		return nil, reflector.Result{}, err
	}
	if err := reflector.Validate(result); err != nil {
		return nil, result, err
	}
	if result.Verdict == reflector.NotRelevant {
		return nil, result, nil
	}
	origin := result.Origin
	if origin.Run == "" {
		origin.Run = trajectory.SessionID
	}
	if origin.Task == "" {
		origin.Task = trajectory.Task
	}
	proposal, err := s.Propose(result.Rule, origin)
	if err != nil {
		return nil, result, err
	}
	return &proposal, result, nil
}

func New(p *project.Project) *Service {
	return &Service{Project: p, Now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Propose(text string, origin schema.Origin) (schema.Proposal, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return schema.Proposal{}, fmt.Errorf("proposal text cannot be empty")
	}
	proposal := schema.Proposal{
		ID:       fmt.Sprintf("p%d", s.Now().UnixNano()),
		Text:     text,
		Origin:   origin,
		Proposed: s.Now(),
	}
	data, err := json.MarshalIndent(proposal, "", "  ")
	if err != nil {
		return schema.Proposal{}, err
	}
	return proposal, project.AtomicWrite(filepath.Join(s.Project.PendingDir(), proposal.ID+".json"), append(data, '\n'), 0o644)
}

func (s *Service) Pending() ([]schema.Proposal, error) {
	entries, err := os.ReadDir(s.Project.PendingDir())
	if errors.Is(err, os.ErrNotExist) {
		return []schema.Proposal{}, nil
	}
	if err != nil {
		return nil, err
	}
	result := []schema.Proposal{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.Project.PendingDir(), entry.Name()))
		if err != nil {
			return nil, err
		}
		var proposal schema.Proposal
		if err := json.Unmarshal(data, &proposal); err != nil {
			return nil, fmt.Errorf("decode proposal %s: %w", entry.Name(), err)
		}
		result = append(result, proposal)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Proposed.Before(result[j].Proposed) })
	return result, nil
}

func (s *Service) Promote(id string) (*schema.Rule, *schema.Rule, error) {
	proposal, path, err := s.read(id)
	if err != nil {
		return nil, nil, err
	}
	value, err := s.Project.LoadLedger()
	if err != nil {
		return nil, nil, err
	}
	rule, duplicate, err := ledger.Add(&value, proposal.Text, proposal.Origin)
	if err != nil {
		return nil, nil, err
	}
	if duplicate != nil {
		return nil, duplicate, os.Remove(path)
	}
	if err := s.Project.SaveLedger(value); err != nil {
		return nil, nil, err
	}
	if err := project.AtomicWrite(s.Project.ArtifactPath(), []byte(ledger.Render(value)), 0o644); err != nil {
		return nil, nil, err
	}
	meta := map[string]any{"run": proposal.Origin.Run, "task": proposal.Origin.Task}
	if _, err := version.New(s.Project).Commit("learned: "+rule.Text, "learned", meta); err != nil {
		return nil, nil, err
	}
	return rule, nil, os.Remove(path)
}

func (s *Service) Reject(id string) error {
	_, path, err := s.read(id)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (s *Service) read(id string) (schema.Proposal, string, error) {
	if !strings.HasPrefix(id, "p") || strings.ContainsAny(id, `/\\`) {
		return schema.Proposal{}, "", fmt.Errorf("invalid proposal id %q", id)
	}
	path := filepath.Join(s.Project.PendingDir(), id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return schema.Proposal{}, "", fmt.Errorf("read proposal %s: %w", id, err)
	}
	var proposal schema.Proposal
	if err := json.Unmarshal(data, &proposal); err != nil {
		return schema.Proposal{}, "", err
	}
	return proposal, path, nil
}

type Savings struct {
	Task    string
	First   int
	Last    int
	Percent float64
	Runs    int
}

func (s *Service) Savings(task string) (*Savings, error) {
	value, err := s.Project.LoadLedger()
	if err != nil {
		return nil, err
	}
	runs := value.Runs[task]
	if len(runs) < 2 {
		return nil, nil
	}
	if runs[0] == 0 {
		return nil, fmt.Errorf("first token count cannot be zero")
	}
	return &Savings{Task: task, First: runs[0], Last: runs[len(runs)-1], Percent: 100 * float64(runs[0]-runs[len(runs)-1]) / float64(runs[0]), Runs: len(runs)}, nil
}

func (s *Service) Record(task string, tokens int) error {
	if strings.TrimSpace(task) == "" {
		return fmt.Errorf("task cannot be empty")
	}
	if tokens <= 0 {
		return fmt.Errorf("tokens must be positive")
	}
	value, err := s.Project.LoadLedger()
	if err != nil {
		return err
	}
	value.Runs[task] = append(value.Runs[task], tokens)
	return s.Project.SaveLedger(value)
}
