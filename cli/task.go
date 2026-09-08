package cli

import (
	"fmt"
	"time"

	logicaltask "github.com/rudrakshkarpe/agentsmd-cli/task"
	"github.com/spf13/cobra"
)

func (a *app) taskCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "task",
		Short: "Group related agent sessions under one task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			active, err := logicaltask.Load(p)
			if logicaltask.IsInactive(err) {
				writeInfo(cmd, "no explicit task is active; feature branches are correlated automatically")
				return nil
			}
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", active.ID, active.Label, active.StartedAt.Format(time.RFC3339))
			return nil
		},
	}
	var id string
	start := &cobra.Command{
		Use:   "start LABEL",
		Short: "Start or resume a logical task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			active, err := logicaltask.Start(p, id, args[0], time.Now().UTC())
			if err != nil {
				return err
			}
			writeSuccess(cmd, "active task "+active.ID)
			return nil
		},
	}
	start.Flags().StringVar(&id, "id", "", "stable task id (defaults to a label slug)")
	command.AddCommand(start, &cobra.Command{
		Use:   "end",
		Short: "Stop assigning new sessions to the active task",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			active, err := logicaltask.End(p)
			if logicaltask.IsInactive(err) {
				return fmt.Errorf("no explicit task is active")
			}
			if err != nil {
				return err
			}
			writeSuccess(cmd, "ended task "+active.ID)
			return nil
		},
	})
	return command
}
