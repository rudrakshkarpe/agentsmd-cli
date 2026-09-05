package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/spf13/cobra"
)

type sessionSummary struct {
	ID         string
	Trajectory schema.Trajectory
}

func (a *app) sessionsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "sessions",
		Short: "Inspect captured agent sessions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			items, err := loadSessions(p.RunsDir())
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no captured sessions")
				return nil
			}
			for _, item := range items {
				outcome := item.Trajectory.Metadata["outcome"]
				if outcome == "" {
					outcome = item.Trajectory.Metadata["provider_status"]
				}
				if outcome == "" {
					outcome = "captured"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-28s %-8s %-10s %6.1fs files=%d tests=%d/%d\n", item.ID, item.Trajectory.Tool, outcome, item.Trajectory.WallTimeS, len(item.Trajectory.Files), item.Trajectory.TestResults.Passed, item.Trajectory.TestResults.Failed)
			}
			return nil
		},
	}
	command.AddCommand(&cobra.Command{
		Use:   "show RUN",
		Short: "Show the complete normalized trajectory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if filepath.Base(args[0]) != args[0] || strings.Contains(args[0], "..") {
				return fmt.Errorf("invalid run id %q", args[0])
			}
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			data, err := os.ReadFile(filepath.Join(p.RunsDir(), strings.TrimSuffix(args[0], ".json")+".json"))
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		},
	})
	return command
}

func loadSessions(dir string) ([]sessionSummary, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []sessionSummary{}, nil
	}
	if err != nil {
		return nil, err
	}
	result := []sessionSummary{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var trajectory schema.Trajectory
		if err := json.Unmarshal(data, &trajectory); err != nil {
			return nil, fmt.Errorf("decode %s: %w", entry.Name(), err)
		}
		result = append(result, sessionSummary{ID: strings.TrimSuffix(entry.Name(), ".json"), Trajectory: trajectory})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
