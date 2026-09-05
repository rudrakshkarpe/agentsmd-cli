package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/learning"
	"github.com/rudrakshkarpe/agentsmd-cli/ledger"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	reflector "github.com/rudrakshkarpe/agentsmd-cli/reflect"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/spf13/cobra"
)

func (a *app) statusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show pending learning proposals",
		RunE:  func(cmd *cobra.Command, _ []string) error { return a.printPending(cmd) },
	}
}

func (a *app) pendingCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pending",
		Short: "List rules awaiting review",
		RunE:  func(cmd *cobra.Command, _ []string) error { return a.printPending(cmd) },
	}
}

func (a *app) printPending(cmd *cobra.Command) error {
	p, err := a.requireProject()
	if err != nil {
		return err
	}
	items, err := learning.New(p).Pending()
	if err != nil {
		return err
	}
	for _, item := range items {
		ui := uiFor(cmd)
		fmt.Fprintf(cmd.OutOrStdout(), "%s%s %s %s\n", ui.icon("⏳"), ui.accent(item.ID), ui.muted("task="+item.Origin.Task), ui.soft(item.Text))
	}
	if len(items) == 0 {
		writeSuccess(cmd, "no pending rules")
	}
	return nil
}

func (a *app) learnCommand() *cobra.Command {
	var rule, run, task, trajectoryPath, reflectCommand string
	command := &cobra.Command{
		Use:   "learn",
		Short: "Propose one targeted rule at a task boundary",
		Long:  "Propose one targeted rule for review. Use --rule for deterministic manual input, or --trajectory with --reflect-command to run a provider that accepts trajectory JSON and returns a reflection verdict.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			service := learning.New(p)
			if rule != "" {
				proposal, err := service.Propose(rule, schema.Origin{Run: run, Task: task})
				if err != nil {
					return err
				}
				writeSuccess(cmd, fmt.Sprintf("proposed %s", proposal.ID))
				return nil
			}
			if trajectoryPath == "" || reflectCommand == "" {
				return fmt.Errorf("provide --rule, or both --trajectory and --reflect-command")
			}
			data, err := os.ReadFile(trajectoryPath)
			if err != nil {
				return err
			}
			var trajectory schema.Trajectory
			if err := json.Unmarshal(data, &trajectory); err != nil {
				return fmt.Errorf("decode trajectory: %w", err)
			}
			proposal, result, err := service.Learn(context.Background(), trajectory, reflector.Command{Argv: strings.Fields(reflectCommand), Timeout: 2 * time.Minute})
			if err != nil {
				return err
			}
			if proposal == nil {
				writeInfo(cmd, fmt.Sprintf("no proposal: %s", result.Verdict))
				return nil
			}
			writeSuccess(cmd, fmt.Sprintf("proposed %s", proposal.ID))
			return nil
		},
	}
	command.Flags().StringVar(&rule, "rule", "", "imperative repository-specific rule")
	command.Flags().StringVar(&run, "run", "", "source session identifier")
	command.Flags().StringVar(&task, "task", "", "source task identifier")
	command.Flags().StringVar(&trajectoryPath, "trajectory", "", "normalized trajectory JSON")
	command.Flags().StringVar(&reflectCommand, "reflect-command", "", "program that maps trajectory JSON to a verdict")
	return command
}

func (a *app) promoteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "promote PROPOSAL",
		Args:  cobra.ExactArgs(1),
		Short: "Promote an approved pending rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			rule, duplicate, err := learning.New(p).Promote(args[0])
			if err != nil {
				return err
			}
			if duplicate != nil {
				writeInfo(cmd, fmt.Sprintf("already covered by %s", duplicate.ID))
				return nil
			}
			writeSuccess(cmd, fmt.Sprintf("promoted %s", rule.ID))
			return nil
		},
	}
}

func (a *app) rejectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reject PROPOSAL",
		Args:  cobra.ExactArgs(1),
		Short: "Discard a pending rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			if err := learning.New(p).Reject(args[0]); err != nil {
				return err
			}
			writeSuccess(cmd, fmt.Sprintf("rejected %s", args[0]))
			return nil
		},
	}
}

func (a *app) savingsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "savings TASK",
		Args:  cobra.ExactArgs(1),
		Short: "Report token change across runs of one task",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			result, err := learning.New(p).Savings(args[0])
			if err != nil {
				return err
			}
			if result == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "task %s needs at least two runs\n", args[0])
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %d -> %d tokens (%+.1f%% over %d runs)\n", result.Task, result.First, result.Last, result.Percent, result.Runs)
			return nil
		},
	}
}

func (a *app) recordCommand() *cobra.Command {
	var task string
	var tokens int
	command := &cobra.Command{
		Use:   "record",
		Short: "Record token usage for a repeatable task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			if err := learning.New(p).Record(task, tokens); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "recorded %d tokens for %s\n", tokens, task)
			return nil
		},
	}
	command.Flags().StringVar(&task, "task", "", "stable task identifier")
	command.Flags().IntVar(&tokens, "tokens", 0, "input plus output tokens")
	_ = command.MarkFlagRequired("task")
	_ = command.MarkFlagRequired("tokens")
	return command
}

func (a *app) pruneCommand() *cobra.Command {
	var apply bool
	command := &cobra.Command{
		Use:   "prune",
		Short: "Retire unused rules after sufficient evidence",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			value, err := p.LoadLedger()
			if err != nil {
				return err
			}
			tasksSeen := 0
			for _, runs := range value.Runs {
				tasksSeen += len(runs)
			}
			if tasksSeen < 20 {
				fmt.Fprintf(cmd.OutOrStdout(), "not enough evidence: %d/20 runs\n", tasksSeen)
				return nil
			}
			kept := value.Rules[:0]
			removed := []string{}
			for _, rule := range value.Rules {
				if rule.Cited == 0 {
					removed = append(removed, rule.ID)
					continue
				}
				kept = append(kept, rule)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "candidates: %v\n", removed)
			if !apply || len(removed) == 0 {
				return nil
			}
			value.Rules = kept
			if err := p.SaveLedger(value); err != nil {
				return err
			}
			return project.AtomicWrite(p.ArtifactPath(), []byte(ledger.Render(value)), 0o644)
		},
	}
	command.Flags().BoolVar(&apply, "apply", false, "remove the listed rules")
	return command
}
