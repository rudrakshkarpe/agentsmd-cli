package automation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/automation"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

func TestProcessReflectsEvaluatesAndPromotes(t *testing.T) {
	t.Setenv("GO_WANT_AUTOMATION_HELPER", "1")
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	if err := project.AtomicWrite(p.ArtifactPath(), []byte("# AGENTS.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	helper := []string{os.Args[0], "-test.run=TestAutomationHelper", "--"}
	config := automation.Config{
		ReflectCommand: append(helper, "reflect"), EvaluateCommand: append(helper, "evaluate"),
		AutoPromote: true, MinConfidence: 0.8,
	}
	if err := automation.Save(p, config); err != nil {
		t.Fatal(err)
	}
	trajectoryPath := filepath.Join(p.RunsDir(), "codex-s1.json")
	data, _ := json.Marshal(schema.Trajectory{SessionID: "s1", Tool: "codex", Metadata: map[string]string{}})
	if err := project.AtomicWrite(trajectoryPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := automation.Process(context.Background(), p, trajectoryPath)
	if err != nil {
		t.Fatal(err)
	}
	if result.ProposalID == "" || result.RuleID != "r000" || result.Evaluation == nil || !result.Evaluation.Passed {
		t.Fatalf("result=%+v", result)
	}
	agents, _ := os.ReadFile(p.ArtifactPath())
	if !strings.Contains(string(agents), "Run focused tests before the full suite") {
		t.Fatalf("AGENTS.md=%s", agents)
	}
	updated, _ := os.ReadFile(trajectoryPath)
	if !strings.Contains(string(updated), `"outcome": "success"`) {
		t.Fatalf("trajectory=%s", updated)
	}
}

func TestAutoPromotionRequiresGate(t *testing.T) {
	config := automation.DefaultConfig()
	config.AutoPromote = true
	config.ReflectCommand = []string{"reflect"}
	if automation.Validate(config) == nil {
		t.Fatal("expected missing evaluation command to fail")
	}
}

func TestProcessRedactsOnlyExternalReflectorCopy(t *testing.T) {
	t.Setenv("GO_WANT_AUTOMATION_HELPER", "1")
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	helper := []string{os.Args[0], "-test.run=TestAutomationHelper", "--", "reflect-redacted"}
	config := automation.DefaultConfig()
	config.ReflectCommand = helper
	config.RedactPatterns = []string{`sk-[A-Za-z0-9]+`}
	if err := automation.Save(p, config); err != nil {
		t.Fatal(err)
	}
	trajectoryPath := filepath.Join(p.RunsDir(), "claude-private.json")
	data, _ := json.Marshal(schema.Trajectory{SessionID: "private", Tool: "claude", Steps: []schema.Step{{Role: "assistant", Summary: "key sk-secret123"}}})
	if err := project.AtomicWrite(trajectoryPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := automation.Process(context.Background(), p, trajectoryPath)
	if err != nil || result.Verdict != "not_an_agentsmd_problem" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	stored, err := os.ReadFile(trajectoryPath)
	if err != nil || !strings.Contains(string(stored), "sk-secret123") {
		t.Fatalf("stored trajectory was redacted: %s err=%v", stored, err)
	}
}

func TestConfigRejectsInvalidRedactionPattern(t *testing.T) {
	config := automation.DefaultConfig()
	config.RedactPatterns = []string{"["}
	if err := automation.Validate(config); err == nil {
		t.Fatal("expected invalid redaction pattern to fail")
	}
	config.RedactPatterns = []string{""}
	if err := automation.Validate(config); err == nil {
		t.Fatal("expected empty redaction pattern to fail")
	}
}

func TestCompletedJobsAreIdempotent(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	trajectoryPath := filepath.Join(p.RunsDir(), "codex-s2.json")
	data, _ := json.Marshal(schema.Trajectory{SessionID: "s2", Tool: "codex"})
	if err := project.AtomicWrite(trajectoryPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	jobPath, err := automation.Enqueue(p, trajectoryPath)
	if err != nil {
		t.Fatal(err)
	}
	first, err := automation.ProcessJob(context.Background(), p, jobPath)
	if err != nil || first.Status != "complete" {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	if _, err := automation.Enqueue(p, trajectoryPath); err != nil {
		t.Fatal(err)
	}
	second, err := automation.ProcessJob(context.Background(), p, jobPath)
	if err != nil || second.Status != "complete" || second.Result.Verdict != "automation-disabled" {
		t.Fatalf("second=%+v err=%v", second, err)
	}
}

func TestAutomationHelper(t *testing.T) {
	if os.Getenv("GO_WANT_AUTOMATION_HELPER") != "1" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "reflect":
		fmt.Print(`{"verdict":"missing_rule","rule":"Run focused tests before the full suite.","confidence":0.95}`)
		os.Exit(0)
	case "evaluate":
		fmt.Print("evaluation passed")
		os.Exit(0)
	case "reflect-redacted":
		input, _ := io.ReadAll(os.Stdin)
		if strings.Contains(string(input), "sk-secret123") || !strings.Contains(string(input), "[REDACTED]") {
			fmt.Fprintln(os.Stderr, "trajectory was not redacted")
			os.Exit(3)
		}
		fmt.Print(`{"verdict":"not_an_agentsmd_problem","confidence":0.9}`)
		os.Exit(0)
	default:
		os.Exit(2)
	}
}
