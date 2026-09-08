package cli

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/spf13/cobra"
)

func (a *app) progressCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "progress [task]",
		Short: "Compare outcomes across sessions of a logical task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			items, err := loadSessions(p.RunsDir())
			if err != nil {
				return err
			}
			grouped := map[string][]sessionSummary{}
			for _, item := range items {
				if item.Trajectory.Task != "" {
					grouped[item.Trajectory.Task] = append(grouped[item.Trajectory.Task], item)
				}
			}
			if len(args) == 1 {
				group := grouped[args[0]]
				if len(group) == 0 {
					return fmt.Errorf("no sessions recorded for task %q", args[0])
				}
				writeTaskProgress(cmd, args[0], group, true)
				return nil
			}
			if len(grouped) == 0 {
				writeInfo(cmd, "no task-correlated sessions")
				return nil
			}
			keys := make([]string, 0, len(grouped))
			for key := range grouped {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				writeTaskProgress(cmd, key, grouped[key], false)
			}
			return nil
		},
	}
}

func writeTaskProgress(cmd *cobra.Command, taskID string, items []sessionSummary, detailed bool) {
	sort.SliceStable(items, func(i, j int) bool {
		return sessionTime(items[i]).Before(sessionTime(items[j]))
	})
	successes := 0
	for _, item := range items {
		if sessionOutcome(item.Trajectory) == "success" {
			successes++
		}
	}
	first, latest := items[0].Trajectory, items[len(items)-1].Trajectory
	fmt.Fprintf(cmd.OutOrStdout(), "%s\truns=%d\tsuccess=%d/%d\tduration=%s\ttokens=%s\n", taskID, len(items), successes, len(items), comparison(first.WallTimeS, latest.WallTimeS, "%.1fs"), comparison(float64(first.Tokens.Total()), float64(latest.Tokens.Total()), "%.0f"))
	if !detailed {
		return
	}
	for _, item := range items {
		trajectory := item.Trajectory
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\t%s\t%.1fs\ttokens=%d\tfiles=%d\ttests=%d/%d\n", item.ID, sessionOutcome(trajectory), trajectory.WallTimeS, trajectory.Tokens.Total(), len(trajectory.Files), trajectory.TestResults.Passed, trajectory.TestResults.Failed)
	}
}

func comparison(first, latest float64, format string) string {
	if first == 0 && latest == 0 {
		return "n/a"
	}
	if first == latest {
		return fmt.Sprintf(format, latest)
	}
	return fmt.Sprintf(format+"→"+format+" (%+.1f%%)", first, latest, percentChange(first, latest))
}

func percentChange(first, latest float64) float64 {
	if first == 0 {
		return 0
	}
	return (latest - first) / first * 100
}

func sessionOutcome(trajectory schema.Trajectory) string {
	if outcome := trajectory.Metadata["outcome"]; outcome != "" {
		return outcome
	}
	if outcome := trajectory.Metadata["provider_status"]; outcome != "" {
		return outcome
	}
	return "captured"
}

func sessionTime(item sessionSummary) time.Time {
	for _, key := range []string{"ended_at", "started_at"} {
		if value, err := time.Parse(time.RFC3339Nano, item.Trajectory.Metadata[key]); err == nil {
			return value
		}
	}
	// Retain deterministic filename ordering for legacy trajectories.
	seconds, _ := strconv.ParseInt(item.ID, 10, 64)
	return time.Unix(seconds, 0)
}
