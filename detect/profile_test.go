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
	for _, expected := range []string{
		"Detected stack: Node.js",
		"Use `pnpm` for JavaScript dependencies",
		"Bootstrap dependencies with `pnpm install`",
		"Test: `pnpm run test`",
		"Lint: `pnpm run lint`",
		"## Testing instructions",
		"## Code and dependency conventions",
		"## Agent working agreement",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("rendered profile missing %q:\n%s", expected, content)
		}
	}
}

func TestInspectWorkspaceAddsRootAndOwnershipGuidance(t *testing.T) {
	root := t.TempDir()
	manifest := `{"workspaces":["packages/*"],"scripts":{"build":"turbo build","typecheck":"tsc --noEmit"}}`
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package-lock.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	profile, err := detect.Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	content := profile.Render()
	for _, expected := range []string{
		"This is a workspace or monorepo",
		"Run dependency commands from the workspace root",
		"Typecheck: `npm run typecheck`",
		"Build: `npm run build`",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("rendered workspace profile missing %q:\n%s", expected, content)
		}
	}
}

func TestInspectEmptyProjectUsesUniversalBaselineWithoutInventingCommands(t *testing.T) {
	profile, err := detect.Inspect(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	content := profile.Render()
	if !strings.Contains(content, "No supported manifest was detected yet") {
		t.Fatalf("missing new-project guidance:\n%s", content)
	}
	if strings.Contains(content, "## Commands and validation") {
		t.Fatalf("empty project should not receive invented commands:\n%s", content)
	}
	if !strings.Contains(content, "## Agent working agreement") {
		t.Fatalf("missing universal working agreement:\n%s", content)
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
