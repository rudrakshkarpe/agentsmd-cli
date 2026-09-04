package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/capture/claude"
	"github.com/rudrakshkarpe/agentsmd-cli/integration"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/spf13/cobra"
)

func (a *app) connectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "connect <codex|claude|goose|cursor>",
		Short: "Connect a coding CLI to the learning loop",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			record, err := integration.Connect(p, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "connected %s\n%s\n", record.Provider, filepath.Join(p.Root, record.Path))
			if record.Provider == "codex" {
				fmt.Fprintln(cmd.OutOrStdout(), "Review and trust the project hook with /hooks in Codex.")
			}
			return nil
		},
	}
}

func (a *app) hookCommand() *cobra.Command {
	command := &cobra.Command{
		Use:    "hook <provider>",
		Short:  "Receive a connected CLI lifecycle event",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), 8<<20))
			if err != nil {
				return err
			}
			return captureHook(a.root, strings.ToLower(args[0]), data)
		},
	}
	command.SetIn(os.Stdin)
	return command
}

func captureHook(root, provider string, data []byte) error {
	if !contains(integration.Supported, provider) {
		return fmt.Errorf("unsupported hook provider %q", provider)
	}
	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("decode %s hook event: %w", provider, err)
	}
	start := root
	if cwd, ok := event["cwd"].(string); ok && cwd != "" {
		start = cwd
	}
	p, err := project.Require(start)
	if err != nil {
		return err
	}
	sessionID := firstString(event, "session_id", "conversation_id", "generation_id")
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s-%d", provider, time.Now().UTC().UnixNano())
	}
	trajectory := &schema.Trajectory{
		SessionID: sessionID,
		Tool:      provider,
		Steps:     []schema.Step{}, ToolCalls: []schema.ToolCall{}, Files: []schema.FileTouch{}, Commands: []schema.Command{},
		Metadata: map[string]string{"hook_event": firstString(event, "hook_event_name", "event")},
	}
	if model := firstString(event, "model"); model != "" {
		trajectory.Metadata["model"] = model
	}
	if provider == "claude" {
		claudeEvent := claude.HookEvent{SessionID: sessionID, TranscriptPath: firstString(event, "transcript_path")}
		if claudeEvent.TranscriptPath != "" {
			if normalized, normalizeErr := (claude.Adapter{Event: claudeEvent}).Latest(context.Background()); normalizeErr == nil {
				trajectory = normalized
			}
		}
	}
	output, err := json.MarshalIndent(trajectory, "", "  ")
	if err != nil {
		return err
	}
	name := safeName(sessionID)
	if name == "" {
		name = fmt.Sprintf("%s-%d", provider, time.Now().UTC().UnixNano())
	}
	return project.AtomicWrite(filepath.Join(p.RunsDir(), name+".json"), append(output, '\n'), 0o644)
}

var unsafeFilename = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func safeName(value string) string {
	return strings.Trim(unsafeFilename.ReplaceAllString(value, "-"), "-.")
}

func firstString(value map[string]any, keys ...string) string {
	for _, key := range keys {
		if text, ok := value[key].(string); ok && text != "" {
			return text
		}
	}
	return ""
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
