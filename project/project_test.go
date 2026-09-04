package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

func TestScaffoldAndFindFromNestedDirectory(t *testing.T) {
	root := t.TempDir()
	p, err := project.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "src", "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	found, err := project.Find(nested)
	if err != nil {
		t.Fatal(err)
	}
	if found != root {
		t.Fatalf("found %s, want %s", found, root)
	}
	if _, err := os.Stat(filepath.Join(root, ".agentsmd", "config.yaml")); err != nil {
		t.Fatal("config was not created")
	}
}

func TestLedgerRoundTrip(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	value, err := p.LoadLedger()
	if err != nil {
		t.Fatal(err)
	}
	value.Runs["demo"] = []int{100, 80}
	if err := p.SaveLedger(value); err != nil {
		t.Fatal(err)
	}
	loaded, err := p.LoadLedger()
	if err != nil || loaded.Runs["demo"][1] != 80 {
		t.Fatalf("loaded=%v err=%v", loaded, err)
	}
}
