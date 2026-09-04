package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rudrakshkarpe/agentsmd-cli/capture/claude"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/spf13/cobra"
)

func (a *app) captureCommand() *cobra.Command {
	command := &cobra.Command{Use: "capture", Short: "Normalize an agent session"}
	var eventPath string
	claudeCommand := &cobra.Command{
		Use:   "claude",
		Short: "Capture a Claude Code hook event",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			data, err := os.ReadFile(eventPath)
			if err != nil {
				return err
			}
			var event claude.HookEvent
			if err := json.Unmarshal(data, &event); err != nil {
				return fmt.Errorf("decode hook event: %w", err)
			}
			if event.SessionID == "" || filepath.Base(event.SessionID) != event.SessionID {
				return fmt.Errorf("invalid Claude session_id %q", event.SessionID)
			}
			trajectory, err := (claude.Adapter{Event: event}).Latest(context.Background())
			if err != nil {
				return err
			}
			output, err := json.MarshalIndent(trajectory, "", "  ")
			if err != nil {
				return err
			}
			path := filepath.Join(p.RunsDir(), trajectory.SessionID+".json")
			if err := project.AtomicWrite(path, append(output, '\n'), 0o644); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	claudeCommand.Flags().StringVar(&eventPath, "event", "", "path to Claude hook event JSON")
	_ = claudeCommand.MarkFlagRequired("event")
	command.AddCommand(claudeCommand)
	return command
}
