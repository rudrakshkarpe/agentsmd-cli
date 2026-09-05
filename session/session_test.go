package session_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/rudrakshkarpe/agentsmd-cli/session"
)

func TestLifecycleCapturesGitAndDurationEvidence(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	runGit(t, p.Root, "init")
	runGit(t, p.Root, "config", "user.email", "test@example.com")
	runGit(t, p.Root, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(p.Root, "demo.txt"), []byte("before\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, p.Root, "add", "demo.txt")
	runGit(t, p.Root, "commit", "-m", "initial")
	if err := os.WriteFile(filepath.Join(p.Root, "preexisting.txt"), []byte("leave me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 9, 5, 1, 0, 0, 0, time.UTC)
	if err := session.Start(p, "codex", "s/1", start); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Root, "demo.txt"), []byte("after\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	trajectory := schema.Trajectory{SessionID: "s/1", Metadata: map[string]string{}}
	if err := session.Complete(p, &trajectory, "codex", start.Add(90*time.Second)); err != nil {
		t.Fatal(err)
	}
	if trajectory.WallTimeS != 90 || !strings.Contains(trajectory.FinalDiff, "+after") || len(trajectory.Files) != 1 {
		t.Fatalf("trajectory=%+v", trajectory)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
