package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/cli"
)

func execute(t *testing.T, args ...string) string {
	t.Helper()
	command := cli.New()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs(args)
	if err := command.Execute(); err != nil {
		t.Fatalf("agentsmd %v: %v\n%s", args, err, output.String())
	}
	return output.String()
}

func TestPrimaryCLIExperience(t *testing.T) {
	welcome := execute(t)
	if !strings.Contains(welcome, "●●●") || !strings.Contains(welcome, "agentsmd doctor") {
		t.Fatalf("welcome output=%q", welcome)
	}
	templates := execute(t, "templates")
	if !strings.Contains(templates, "team") || !strings.Contains(templates, "monorepo") {
		t.Fatalf("templates output=%q", templates)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	execute(t, "--root", root, "init")
	data, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if !strings.Contains(string(data), "Detected stack: Go") || !strings.Contains(string(data), "go test ./...") {
		t.Fatalf("auto-generated AGENTS.md:\n%s", data)
	}
	output := execute(t, "--root", root, "connect", "codex")
	if !strings.Contains(output, "connected codex") || !strings.Contains(output, "/hooks") {
		t.Fatalf("connect output=%q", output)
	}
	if _, err := os.Stat(filepath.Join(root, ".codex", "hooks.json")); err != nil {
		t.Fatal(err)
	}
	doctor := execute(t, "--root", root, "doctor")
	if !strings.Contains(doctor, "AGENTS.md") || !strings.Contains(doctor, "Codex") {
		t.Fatalf("doctor output=%q", doctor)
	}
}

func TestTaskBoundaryLearningDemo(t *testing.T) {
	root := t.TempDir()
	output := execute(t, "--root", root, "init", "--scratch")
	if !strings.Contains(output, "v0000") {
		t.Fatalf("init output=%q", output)
	}
	output = execute(t, "--root", root, "learn", "--rule", "Run focused tests first.", "--run", "session-1", "--task", "demo")
	fields := strings.Fields(output)
	if len(fields) != 2 || fields[0] != "proposed" {
		t.Fatalf("learn output=%q", output)
	}
	proposalID := fields[1]
	output = execute(t, "--root", root, "promote", proposalID)
	if !strings.Contains(output, "r000") {
		t.Fatalf("promote output=%q", output)
	}
	output = execute(t, "--root", root, "blame")
	if !strings.Contains(output, "run=session-1 task=demo") {
		t.Fatalf("blame output=%q", output)
	}
}
