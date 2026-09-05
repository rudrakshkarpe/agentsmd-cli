// Package session captures cross-provider lifecycle and repository evidence.
package session

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

type State struct {
	Provider  string    `json:"provider"`
	SessionID string    `json:"session_id"`
	StartedAt time.Time `json:"started_at"`
	GitHead   string    `json:"git_head,omitempty"`
	GitStatus string    `json:"git_status,omitempty"`
	GitDiff   string    `json:"git_diff,omitempty"`
}

func IsStart(event map[string]any) bool {
	name := strings.ToLower(firstString(event, "hook_event_name", "event"))
	return strings.Contains(name, "start")
}

func Start(p *project.Project, provider, sessionID string, now time.Time) error {
	state := State{Provider: provider, SessionID: sessionID, StartedAt: now, GitHead: gitOutput(p.Root, "rev-parse", "HEAD"), GitStatus: filterStatus(gitOutput(p.Root, "status", "--porcelain=v1")), GitDiff: gitOutput(p.Root, "diff", "--no-ext-diff", "--no-color", "HEAD")}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return project.AtomicWrite(StatePath(p, provider, sessionID), append(data, '\n'), 0o644)
}

func Complete(p *project.Project, trajectory *schema.Trajectory, provider string, now time.Time) error {
	state, err := load(p, provider, trajectory.SessionID)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if trajectory.Metadata == nil {
		trajectory.Metadata = map[string]string{}
	}
	trajectory.Metadata["ended_at"] = now.Format(time.RFC3339Nano)
	if err == nil {
		trajectory.WallTimeS = now.Sub(state.StartedAt).Seconds()
		trajectory.Metadata["started_at"] = state.StartedAt.Format(time.RFC3339Nano)
		trajectory.Metadata["git_before"] = state.GitHead
		trajectory.Metadata["status_before"] = state.GitStatus
		if state.GitStatus != "" {
			trajectory.Metadata["baseline_dirty"] = "true"
		}
	}
	headAfter := gitOutput(p.Root, "rev-parse", "HEAD")
	trajectory.Metadata["git_after"] = headAfter
	status := filterStatus(gitOutput(p.Root, "status", "--porcelain=v1"))
	trajectory.Metadata["status_after"] = status
	workingDiff := gitOutput(p.Root, "diff", "--no-ext-diff", "--no-color", "HEAD")
	if err == nil && state.GitHead != "" && headAfter != "" && state.GitHead != headAfter {
		committedDiff := gitOutput(p.Root, "diff", "--no-ext-diff", "--no-color", state.GitHead, headAfter)
		trajectory.FinalDiff = strings.TrimSpace(committedDiff + "\n" + workingDiff)
	} else {
		if err == nil && state.GitDiff == workingDiff {
			trajectory.FinalDiff = ""
		} else {
			trajectory.FinalDiff = workingDiff
		}
	}
	if err == nil {
		trajectory.Files = fileTouches(statusDifference(state.GitStatus, status))
	} else {
		trajectory.Files = fileTouches(status)
	}
	return nil
}

func StatePath(p *project.Project, provider, sessionID string) string {
	return filepath.Join(p.SessionsDir(), provider+"-"+SafeName(sessionID)+".json")
}

func RunPath(p *project.Project, provider, sessionID string) string {
	return filepath.Join(p.RunsDir(), provider+"-"+SafeName(sessionID)+".json")
}

func SafeName(value string) string {
	var output strings.Builder
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("._-", char) {
			output.WriteRune(char)
		} else {
			output.WriteByte('-')
		}
	}
	return strings.Trim(output.String(), "-.")
}

func load(p *project.Project, provider, sessionID string) (State, error) {
	data, err := os.ReadFile(StatePath(p, provider, sessionID))
	if err != nil {
		return State{}, err
	}
	var value State
	return value, json.Unmarshal(data, &value)
}

func gitOutput(root string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	data, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(data), "\r\n")
}

func fileTouches(status string) []schema.FileTouch {
	result := []schema.FileTouch{}
	for _, line := range strings.Split(status, "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if index := strings.LastIndex(path, " -> "); index >= 0 {
			path = path[index+4:]
		}
		if path == project.DirName || strings.HasPrefix(path, project.DirName+"/") {
			continue
		}
		result = append(result, schema.FileTouch{Path: path})
	}
	return result
}

func filterStatus(status string) string {
	lines := []string{}
	for _, line := range strings.Split(status, "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if path == project.DirName || strings.HasPrefix(path, project.DirName+"/") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func statusDifference(before, after string) string {
	baseline := map[string]bool{}
	for _, line := range strings.Split(before, "\n") {
		baseline[line] = true
	}
	changed := []string{}
	for _, line := range strings.Split(after, "\n") {
		if line != "" && !baseline[line] {
			changed = append(changed, line)
		}
	}
	return strings.Join(changed, "\n")
}

func firstString(value map[string]any, keys ...string) string {
	for _, key := range keys {
		if text, ok := value[key].(string); ok && text != "" {
			return text
		}
	}
	return ""
}
