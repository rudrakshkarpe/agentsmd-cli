package cli

import (
	"fmt"
	"strconv"

	"github.com/rudrakshkarpe/agentsmd-cli/version"
	"github.com/spf13/cobra"
)

func (a *app) logCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "Show typed version history",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			items, err := version.New(p).Log()
			if err != nil {
				return err
			}
			for _, item := range items {
				extra := ""
				if delta, ok := item.Meta["token_delta"]; ok {
					extra = " tokens=" + fmt.Sprint(delta)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s [%s] %s%s\n", item.ID, item.Time.Format("2006-01-02T15:04:05Z"), item.Reason, item.Message, extra)
			}
			return nil
		},
	}
}

func (a *app) diffCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "diff [FROM] [TO]",
		Args:  cobra.MaximumNArgs(2),
		Short: "Compare two versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			store := version.New(p)
			if len(args) == 0 {
				items, err := store.Log()
				if err != nil {
					return err
				}
				if len(items) < 2 {
					return fmt.Errorf("need two versions to diff")
				}
				args = []string{items[len(items)-2].ID, items[len(items)-1].ID}
			} else if len(args) == 1 {
				return fmt.Errorf("provide both FROM and TO, or neither")
			}
			result, err := store.Diff(args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), result)
			return nil
		},
	}
}

func (a *app) commitCommand() *cobra.Command {
	var message string
	command := &cobra.Command{
		Use:   "commit",
		Short: "Snapshot the current AGENTS.md",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			item, err := version.New(p).Commit(message, "manual", map[string]any{})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), item.ID)
			return nil
		},
	}
	command.Flags().StringVarP(&message, "message", "m", "", "reason for the snapshot")
	_ = command.MarkFlagRequired("message")
	return command
}

func (a *app) revertCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "revert VERSION",
		Args:  cobra.ExactArgs(1),
		Short: "Restore a version as a new snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			item, err := version.New(p).Revert(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), item.ID)
			return nil
		},
	}
}

func (a *app) tagCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tag VERSION NAME",
		Args:  cobra.ExactArgs(2),
		Short: "Give a version a stable name",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			if err := version.New(p).Tag(args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s\n", args[1], args[0])
			return nil
		},
	}
}

func (a *app) blameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "blame",
		Short: "Show per-rule provenance",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			value, err := p.LoadLedger()
			if err != nil {
				return err
			}
			for _, rule := range value.Rules {
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] run=%s task=%s cited=%s %s\n", rule.ID, rule.Origin.Run, rule.Origin.Task, strconv.Itoa(rule.Cited), rule.Text)
			}
			return nil
		},
	}
}
