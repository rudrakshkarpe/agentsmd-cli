package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

func TestTaskCommandManagesExplicitCorrelation(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	output := executeCLI(t, p.Root, "task", "start", "Fix config precedence", "--id", "config-42")
	if !strings.Contains(output, "config-42") {
		t.Fatalf("start=%q", output)
	}
	output = executeCLI(t, p.Root, "task")
	if !strings.Contains(output, "config-42\tFix config precedence") {
		t.Fatalf("show=%q", output)
	}
	output = executeCLI(t, p.Root, "task", "end")
	if !strings.Contains(output, "ended task config-42") {
		t.Fatalf("end=%q", output)
	}
}

func TestProgressComparesTaskSessionsChronologically(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	writeTrajectory(t, p, "claude-first", schema.Trajectory{
		SessionID: "first", Tool: "claude", Task: "config-42", WallTimeS: 100,
		Tokens: schema.Tokens{Input: 800, Output: 200}, TestResults: schema.TestResults{Failed: 1},
		Metadata: map[string]string{"ended_at": "2026-09-09T01:00:00Z", "outcome": "failure"},
	})
	writeTrajectory(t, p, "cursor-second", schema.Trajectory{
		SessionID: "second", Tool: "cursor", Task: "config-42", WallTimeS: 50,
		Tokens: schema.Tokens{Input: 500, Output: 100}, TestResults: schema.TestResults{Passed: 1}, Files: []schema.FileTouch{{Path: "config.go"}},
		Metadata: map[string]string{"ended_at": "2026-09-09T02:00:00Z", "outcome": "success"},
	})
	output := executeCLI(t, p.Root, "progress", "config-42")
	for _, wanted := range []string{"runs=2", "success=1/2", "100.0s→50.0s (-50.0%)", "1000→600 (-40.0%)", "claude-first", "cursor-second"} {
		if !strings.Contains(output, wanted) {
			t.Fatalf("progress missing %q: %s", wanted, output)
		}
	}
}

func writeTrajectory(t *testing.T, p *project.Project, id string, trajectory schema.Trajectory) {
	t.Helper()
	data, err := json.Marshal(trajectory)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.RunsDir(), id+".json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func executeCLI(t *testing.T, root string, args ...string) string {
	t.Helper()
	command := New()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs(append([]string{"--root", root}, args...))
	if err := command.Execute(); err != nil {
		t.Fatalf("agentsmd %v: %v\n%s", args, err, output.String())
	}
	return output.String()
}
