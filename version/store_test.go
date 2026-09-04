package version_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/version"
)

func TestCommitDiffAndRevert(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	store := version.New(p)
	store.Now = func() time.Time { return time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC) }
	if err := os.WriteFile(p.ArtifactPath(), []byte("first\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	first, err := store.Commit("initial", "manual", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.ArtifactPath(), []byte("second\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := store.Commit("change", "learned", map[string]any{"token_delta": -20})
	if err != nil {
		t.Fatal(err)
	}
	diff, err := store.Diff(first.ID, second.ID)
	if err != nil || !strings.Contains(diff, "-first") || !strings.Contains(diff, "+second") {
		t.Fatalf("diff=%q err=%v", diff, err)
	}
	if _, err := store.Revert(first.ID); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p.ArtifactPath())
	if string(data) != "first\n" {
		t.Fatalf("artifact=%q", data)
	}
}

func TestRejectsVersionPathTraversal(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if _, err := version.New(p).Content("../../secret"); err == nil {
		t.Fatal("expected invalid version error")
	}
}
