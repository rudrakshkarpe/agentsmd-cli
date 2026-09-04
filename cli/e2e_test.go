package cli_test

import (
	"bytes"
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
