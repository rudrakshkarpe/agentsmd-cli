package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/benchmark"
	"github.com/spf13/cobra"
)

func (a *app) benchmarkCommand() *cobra.Command {
	var specPath, suitePath, agentCommand, outputDir string
	var trials int
	var timeout time.Duration
	command := &cobra.Command{
		Use:   "benchmark",
		Short: "Compare baseline and learned AGENTS.md guidance",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if suitePath != "" {
				suite, err := benchmark.LoadSuite(suitePath)
				if err != nil {
					return err
				}
				specs, err := benchmark.SuiteSpecs(suitePath, suite)
				if err != nil {
					return err
				}
				agentArgs := strings.Fields(agentCommand)
				if len(agentArgs) == 0 {
					agentArgs, err = commonBenchmarkAgent(specs)
					if err != nil {
						return err
					}
				}
				runner := benchmark.Runner{AgentCommand: agentArgs, Seeds: trials, OutputDir: outputDir, Timeout: timeout}
				writeSection(cmd, "🧪", "agentsmd CLI · benchmark suite")
				writeInfo(cmd, fmt.Sprintf("%s · %d tasks · %d trials per condition", suite.Name, len(specs), trials))
				report, err := runner.RunSuite(context.Background(), suitePath, suite)
				if err != nil {
					return err
				}
				for _, task := range report.Tasks {
					baseline := summarizeRuns(task.Report.Runs, "baseline")
					learned := summarizeRuns(task.Report.Runs, "learned")
					fmt.Fprintf(cmd.OutOrStdout(), "%s  baseline %d/%d · learned %d/%d\n", task.Report.Spec.Task, baseline.passed, baseline.total, learned.passed, learned.total)
				}
				if len(report.Regressions) > 0 {
					return fmt.Errorf("regression gate failed for: %s", strings.Join(report.Regressions, ", "))
				}
				writeSuccess(cmd, fmt.Sprintf("suite report written to %s", runner.OutputDir))
				return nil
			}
			spec, err := benchmark.LoadSpec(specPath)
			if err != nil {
				return err
			}
			agentArgs := strings.Fields(agentCommand)
			if len(agentArgs) == 0 {
				agentArgs, err = defaultBenchmarkAgent(spec)
				if err != nil {
					return err
				}
			}
			runner := benchmark.Runner{AgentCommand: agentArgs, Seeds: trials, OutputDir: outputDir, Timeout: timeout}
			writeSection(cmd, "🧪", "agentsmd CLI · benchmark")
			writeInfo(cmd, fmt.Sprintf("%s · %d trials × %d conditions", spec.Name, trials, 2+len(spec.Ablations)))
			report, err := runner.Run(context.Background(), specPath, spec)
			if err != nil {
				return err
			}
			baseline, learned := summarizeRuns(report.Runs, "baseline"), summarizeRuns(report.Runs, "learned")
			fmt.Fprintf(cmd.OutOrStdout(), "baseline  %d/%d passed · median %d tokens\n", baseline.passed, baseline.total, baseline.medianTokens)
			fmt.Fprintf(cmd.OutOrStdout(), "learned   %d/%d passed · median %d tokens\n", learned.passed, learned.total, learned.medianTokens)
			writeSuccess(cmd, fmt.Sprintf("report written to %s", runner.OutputDir))
			return nil
		},
	}
	command.Flags().StringVar(&specPath, "spec", "benchmarks/config-precedence/spec.json", "benchmark specification")
	command.Flags().StringVar(&suitePath, "suite", "", "multi-task benchmark suite specification")
	command.Flags().StringVar(&agentCommand, "agent-command", "", "agent command that reads the task from stdin (defaults to the Codex model in the spec)")
	command.Flags().IntVar(&trials, "trials", 3, "independent trials per condition")
	command.Flags().StringVar(&outputDir, "output", "", "artifact directory (defaults beside the spec)")
	command.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "maximum time per agent trial")
	return command
}

func commonBenchmarkAgent(specs []benchmark.Spec) ([]string, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("benchmark suite has no task specs")
	}
	wanted, err := defaultBenchmarkAgent(specs[0])
	if err != nil {
		return nil, err
	}
	for _, spec := range specs[1:] {
		candidate, err := defaultBenchmarkAgent(spec)
		if err != nil {
			return nil, err
		}
		if strings.Join(candidate, "\x00") != strings.Join(wanted, "\x00") {
			return nil, fmt.Errorf("suite tasks must use one agent, model, and reasoning effort unless --agent-command is supplied")
		}
	}
	return wanted, nil
}

func defaultBenchmarkAgent(spec benchmark.Spec) ([]string, error) {
	if spec.Agent != "codex" {
		return nil, fmt.Errorf("--agent-command is required for agent %q", spec.Agent)
	}
	if spec.Model == "" || strings.ContainsAny(spec.Model, " \t\r\n") {
		return nil, fmt.Errorf("a valid model is required for the default Codex runner")
	}
	args := []string{"codex", "--ask-for-approval", "never", "exec", "-m", spec.Model}
	if spec.Reasoning != "" {
		if strings.ContainsAny(spec.Reasoning, " \t\r\n") {
			return nil, fmt.Errorf("reasoning_effort cannot contain whitespace")
		}
		args = append(args, "-c", "model_reasoning_effort="+spec.Reasoning)
	}
	return append(args, "--ignore-user-config", "--json", "--ephemeral", "--sandbox", "workspace-write", "-"), nil
}

type runSummary struct{ passed, total, medianTokens int }

func summarizeRuns(runs []benchmark.Run, condition string) runSummary {
	values := []int{}
	result := runSummary{}
	for _, run := range runs {
		if run.Condition != condition {
			continue
		}
		result.total++
		if run.Passed {
			result.passed++
		}
		values = append(values, run.TotalTokens())
	}
	sort.Ints(values)
	if len(values) > 0 {
		result.medianTokens = values[len(values)/2]
	}
	return result
}
