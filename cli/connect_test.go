package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	data, err := os.ReadFile(filepath.Join(p.RunsDir(), "session-one.json"))
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
}
