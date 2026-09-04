// Package project manages the on-disk .agentsmd project layout.
package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

const (
	DirName  = ".agentsmd"
	Artifact = "AGENTS.md"
)

type Project struct {
	Root string
}

func Find(start string) (string, error) {
	path, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		info, statErr := os.Stat(filepath.Join(path, DirName))
		if statErr == nil && info.IsDir() {
			return path, nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", os.ErrNotExist
		}
		path = parent
	}
}

func Open(root string) (*Project, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Project{Root: abs}, nil
}

func Require(start string) (*Project, error) {
	root, err := Find(start)
	if err != nil {
		return nil, fmt.Errorf("not an agentsmd project; run agentsmd init: %w", err)
	}
	return &Project{Root: root}, nil
}

func (p *Project) StateDir() string        { return filepath.Join(p.Root, DirName) }
func (p *Project) ArtifactPath() string    { return filepath.Join(p.Root, Artifact) }
func (p *Project) LedgerPath() string      { return filepath.Join(p.StateDir(), "ledger.json") }
func (p *Project) VersionsDir() string     { return filepath.Join(p.StateDir(), "versions") }
func (p *Project) PendingDir() string      { return filepath.Join(p.StateDir(), "pending") }
func (p *Project) RunsDir() string         { return filepath.Join(p.StateDir(), "runs") }
func (p *Project) ConnectionsPath() string { return filepath.Join(p.StateDir(), "connections.json") }

func (p *Project) Scaffold() error {
	for _, path := range []string{p.StateDir(), p.VersionsDir(), p.PendingDir(), p.RunsDir()} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
	}
	config := filepath.Join(p.StateDir(), "config.yaml")
	if _, err := os.Stat(config); errors.Is(err, os.ErrNotExist) {
		return AtomicWrite(config, []byte("version: 1\nadapters: []\n"), 0o644)
	}
	return nil
}

func (p *Project) LoadLedger() (schema.Ledger, error) {
	data, err := os.ReadFile(p.LedgerPath())
	if errors.Is(err, os.ErrNotExist) {
		return schema.Ledger{Rules: []schema.Rule{}, Runs: map[string][]int{}}, nil
	}
	if err != nil {
		return schema.Ledger{}, err
	}
	var value schema.Ledger
	if err := json.Unmarshal(data, &value); err != nil {
		return schema.Ledger{}, fmt.Errorf("decode ledger: %w", err)
	}
	if value.Runs == nil {
		value.Runs = map[string][]int{}
	}
	return value, nil
}

func (p *Project) SaveLedger(value schema.Ledger) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return AtomicWrite(p.LedgerPath(), append(data, '\n'), 0o644)
}

func AtomicWrite(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".agentsmd-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
