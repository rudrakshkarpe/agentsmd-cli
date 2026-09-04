// Package integration installs project-local hooks for supported coding agents.
package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

var Supported = []string{"codex", "claude", "goose", "cursor"}

type Record struct {
	Provider string `json:"provider"`
	Path     string `json:"path"`
}

func Connect(p *project.Project, provider string) (Record, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	var path string
	var value any
	switch provider {
	case "codex":
		path = filepath.Join(p.Root, ".codex", "hooks.json")
		value = hookSettings("SessionEnd", "agentsmd hook codex", 3)
	case "claude":
		path = filepath.Join(p.Root, ".claude", "settings.local.json")
		value = hookSettings("SessionEnd", "agentsmd hook claude", 10)
	case "cursor":
		path = filepath.Join(p.Root, ".cursor", "hooks.json")
		value = map[string]any{"version": 1, "hooks": map[string]any{"sessionEnd": []any{map[string]any{"command": "agentsmd hook cursor", "timeout": 10}}}}
	case "goose":
		return connectGoose(p)
	default:
		return Record{}, fmt.Errorf("unsupported CLI %q; choose %s", provider, strings.Join(Supported, ", "))
	}
	if err := mergeJSON(path, value); err != nil {
		return Record{}, err
	}
	record := Record{Provider: provider, Path: path}
	return record, saveRecord(p, record)
}

func hookSettings(event, command string, timeout int) map[string]any {
	handler := map[string]any{"type": "command", "command": command, "timeout": timeout}
	return map[string]any{"hooks": map[string]any{event: []any{map[string]any{"hooks": []any{handler}}}}}
}

func connectGoose(p *project.Project) (Record, error) {
	dir := filepath.Join(p.Root, ".agents", "plugins", "agentsmd")
	manifest := map[string]any{"name": "agentsmd", "version": "1.0.0", "description": "Capture goose sessions for AGENTS.md improvement"}
	hooks := map[string]any{"hooks": map[string]any{"SessionEnd": []any{map[string]any{"type": "command", "command": "agentsmd hook goose", "timeout": 10}}}}
	if err := writeJSON(filepath.Join(dir, "plugin.json"), manifest); err != nil {
		return Record{}, err
	}
	path := filepath.Join(dir, "hooks", "hooks.json")
	if err := writeJSON(path, hooks); err != nil {
		return Record{}, err
	}
	record := Record{Provider: "goose", Path: path}
	return record, saveRecord(p, record)
}

func mergeJSON(path string, patch any) error {
	base := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &base); err != nil {
			return fmt.Errorf("decode existing %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	mergeMap(base, patch.(map[string]any))
	return writeJSON(path, base)
}

func mergeMap(dst, src map[string]any) {
	for key, value := range src {
		if incoming, ok := value.(map[string]any); ok {
			current, _ := dst[key].(map[string]any)
			if current == nil {
				current = map[string]any{}
			}
			mergeMap(current, incoming)
			dst[key] = current
			continue
		}
		if incoming, ok := value.([]any); ok {
			current, _ := dst[key].([]any)
			dst[key] = appendUnique(current, incoming...)
			continue
		}
		dst[key] = value
	}
}

func appendUnique(values []any, incoming ...any) []any {
	seen := map[string]bool{}
	for _, value := range values {
		data, _ := json.Marshal(value)
		seen[string(data)] = true
	}
	for _, value := range incoming {
		data, _ := json.Marshal(value)
		if !seen[string(data)] {
			values = append(values, value)
			seen[string(data)] = true
		}
	}
	return values
}

func saveRecord(p *project.Project, record Record) error {
	records, _ := Load(p)
	found := false
	for index := range records {
		if records[index].Provider == record.Provider {
			records[index] = record
			found = true
		}
	}
	if !found {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Provider < records[j].Provider })
	return writeJSON(p.ConnectionsPath(), records)
}

func Load(p *project.Project) ([]Record, error) {
	data, err := os.ReadFile(p.ConnectionsPath())
	if os.IsNotExist(err) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, err
	}
	var records []Record
	return records, json.Unmarshal(data, &records)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return project.AtomicWrite(path, append(data, '\n'), 0o644)
}
