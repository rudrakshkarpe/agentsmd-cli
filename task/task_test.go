package task_test

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/task"
)

func TestActiveTaskTakesPriorityOverBranch(t *testing.T) {
	t.Setenv("AGENTSMD_TASK_ID", "")
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	runGit(t, p.Root, "init", "-b", "feature/example")
	started, err := task.Start(p, "task-42", "Fix config precedence", time.Date(2026, 9, 9, 1, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if started.ID != "task-42" {
		t.Fatalf("started=%+v", started)
	}
	identity := task.Resolve(p, "", "claude", "session-1")
	if identity.ID != "task-42" || identity.Source != "active-task" {
		t.Fatalf("identity=%+v", identity)
	}
	ended, err := task.End(p)
	if err != nil || ended.ID != "task-42" {
		t.Fatalf("ended=%+v err=%v", ended, err)
	}
	identity = task.Resolve(p, "", "claude", "session-2")
	if identity.ID != "branch:feature/example" || identity.Source != "git-branch" {
		t.Fatalf("identity=%+v", identity)
	}
}

func TestResolveAvoidsGroupingDefaultBranchRuns(t *testing.T) {
	t.Setenv("AGENTSMD_TASK_ID", "")
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	runGit(t, p.Root, "init", "-b", "main")
	identity := task.Resolve(p, "", "goose", "session-9")
	if identity.ID != "run:goose:session-9" || identity.Source != "run" {
		t.Fatalf("identity=%+v", identity)
	}
	provided := task.Resolve(p, "external-7", "goose", "session-9")
	if provided.ID != "external-7" || provided.Source != "provider" {
		t.Fatalf("provided=%+v", provided)
	}
}

func TestSlug(t *testing.T) {
	if got := task.Slug(" Fix: Config / Precedence "); got != "fix-config-precedence" {
		t.Fatalf("slug=%q", got)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	command.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
