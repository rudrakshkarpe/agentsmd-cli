package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/integration"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

func TestConnectAllProviders(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	claudePath := filepath.Join(p.Root, ".claude", "settings.local.json")
	if err := project.AtomicWrite(claudePath, []byte(`{"permissions":{"allow":["Bash(go test ./...)"]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, provider := range integration.Supported {
		if _, err := integration.Connect(p, provider); err != nil {
			t.Fatalf("connect %s: %v", provider, err)
		}
	}
	data, _ := os.ReadFile(claudePath)
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	if settings["permissions"] == nil || settings["hooks"] == nil {
		t.Fatalf("existing Claude settings were not preserved: %s", data)
	}
	records, err := integration.Load(p)
	if err != nil || len(records) != len(integration.Supported) {
		t.Fatalf("records=%v err=%v", records, err)
	}
	for _, path := range []string{
		filepath.Join(p.Root, ".codex", "hooks.json"),
		filepath.Join(p.Root, ".cursor", "hooks.json"),
		filepath.Join(p.Root, ".agents", "plugins", "agentsmd", "plugin.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}
	gooseData, _ := os.ReadFile(filepath.Join(p.Root, ".agents", "plugins", "agentsmd", "hooks", "hooks.json"))
	var gooseSettings struct {
		Hooks map[string][]struct {
			Hooks []map[string]any `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(gooseData, &gooseSettings); err != nil {
		t.Fatal(err)
	}
	if len(gooseSettings.Hooks["SessionEnd"]) != 1 || len(gooseSettings.Hooks["SessionEnd"][0].Hooks) != 1 {
		t.Fatalf("invalid goose hook nesting: %s", gooseData)
	}
}

func TestConnectIsIdempotent(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := integration.Connect(p, "codex"); err != nil {
			t.Fatal(err)
		}
	}
	data, _ := os.ReadFile(filepath.Join(p.Root, ".codex", "hooks.json"))
	var settings struct {
		Hooks map[string][]any `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	if len(settings.Hooks["SessionEnd"]) != 1 {
		t.Fatalf("duplicate hooks: %s", data)
	}
	if len(settings.Hooks["SessionStart"]) != 1 {
		t.Fatalf("missing lifecycle start hook: %s", data)
	}
}

func TestKlaatCodePreservesHooksAndReconnects(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(p.Root, ".klaatai", "hooks.json")
	original := `{"session_start":["echo existing"],"session_end":[{"command":"echo done","timeout":2}],"before_tool":[{"command":"check","matcher":"write"}]}`
	if err := project.AtomicWrite(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		record, err := integration.Connect(p, "klaatcode")
		if err != nil {
			t.Fatal(err)
		}
		if record.Path != filepath.Join(".klaatai", "hooks.json") {
			t.Fatalf("record=%+v", record)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var hooks map[string][]any
	if err := json.Unmarshal(data, &hooks); err != nil {
		t.Fatal(err)
	}
	for _, event := range []string{"session_start", "session_end"} {
		if len(hooks[event]) != 2 {
			t.Fatalf("duplicate or missing %s: %s", event, data)
		}
		entry := hooks[event][1].(map[string]any)
		if entry["command"] != "agentsmd hook klaatcode" || entry["timeout"] != float64(10) {
			t.Fatalf("invalid entry: %v", entry)
		}
	}
	if hooks["session_start"][0] != "echo existing" || len(hooks["before_tool"]) != 1 {
		t.Fatalf("existing hooks lost: %s", data)
	}
	if hooks["session_end"][0].(map[string]any)["command"] != "echo done" || hooks["before_tool"][0].(map[string]any)["matcher"] != "write" {
		t.Fatalf("existing hooks changed: %s", data)
	}
}

func TestConnectPreservesPrivateHookPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not implement POSIX permission bits")
	}
	for _, provider := range []string{"klaatcode", "claude", "codex", "cursor"} {
		t.Run(provider, func(t *testing.T) {
			p, _ := project.Open(t.TempDir())
			if err := p.Scaffold(); err != nil {
				t.Fatal(err)
			}
			record, err := integration.Connect(p, provider)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(p.Root, record.Path)
			if err := os.Chmod(path, 0o600); err != nil {
				t.Fatal(err)
			}
			for range 2 {
				if _, err := integration.Connect(p, provider); err != nil {
					t.Fatal(err)
				}
				info, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				if info.Mode().Perm() != 0o600 {
					t.Fatalf("hook permissions widened to %o", info.Mode().Perm())
				}
			}
		})
	}
}
