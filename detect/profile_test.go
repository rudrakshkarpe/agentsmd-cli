package detect_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/detect"
)

func TestInspectNodeProjectUsesLockfileAndManifestScripts(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"scripts":{"test":"vitest","lint":"eslint ."}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pnpm-lock.yaml"), []byte("lockfileVersion: 9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	profile, err := detect.Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	content := profile.Render()
	for _, expected := range []string{"Node.js", "pnpm install", "pnpm run test", "pnpm run lint"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("rendered profile missing %q:\n%s", expected, content)
		}
	}
}

func TestInspectMixedGoAndPythonProject(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"go.mod", "pyproject.toml", "uv.lock"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	profile, err := detect.Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	content := profile.Render()
	for _, expected := range []string{"Go, Python", "go test ./...", "uv sync", "python -m pytest"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("rendered profile missing %q:\n%s", expected, content)
		}
	}
}
