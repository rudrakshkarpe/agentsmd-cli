package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

func TestCaptureHookStoresNormalizedTrajectory(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	event, _ := json.Marshal(map[string]any{"session_id": "session/one", "cwd": p.Root, "hook_event_name": "SessionEnd", "model": "example"})
	if err := captureHook(p.Root, "codex", event); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(p.RunsDir(), "codex-session-one.json"))
	if err != nil {
		t.Fatal(err)
	}
	var trajectory schema.Trajectory
	if err := json.Unmarshal(data, &trajectory); err != nil {
		t.Fatal(err)
	}
	if trajectory.SessionID != "session/one" || trajectory.Tool != "codex" || trajectory.Metadata["hook_event"] != "SessionEnd" {
		t.Fatalf("trajectory=%+v", trajectory)
	}
	command := New()
	var output strings.Builder
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"--root", p.Root, "sessions"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "codex-session-one") {
		t.Fatalf("sessions=%q", output.String())
	}
}

// Payload fields match KlaatCode's interactive runHooks contract. The hook
// process may run outside the project; project_root must choose the workspace.
func TestKlaatCodeLifecycleUsesProjectRoot(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	start, _ := json.Marshal(map[string]any{"event": "session_start", "session_id": "klaatu-one", "project_root": p.Root})
	if err := receiveHook(outside, "klaatcode", start); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(p.RunsDir(), "klaatcode-klaatu-one.json")); !os.IsNotExist(err) {
		t.Fatalf("start produced a completed run: %v", err)
	}
	end, _ := json.Marshal(map[string]any{"event": "session_end", "session_id": "klaatu-one", "project_root": p.Root})
	if err := captureHook(outside, "klaatcode", end); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(p.RunsDir(), "klaatcode-klaatu-one.json"))
	if err != nil {
		t.Fatal(err)
	}
	var trajectory schema.Trajectory
	if err := json.Unmarshal(data, &trajectory); err != nil {
		t.Fatal(err)
	}
	if trajectory.Tool != "klaatcode" || trajectory.SessionID != "klaatu-one" || trajectory.Metadata["started_at"] == "" || trajectory.Metadata["hook_event"] != "session_end" {
		t.Fatalf("trajectory=%+v", trajectory)
	}
	if len(trajectory.Steps) != 0 || len(trajectory.ToolCalls) != 0 {
		t.Fatal("lifecycle event invented transcript evidence")
	}
}
