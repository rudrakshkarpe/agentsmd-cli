// Package benchmark runs reproducible before/after coding-agent evaluations.
package benchmark

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/rudrakshkarpe/agentsmd-cli/session"
)

type Spec struct {
	Name        string   `json:"name"`
	Task        string   `json:"task"`
	Prompt      string   `json:"prompt"`
	Fixture     string   `json:"fixture"`
	Baseline    string   `json:"baseline_agents_md"`
	Learned     string   `json:"learned_agents_md"`
	HeldOut     string   `json:"held_out"`
	Verify      []string `json:"verify"`
	Agent       string   `json:"agent"`
	Model       string   `json:"model"`
	Reasoning   string   `json:"reasoning_effort,omitempty"`
	Description string   `json:"description,omitempty"`
}

type Runner struct {
	AgentCommand []string
	Seeds        int
	OutputDir    string
	Timeout      time.Duration
}

type Run struct {
	ID             string   `json:"id"`
	Condition      string   `json:"condition"`
	Trial          int      `json:"trial"`
	Passed         bool     `json:"passed"`
	AgentExitCode  int      `json:"agent_exit_code"`
	VerifyExitCode int      `json:"verify_exit_code"`
	InputTokens    int      `json:"input_tokens"`
	CachedTokens   int      `json:"cached_tokens"`
	OutputTokens   int      `json:"output_tokens"`
	Commands       int      `json:"commands"`
	FilesChanged   []string `json:"files_changed"`
	DurationS      float64  `json:"duration_s"`
	ArtifactDir    string   `json:"artifact_dir"`
}

func (r Run) TotalTokens() int { return r.InputTokens + r.OutputTokens }

type Report struct {
	SchemaVersion int       `json:"schema_version"`
	Spec          Spec      `json:"spec"`
	GeneratedAt   time.Time `json:"generated_at"`
	Runs          []Run     `json:"runs"`
}

func LoadSpec(path string) (Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return Spec{}, fmt.Errorf("decode benchmark spec: %w", err)
	}
	if spec.Name == "" || spec.Task == "" || spec.Prompt == "" || spec.Fixture == "" || spec.Baseline == "" || spec.Learned == "" || len(spec.Verify) == 0 {
		return Spec{}, fmt.Errorf("benchmark spec requires name, task, prompt, fixture, both AGENTS.md conditions, and verify command")
	}
	if spec.Agent == "" {
		spec.Agent = "agent"
	}
	return spec, nil
}

func (r *Runner) Run(ctx context.Context, specPath string, spec Spec) (Report, error) {
	if len(r.AgentCommand) == 0 {
		return Report{}, fmt.Errorf("agent command cannot be empty")
	}
	if r.Seeds < 1 {
		return Report{}, fmt.Errorf("trials must be positive")
	}
	if r.Timeout <= 0 {
		r.Timeout = 10 * time.Minute
	}
	base := filepath.Dir(specPath)
	if r.OutputDir == "" {
		r.OutputDir = filepath.Join(base, "results", time.Now().UTC().Format("20060102T150405Z"))
	}
	if err := os.MkdirAll(r.OutputDir, 0o755); err != nil {
		return Report{}, err
	}
	report := Report{SchemaVersion: 1, Spec: spec, GeneratedAt: time.Now().UTC(), Runs: []Run{}}
	conditions := []struct{ name, guide string }{{"baseline", spec.Baseline}, {"learned", spec.Learned}}
	for _, condition := range conditions {
		for trial := 1; trial <= r.Seeds; trial++ {
			result, err := r.runOne(ctx, base, spec, condition.name, condition.guide, trial)
			if err != nil {
				return report, err
			}
			report.Runs = append(report.Runs, result)
			if err := saveReport(r.OutputDir, report); err != nil {
				return report, err
			}
		}
	}
	return report, saveReport(r.OutputDir, report)
}

func (r *Runner) runOne(ctx context.Context, base string, spec Spec, condition, guide string, trial int) (Run, error) {
	id := condition + "-" + strconv.Itoa(trial)
	artifactDir := filepath.Join(r.OutputDir, id)
	workspace := filepath.Join(artifactDir, "workspace")
	if err := copyTree(filepath.Join(base, spec.Fixture), workspace); err != nil {
		return Run{}, err
	}
	guideData, err := os.ReadFile(filepath.Join(base, guide))
	if err != nil {
		return Run{}, err
	}
	if err := os.WriteFile(filepath.Join(workspace, "AGENTS.md"), guideData, 0o644); err != nil {
		return Run{}, err
	}
	if err := initializeGit(workspace); err != nil {
		return Run{}, err
	}
	p, _ := project.Open(workspace)
	if err := p.Scaffold(); err != nil {
		return Run{}, err
	}
	start := time.Now().UTC()
	if err := session.Start(p, spec.Agent, id, start); err != nil {
		return Run{}, err
	}
	runCtx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	command := exec.CommandContext(runCtx, r.AgentCommand[0], r.AgentCommand[1:]...)
	command.Dir = workspace
	command.Stdin = strings.NewReader(spec.Prompt + "\n")
	command.Env = append(os.Environ(), "AGENTSMDBENCH_TRIAL="+strconv.Itoa(trial), "AGENTSMDBENCH_CONDITION="+condition)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	agentErr := command.Run()
	ended := time.Now().UTC()
	if err := os.WriteFile(filepath.Join(artifactDir, "agent.jsonl"), stdout.Bytes(), 0o644); err != nil {
		return Run{}, err
	}
	if err := os.WriteFile(filepath.Join(artifactDir, "agent.stderr.txt"), stderr.Bytes(), 0o644); err != nil {
		return Run{}, err
	}
	trajectory := parseJSONL(stdout.Bytes())
	trajectory.SessionID, trajectory.Tool, trajectory.Task = id, spec.Agent, spec.Task
	trajectory.Metadata = map[string]string{"condition": condition, "trial": strconv.Itoa(trial), "model": spec.Model}
	if err := session.Complete(p, &trajectory, spec.Agent, ended); err != nil {
		return Run{}, err
	}
	if spec.HeldOut != "" {
		if err := copyTree(filepath.Join(base, spec.HeldOut), workspace); err != nil {
			return Run{}, err
		}
	}
	verify := exec.CommandContext(ctx, spec.Verify[0], spec.Verify[1:]...)
	verify.Dir = workspace
	var verifyOut bytes.Buffer
	verify.Stdout, verify.Stderr = &verifyOut, &verifyOut
	verifyErr := verify.Run()
	if err := os.WriteFile(filepath.Join(artifactDir, "verifier.txt"), verifyOut.Bytes(), 0o644); err != nil {
		return Run{}, err
	}
	passed := agentErr == nil && verifyErr == nil
	if passed {
		trajectory.TestResults.Passed = 1
		trajectory.Metadata["outcome"] = "success"
	} else {
		trajectory.TestResults.Failed = 1
		trajectory.Metadata["outcome"] = "failure"
	}
	trajectoryData, _ := json.MarshalIndent(trajectory, "", "  ")
	if err := os.WriteFile(filepath.Join(artifactDir, "trajectory.json"), append(trajectoryData, '\n'), 0o644); err != nil {
		return Run{}, err
	}
	files := make([]string, 0, len(trajectory.Files))
	for _, file := range trajectory.Files {
		files = append(files, file.Path)
	}
	sort.Strings(files)
	// Keep the solved source tree for audit, but omit nested repository and
	// agentsmd runtime state; their durable evidence is stored beside it. The
	// workspace is disposable, so ignored build caches can be discarded first.
	clean := exec.Command("git", "clean", "-fdX")
	clean.Dir = workspace
	if output, err := clean.CombinedOutput(); err != nil {
		return Run{}, fmt.Errorf("clean ignored benchmark files: %w: %s", err, output)
	}
	if err := os.RemoveAll(filepath.Join(workspace, ".git")); err != nil {
		return Run{}, err
	}
	if err := os.RemoveAll(filepath.Join(workspace, project.DirName)); err != nil {
		return Run{}, err
	}
	return Run{ID: id, Condition: condition, Trial: trial, Passed: passed, AgentExitCode: exitCode(agentErr), VerifyExitCode: exitCode(verifyErr), InputTokens: trajectory.Tokens.Input, CachedTokens: trajectory.Tokens.Cached, OutputTokens: trajectory.Tokens.Output, Commands: len(trajectory.Commands), FilesChanged: files, DurationS: ended.Sub(start).Seconds(), ArtifactDir: id}, nil
}

func parseJSONL(data []byte) schema.Trajectory {
	trajectory := schema.Trajectory{Steps: []schema.Step{}, ToolCalls: []schema.ToolCall{}, Files: []schema.FileTouch{}, Commands: []schema.Command{}, Metadata: map[string]string{}}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 8<<20)
	for scanner.Scan() {
		var event struct {
			Type string `json:"type"`
			Item struct {
				Type, Command, Text string
				ExitCode            *int `json:"exit_code"`
			} `json:"item"`
		}
		var raw map[string]any
		if json.Unmarshal(scanner.Bytes(), &raw) != nil || json.Unmarshal(scanner.Bytes(), &event) != nil {
			continue
		}
		if usage, ok := raw["usage"].(map[string]any); ok {
			trajectory.Tokens.Input = int(number(usage["input_tokens"]))
			trajectory.Tokens.Cached = int(number(usage["cached_input_tokens"]))
			trajectory.Tokens.Output = int(number(usage["output_tokens"]))
		}
		if event.Type == "item.completed" && event.Item.Type == "command_execution" {
			exitCode := 0
			if event.Item.ExitCode != nil {
				exitCode = *event.Item.ExitCode
			}
			trajectory.Commands = append(trajectory.Commands, schema.Command{Argv: []string{"sh", "-lc", event.Item.Command}, ExitCode: exitCode})
		}
		if event.Type == "item.completed" && event.Item.Type == "agent_message" && event.Item.Text != "" {
			trajectory.Steps = append(trajectory.Steps, schema.Step{Role: "assistant", Summary: event.Item.Text})
		}
	}
	return trajectory
}

func number(value any) float64 { number, _ := value.(float64); return number }

func saveReport(dir string, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "report.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "report.md"), []byte(renderMarkdown(report)), 0o644)
}

func renderMarkdown(report Report) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# %s benchmark\n\n", report.Spec.Name)
	model := report.Spec.Model
	if report.Spec.Reasoning != "" {
		model += " (" + report.Spec.Reasoning + " reasoning)"
	}
	fmt.Fprintf(&out, "Model: `%s` · Trials per condition: %d · Generated: %s\n\n", model, count(report.Runs, "baseline"), report.GeneratedAt.Format(time.RFC3339))
	out.WriteString("| Condition | Trial | Passed | Tokens | Commands | Duration |\n|---|---:|:---:|---:|---:|---:|\n")
	for _, run := range report.Runs {
		fmt.Fprintf(&out, "| %s | %d | %t | %d | %d | %.1fs |\n", run.Condition, run.Trial, run.Passed, run.TotalTokens(), run.Commands, run.DurationS)
	}
	for _, condition := range []string{"baseline", "learned"} {
		selected := selectRuns(report.Runs, condition)
		fmt.Fprintf(&out, "\n**%s:** %.0f%% success, median %d tokens, median %d commands, median %.1fs.\n", condition, successRate(selected), medianInts(selected, func(r Run) int { return r.TotalTokens() }), medianInts(selected, func(r Run) int { return r.Commands }), medianDurations(selected))
	}
	return out.String()
}

func selectRuns(runs []Run, condition string) []Run {
	var result []Run
	for _, run := range runs {
		if run.Condition == condition {
			result = append(result, run)
		}
	}
	return result
}
func count(runs []Run, condition string) int { return len(selectRuns(runs, condition)) }
func successRate(runs []Run) float64 {
	if len(runs) == 0 {
		return 0
	}
	passed := 0
	for _, r := range runs {
		if r.Passed {
			passed++
		}
	}
	return 100 * float64(passed) / float64(len(runs))
}
func medianInts(runs []Run, value func(Run) int) int {
	values := make([]int, 0, len(runs))
	for _, r := range runs {
		values = append(values, value(r))
	}
	sort.Ints(values)
	if len(values) == 0 {
		return 0
	}
	return values[len(values)/2]
}
func medianDurations(runs []Run) float64 {
	values := make([]float64, 0, len(runs))
	for _, r := range runs {
		values = append(values, r.DurationS)
	}
	sort.Float64s(values)
	if len(values) == 0 {
		return 0
	}
	return values[len(values)/2]
}
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode()
	}
	return -1
}

func initializeGit(root string) error {
	commands := [][]string{{"init", "-q"}, {"config", "user.email", "benchmark@agentsmd.local"}, {"config", "user.name", "agentsmd benchmark"}, {"add", "."}, {"commit", "-qm", "benchmark fixture"}}
	for _, args := range commands {
		command := exec.Command("git", args...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("git %s: %w: %s", args[0], err, output)
		}
	}
	return nil
}

func copyTree(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("benchmark fixtures cannot contain symlinks: %s", path)
		}
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		output, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(output, input)
		inputErr := input.Close()
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		if inputErr != nil {
			return inputErr
		}
		return closeErr
	})
}
