package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunnerPreservesEvidenceAndScoresBothConditions(t *testing.T) {
	if os.Getenv("AGENTSMDBENCH_HELPER") == "1" {
		helperAgent(t)
		return
	}
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "fixture", "go.mod"), "module example.com/fixture\n\ngo 1.23\n")
	mustWrite(t, filepath.Join(root, "fixture", "main.go"), "package main\nfunc main() {}\n")
	mustWrite(t, filepath.Join(root, "baseline.md"), "# AGENTS.md\n\nBaseline.\n")
	mustWrite(t, filepath.Join(root, "learned.md"), "# AGENTS.md\n\nLearned.\n")
	mustWrite(t, filepath.Join(root, "heldout", "main_test.go"), "package main\nimport \"testing\"\nfunc TestFixed(t *testing.T) { if fixed != true { t.Fatal(\"not fixed\") } }\n")
	spec := Spec{Name: "fixture", Task: "fix", Prompt: "fix it", Fixture: "fixture", Baseline: "baseline.md", Learned: "learned.md", HeldOut: "heldout", Verify: []string{"go", "test", "./..."}, Agent: "test", Model: "fake"}
	runner := Runner{AgentCommand: []string{os.Args[0], "-test.run=TestRunnerPreservesEvidenceAndScoresBothConditions"}, Seeds: 1, OutputDir: filepath.Join(root, "results")}
	t.Setenv("AGENTSMDBENCH_HELPER", "1")
	report, err := runner.Run(context.Background(), filepath.Join(root, "spec.json"), spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Runs) != 2 {
		t.Fatalf("runs=%d", len(report.Runs))
	}
	for _, run := range report.Runs {
		if !run.Passed || run.TotalTokens() != 130 || run.Commands != 1 {
			t.Fatalf("run=%+v", run)
		}
		if runDir := filepath.Join(root, "results", run.ID); commandExitCode(t, runDir) != 7 {
			t.Fatalf("command exit code was not preserved for %s", run.ID)
		}
		if _, err := os.Stat(filepath.Join(root, "results", run.ID, "workspace", ".cache")); !os.IsNotExist(err) {
			t.Fatalf("ignored build cache was retained for %s", run.ID)
		}
		for _, name := range []string{"agent.jsonl", "trajectory.json", "verifier.txt"} {
			if _, err := os.Stat(filepath.Join(root, "results", run.ID, name)); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func helperAgent(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("main.go", append(data, []byte("\nvar fixed = true\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".cache", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".gitignore", []byte(".cache/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".cache/build", []byte("temporary\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fmt.Println(`{"type":"item.completed","item":{"type":"command_execution","command":"apply patch","exit_code":7}}`)
	fmt.Println(`{"type":"turn.completed","usage":{"input_tokens":100,"cached_input_tokens":40,"output_tokens":30}}`)
}

func commandExitCode(t *testing.T, runDir string) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(runDir, "trajectory.json"))
	if err != nil {
		t.Fatal(err)
	}
	var trajectory struct {
		Commands []struct {
			ExitCode int `json:"exit_code"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(data, &trajectory); err != nil {
		t.Fatal(err)
	}
	if len(trajectory.Commands) != 1 {
		t.Fatalf("commands=%d", len(trajectory.Commands))
	}
	return trajectory.Commands[0].ExitCode
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
