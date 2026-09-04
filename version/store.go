// Package version stores typed AGENTS.md snapshots and provenance.
package version

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

const indexName = "versions.jsonl"

type Store struct {
	Project *project.Project
	Now     func() time.Time
}

func New(p *project.Project) *Store {
	return &Store{Project: p, Now: func() time.Time { return time.Now().UTC() }}
}

func (s *Store) Log() ([]schema.Version, error) {
	file, err := os.Open(filepath.Join(s.Project.VersionsDir(), indexName))
	if errors.Is(err, os.ErrNotExist) {
		return []schema.Version{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := []schema.Version{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var item schema.Version
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			return nil, fmt.Errorf("decode version index: %w", err)
		}
		result = append(result, item)
	}
	return result, scanner.Err()
}

func (s *Store) Commit(message, reason string, meta map[string]any) (schema.Version, error) {
	if strings.TrimSpace(message) == "" {
		return schema.Version{}, fmt.Errorf("commit message cannot be empty")
	}
	items, err := s.Log()
	if err != nil {
		return schema.Version{}, err
	}
	id := fmt.Sprintf("v%04d", len(items))
	var parent *string
	if len(items) > 0 {
		value := items[len(items)-1].ID
		parent = &value
	}
	content, err := os.ReadFile(s.Project.ArtifactPath())
	if errors.Is(err, os.ErrNotExist) {
		content = []byte{}
	} else if err != nil {
		return schema.Version{}, err
	}
	if err := project.AtomicWrite(filepath.Join(s.Project.VersionsDir(), id+".md"), content, 0o644); err != nil {
		return schema.Version{}, err
	}
	item := schema.Version{ID: id, Parent: parent, Time: s.Now(), Reason: reason, Message: message, Meta: meta}
	data, err := json.Marshal(item)
	if err != nil {
		return schema.Version{}, err
	}
	index, err := os.OpenFile(filepath.Join(s.Project.VersionsDir(), indexName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return schema.Version{}, err
	}
	defer index.Close()
	if _, err := index.Write(append(data, '\n')); err != nil {
		return schema.Version{}, err
	}
	return item, index.Sync()
}

func (s *Store) Content(id string) ([]byte, error) {
	if !validID(id) {
		return nil, fmt.Errorf("invalid version id %q", id)
	}
	data, err := os.ReadFile(filepath.Join(s.Project.VersionsDir(), id+".md"))
	if err != nil {
		return nil, fmt.Errorf("read version %s: %w", id, err)
	}
	return data, nil
}

func (s *Store) Revert(id string) (schema.Version, error) {
	data, err := s.Content(id)
	if err != nil {
		return schema.Version{}, err
	}
	if err := project.AtomicWrite(s.Project.ArtifactPath(), data, 0o644); err != nil {
		return schema.Version{}, err
	}
	return s.Commit("revert to "+id, "manual", map[string]any{"reverted_to": id})
}

func (s *Store) Tag(id, name string) error {
	if _, err := s.Content(id); err != nil {
		return err
	}
	if name == "" || strings.ContainsAny(name, `/\\`) {
		return fmt.Errorf("invalid tag name %q", name)
	}
	path := filepath.Join(s.Project.VersionsDir(), "tags.json")
	tags := map[string]string{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &tags); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	tags[name] = id
	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return err
	}
	return project.AtomicWrite(path, append(data, '\n'), 0o644)
}

func validID(id string) bool {
	if len(id) != 5 || id[0] != 'v' {
		return false
	}
	for _, value := range id[1:] {
		if value < '0' || value > '9' {
			return false
		}
	}
	return true
}
