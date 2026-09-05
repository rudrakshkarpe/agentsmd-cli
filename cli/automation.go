package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rudrakshkarpe/agentsmd-cli/automation"
	"github.com/spf13/cobra"
)

func (a *app) automateCommand() *cobra.Command {
	var reflectCommand, evaluateCommand string
	var autoPromote bool
	var minConfidence float64
	command := &cobra.Command{
		Use:   "automate",
		Short: "Configure automatic reflection and evaluation",
		Long:  "Configure processing after a captured session. Reflection creates a pending proposal. Automatic promotion is opt-in and requires a successful evaluation command.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			config, err := automation.Load(p)
			if err != nil {
				return err
			}
			changed := false
			if cmd.Flags().Changed("reflect-command") {
				config.ReflectCommand, changed = strings.Fields(reflectCommand), true
			}
			if cmd.Flags().Changed("evaluate-command") {
				config.EvaluateCommand, changed = strings.Fields(evaluateCommand), true
			}
			if cmd.Flags().Changed("auto-promote") {
				config.AutoPromote, changed = autoPromote, true
			}
			if cmd.Flags().Changed("min-confidence") {
				config.MinConfidence, changed = minConfidence, true
			}
			if changed {
				if err := automation.Save(p, config); err != nil {
					return err
				}
			}
			ui := uiFor(cmd)
			if ui.interactive {
				fmt.Fprintln(cmd.OutOrStdout(), ui.icon("🧠")+ui.brand("agentsmd CLI · automation"))
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", ui.accent("Reflection "), ui.muted(displayCommand(config.ReflectCommand)))
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", ui.accent("Evaluation "), ui.muted(displayCommand(config.EvaluateCommand)))
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", ui.accent("Promotion  "), ui.muted(map[bool]string{true: "automatic", false: "manual"}[config.AutoPromote]))
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", ui.accent("Confidence "), ui.muted(fmt.Sprintf("%.2f", config.MinConfidence)))
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "reflection: %s\nevaluation: %s\nauto-promote: %t\nminimum confidence: %.2f\n", displayCommand(config.ReflectCommand), displayCommand(config.EvaluateCommand), config.AutoPromote, config.MinConfidence)
			return nil
		},
	}
	command.Flags().StringVar(&reflectCommand, "reflect-command", "", "program that maps trajectory JSON to a reflection verdict")
	command.Flags().StringVar(&evaluateCommand, "evaluate-command", "", "command that must exit successfully before automatic promotion")
	command.Flags().BoolVar(&autoPromote, "auto-promote", false, "promote proposals that pass evaluation and confidence policy")
	command.Flags().Float64Var(&minConfidence, "min-confidence", 0.8, "minimum reflection confidence for automatic promotion")
	return command
}

func (a *app) processCommand() *cobra.Command {
	var jobPath string
	command := &cobra.Command{
		Use:    "process",
		Short:  "Process a queued reflection job",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			job, err := automation.ProcessJob(context.Background(), p, jobPath)
			if err != nil {
				return err
			}
			data, _ := json.Marshal(job.Result)
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
	command.Flags().StringVar(&jobPath, "job", "", "queued job path")
	_ = command.MarkFlagRequired("job")
	return command
}

func displayCommand(argv []string) string {
	if len(argv) == 0 {
		return "disabled"
	}
	return strings.Join(argv, " ")
}
