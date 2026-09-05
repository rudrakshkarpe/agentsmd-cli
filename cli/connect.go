package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/automation"
	"github.com/rudrakshkarpe/agentsmd-cli/capture/claude"
	"github.com/rudrakshkarpe/agentsmd-cli/integration"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/rudrakshkarpe/agentsmd-cli/session"
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
			provider := record.Provider
			if uiFor(cmd).interactive {
				provider = displayProvider(record.Provider)
			}
			writeSuccess(cmd, fmt.Sprintf("connected %s", provider))
			writeInfo(cmd, filepath.Join(p.Root, record.Path))
			if record.Provider == "codex" {
				writeWarning(cmd, "Review and trust the project hook with /hooks in Codex.")
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
			return receiveHook(a.root, strings.ToLower(args[0]), data)
		},
	}
	command.SetIn(os.Stdin)
	return command
}

func (a *app) ingestCommand() *cobra.Command {
	var provider, eventPath string
	command := &cobra.Command{
		Use:    "ingest",
		Short:  "Normalize a queued lifecycle event",
		Hidden: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			data, err := os.ReadFile(eventPath)
			if err != nil {
				return err
			}
			return captureHook(a.root, provider, data)
		},
	}
	command.Flags().StringVar(&provider, "provider", "", "hook provider")
	command.Flags().StringVar(&eventPath, "event", "", "queued event path")
	_ = command.MarkFlagRequired("provider")
	_ = command.MarkFlagRequired("event")
	return command
}

func receiveHook(root, provider string, data []byte) error {
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
	if session.IsStart(event) {
		return captureHook(root, provider, data)
	}
	sessionID := firstString(event, "session_id", "conversation_id", "generation_id")
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s-%d", provider, time.Now().UTC().UnixNano())
	}
	eventPath := filepath.Join(p.InboxDir(), provider+"-"+session.SafeName(sessionID)+".json")
	if err := project.AtomicWrite(eventPath, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return launchDetached(p, "--root", p.Root, "ingest", "--provider", provider, "--event", eventPath)
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
	if session.IsStart(event) {
		return session.Start(p, provider, sessionID, time.Now().UTC())
	}
	trajectory := &schema.Trajectory{
		SessionID: sessionID,
		Tool:      provider,
		Task:      firstString(event, "task", "task_id"),
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
	if trajectory.Metadata == nil {
		trajectory.Metadata = map[string]string{}
	}
	trajectory.Metadata["hook_event"] = firstString(event, "hook_event_name", "event")
	if model := firstString(event, "model"); model != "" {
		trajectory.Metadata["model"] = model
	}
	if status := firstString(event, "status", "reason"); status != "" {
		trajectory.Metadata["provider_status"] = status
	}
	if err := session.Complete(p, trajectory, provider, time.Now().UTC()); err != nil {
		return err
	}
	output, err := json.MarshalIndent(trajectory, "", "  ")
	if err != nil {
		return err
	}
	path := session.RunPath(p, provider, sessionID)
	if err := project.AtomicWrite(path, append(output, '\n'), 0o644); err != nil {
		return err
	}
	config, err := automation.Load(p)
	if err != nil || len(config.ReflectCommand) == 0 {
		return err
	}
	jobPath, err := automation.Enqueue(p, path)
	if err != nil {
		return err
	}
	return launchProcessor(p, jobPath)
}

func launchProcessor(p *project.Project, jobPath string) error {
	return launchDetached(p, "--root", p.Root, "process", "--job", jobPath)
}

func launchDetached(p *project.Project, args ...string) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	log, err := os.OpenFile(filepath.Join(p.StateDir(), "automation.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	command := exec.Command(executable, args...)
	configureDetached(command)
	command.Stdin = nil
	command.Stdout, command.Stderr = log, log
	if err := command.Start(); err != nil {
		log.Close()
		return err
	}
	log.Close()
	return command.Process.Release()
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
