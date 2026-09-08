// Package task assigns stable, project-owned identities to related agent runs.
package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

type Active struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	StartedAt time.Time `json:"started_at"`
}

type Identity struct {
	ID     string
	Source string
}

func Start(p *project.Project, id, label string, now time.Time) (Active, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return Active{}, fmt.Errorf("task label cannot be empty")
	}
	if id == "" {
		id = Slug(label)
	}
	if id = strings.TrimSpace(id); id == "" {
		return Active{}, fmt.Errorf("task id cannot be empty")
	}
	value := Active{ID: id, Label: label, StartedAt: now.UTC()}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return Active{}, err
	}
	return value, project.AtomicWrite(p.ActiveTaskPath(), append(data, '\n'), 0o644)
}

func Load(p *project.Project) (Active, error) {
	data, err := os.ReadFile(p.ActiveTaskPath())
	if err != nil {
		return Active{}, err
	}
	var value Active
	if err := json.Unmarshal(data, &value); err != nil {
		return Active{}, fmt.Errorf("decode active task: %w", err)
	}
	if value.ID == "" {
		return Active{}, fmt.Errorf("active task has no id")
	}
	return value, nil
}

func End(p *project.Project) (Active, error) {
	value, err := Load(p)
	if err != nil {
		return Active{}, err
	}
	return value, os.Remove(p.ActiveTaskPath())
}

// Resolve uses explicit provider data first, then project state, then a Git
// branch. It deliberately falls back to a run-scoped identity instead of
// falsely grouping unrelated work performed on a default branch.
func Resolve(p *project.Project, explicit, provider, sessionID string) Identity {
	if explicit = strings.TrimSpace(explicit); explicit != "" {
		return Identity{ID: explicit, Source: "provider"}
	}
	if value := strings.TrimSpace(os.Getenv("AGENTSMD_TASK_ID")); value != "" {
		return Identity{ID: value, Source: "environment"}
	}
	if active, err := Load(p); err == nil {
		return Identity{ID: active.ID, Source: "active-task"}
	}
	if branch := gitBranch(p.Root); branch != "" && branch != "main" && branch != "master" {
		return Identity{ID: "branch:" + branch, Source: "git-branch"}
	}
	return Identity{ID: "run:" + provider + ":" + sessionID, Source: "run"}
}

func Slug(value string) string {
	var result strings.Builder
	dash := false
	for _, char := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			if dash && result.Len() > 0 {
				result.WriteByte('-')
			}
			result.WriteRune(char)
			dash = false
			continue
		}
		dash = true
	}
	return strings.Trim(result.String(), "-")
}

func gitBranch(root string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	data, err := exec.CommandContext(ctx, "git", "-C", root, "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func IsInactive(err error) bool { return errors.Is(err, os.ErrNotExist) }
