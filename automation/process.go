package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/learning"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	reflector "github.com/rudrakshkarpe/agentsmd-cli/reflect"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

type Evaluation struct {
	ProposalID string    `json:"proposal_id"`
	RunID      string    `json:"run_id"`
	Command    []string  `json:"command"`
	Passed     bool      `json:"passed"`
	ExitCode   int       `json:"exit_code"`
	Stdout     string    `json:"stdout,omitempty"`
	Stderr     string    `json:"stderr,omitempty"`
	Time       time.Time `json:"time"`
}

type Result struct {
	ProposalID string      `json:"proposal_id,omitempty"`
	RuleID     string      `json:"rule_id,omitempty"`
	Verdict    string      `json:"verdict"`
	Evaluation *Evaluation `json:"evaluation,omitempty"`
}

func Process(ctx context.Context, p *project.Project, trajectoryPath string) (Result, error) {
	config, err := Load(p)
	if err != nil {
		return Result{}, err
	}
	if len(config.ReflectCommand) == 0 {
		return Result{Verdict: "automation-disabled"}, nil
	}
	data, err := os.ReadFile(trajectoryPath)
	if err != nil {
		return Result{}, err
	}
	var trajectory schema.Trajectory
	if err := json.Unmarshal(data, &trajectory); err != nil {
		return Result{}, fmt.Errorf("decode trajectory: %w", err)
	}
	if trajectory.Metadata == nil {
		trajectory.Metadata = map[string]string{}
	}
	proposal, reflected, err := learning.New(p).Learn(ctx, trajectory, reflector.Command{Argv: config.ReflectCommand, Timeout: 2 * time.Minute})
	if err != nil {
		return Result{}, err
	}
	result := Result{Verdict: string(reflected.Verdict)}
	if proposal == nil {
		return result, nil
	}
	result.ProposalID = proposal.ID
	if len(config.EvaluateCommand) == 0 {
		return result, nil
	}
	evaluation := evaluate(ctx, p, config.EvaluateCommand, *proposal, trajectory)
	result.Evaluation = &evaluation
	if err := saveEvaluation(p, evaluation); err != nil {
		return Result{}, err
	}
	trajectory.Commands = append(trajectory.Commands, schema.Command{Argv: config.EvaluateCommand, ExitCode: evaluation.ExitCode})
	if evaluation.Passed {
		trajectory.TestResults.Passed++
		trajectory.Metadata["outcome"] = "success"
	} else {
		trajectory.TestResults.Failed++
		trajectory.Metadata["outcome"] = "failure"
	}
	if err := saveTrajectory(trajectoryPath, trajectory); err != nil {
		return Result{}, err
	}
	if !config.AutoPromote || !evaluation.Passed || reflected.Confidence < config.MinConfidence {
		return result, nil
	}
	rule, duplicate, err := learning.New(p).Promote(proposal.ID)
	if err != nil {
		return Result{}, err
	}
	if duplicate != nil {
		result.RuleID = duplicate.ID
	} else {
		result.RuleID = rule.ID
	}
	return result, nil
}

func saveTrajectory(path string, value schema.Trajectory) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return project.AtomicWrite(path, append(data, '\n'), 0o644)
}

func evaluate(ctx context.Context, p *project.Project, argv []string, proposal schema.Proposal, trajectory schema.Trajectory) Evaluation {
	result := Evaluation{ProposalID: proposal.ID, RunID: trajectory.SessionID, Command: argv, ExitCode: -1, Time: time.Now().UTC()}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, argv[0], argv[1:]...)
	command.Dir = p.Root
	command.Env = append(os.Environ(), "AGENTSMD_PROPOSAL_ID="+proposal.ID, "AGENTSMD_RULE="+proposal.Text, "AGENTSMD_RUN_ID="+trajectory.SessionID)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	result.Stdout, result.Stderr = stdout.String(), stderr.String()
	if err == nil {
		result.ExitCode, result.Passed = 0, true
	} else if exit, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exit.ExitCode()
	}
	return result
}

func saveEvaluation(p *project.Project, value Evaluation) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return project.AtomicWrite(filepath.Join(p.EvaluationsDir(), value.ProposalID+".json"), append(data, '\n'), 0o644)
}
