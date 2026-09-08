package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/rudrakshkarpe/agentsmd-cli/session"
	"github.com/rudrakshkarpe/agentsmd-cli/task"
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

func TestCaptureHookCorrelatesSessionsThroughActiveTask(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	if _, err := task.Start(p, "task-17", "Repair retries", time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, sessionID := range []string{"one", "two"} {
		start, _ := json.Marshal(map[string]any{"session_id": sessionID, "cwd": p.Root, "hook_event_name": "SessionStart"})
		if err := captureHook(p.Root, "claude", start); err != nil {
			t.Fatal(err)
		}
		end, _ := json.Marshal(map[string]any{"session_id": sessionID, "cwd": p.Root, "hook_event_name": "SessionEnd"})
		if err := captureHook(p.Root, "claude", end); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(session.RunPath(p, "claude", sessionID))
		if err != nil {
			t.Fatal(err)
		}
		var trajectory schema.Trajectory
		if err := json.Unmarshal(data, &trajectory); err != nil {
			t.Fatal(err)
		}
		if trajectory.Task != "task-17" || trajectory.Metadata["task_source"] != "active-task" {
			t.Fatalf("trajectory=%+v", trajectory)
		}
	}
}
