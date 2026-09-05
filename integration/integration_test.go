package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	if err != nil || len(records) != 4 {
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
