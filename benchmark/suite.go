package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/session"
)

type Suite struct {
	Name  string   `json:"name"`
	Specs []string `json:"specs"`
}

type TaskReport struct {
	SpecPath string `json:"spec_path"`
	Report   Report `json:"report"`
}

type SuiteReport struct {
	SchemaVersion int          `json:"schema_version"`
	Suite         Suite        `json:"suite"`
	GeneratedAt   time.Time    `json:"generated_at"`
	Tasks         []TaskReport `json:"tasks"`
	Regressions   []string     `json:"regressions"`
}

func LoadSuite(path string) (Suite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Suite{}, err
	}
	var suite Suite
	if err := json.Unmarshal(data, &suite); err != nil {
		return Suite{}, fmt.Errorf("decode benchmark suite: %w", err)
	}
	if suite.Name == "" || len(suite.Specs) < 2 {
		return Suite{}, fmt.Errorf("benchmark suite requires a name and at least two task specs")
	}
	seen := map[string]bool{}
	for _, ref := range suite.Specs {
		if filepath.IsAbs(ref) || ref == "" {
			return Suite{}, fmt.Errorf("suite spec paths must be relative")
		}
		clean := filepath.Clean(ref)
		if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return Suite{}, fmt.Errorf("suite spec path escapes suite directory: %s", ref)
		}
		if seen[clean] {
			return Suite{}, fmt.Errorf("duplicate suite spec %q", ref)
		}
		seen[clean] = true
	}
	return suite, nil
}

func SuiteSpecs(suitePath string, suite Suite) ([]Spec, error) {
	result := make([]Spec, 0, len(suite.Specs))
	names := map[string]bool{}
	tasks := map[string]bool{}
	for _, ref := range suite.Specs {
		spec, err := LoadSpec(filepath.Join(filepath.Dir(suitePath), ref))
		if err != nil {
			return nil, fmt.Errorf("load suite spec %s: %w", ref, err)
		}
		if names[spec.Name] || tasks[spec.Task] {
			return nil, fmt.Errorf("suite task names and logical task ids must be unique: %s", ref)
		}
		names[spec.Name], tasks[spec.Task] = true, true
		result = append(result, spec)
	}
	return result, nil
}

func (r *Runner) RunSuite(ctx context.Context, suitePath string, suite Suite) (SuiteReport, error) {
	if len(r.AgentCommand) == 0 {
		return SuiteReport{}, fmt.Errorf("agent command cannot be empty")
	}
	root := r.OutputDir
	if root == "" {
		root = filepath.Join(filepath.Dir(suitePath), "results", time.Now().UTC().Format("20060102T150405Z"))
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return SuiteReport{}, err
	}
	report := SuiteReport{SchemaVersion: 1, Suite: suite, GeneratedAt: time.Now().UTC(), Tasks: []TaskReport{}, Regressions: []string{}}
	for _, ref := range suite.Specs {
		specPath := filepath.Join(filepath.Dir(suitePath), ref)
		spec, err := LoadSpec(specPath)
		if err != nil {
			return report, fmt.Errorf("load suite spec %s: %w", ref, err)
		}
		taskRunner := *r
		taskRunner.OutputDir = filepath.Join(root, session.SafeName(spec.Name))
		taskReport, err := taskRunner.Run(ctx, specPath, spec)
		if err != nil {
			return report, fmt.Errorf("run suite task %s: %w", spec.Name, err)
		}
		report.Tasks = append(report.Tasks, TaskReport{SpecPath: ref, Report: taskReport})
		if successRate(selectRuns(taskReport.Runs, "learned")) < successRate(selectRuns(taskReport.Runs, "baseline")) {
			report.Regressions = append(report.Regressions, spec.Task)
		}
		if err := saveSuiteReport(root, report); err != nil {
			return report, err
		}
	}
	r.OutputDir = root
	return report, saveSuiteReport(root, report)
}

func saveSuiteReport(dir string, report SuiteReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "suite-report.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "suite-report.md"), []byte(renderSuiteMarkdown(report)), 0o644)
}

func renderSuiteMarkdown(report SuiteReport) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# %s benchmark suite\n\n", report.Suite.Name)
	fmt.Fprintf(&out, "Generated: %s · Tasks: %d\n\n", report.GeneratedAt.Format(time.RFC3339), len(report.Tasks))
	out.WriteString("| Task | Condition | Success | Median tokens | Median commands | Median duration |\n|---|---|---:|---:|---:|---:|\n")
	for _, task := range report.Tasks {
		for _, condition := range benchmarkConditions(task.Report.Spec) {
			runs := selectRuns(task.Report.Runs, condition.Name)
			fmt.Fprintf(&out, "| %s | %s | %.0f%% | %d | %d | %.1fs |\n", task.Report.Spec.Task, condition.Name, successRate(runs), medianInts(runs, func(run Run) int { return run.TotalTokens() }), medianInts(runs, func(run Run) int { return run.Commands }), medianDurations(runs))
		}
	}
	if len(report.Regressions) == 0 {
		out.WriteString("\n**Regression gate:** passed; learned guidance did not reduce task success.\n")
	} else {
		fmt.Fprintf(&out, "\n**Regression gate:** failed for %s.\n", strings.Join(report.Regressions, ", "))
	}
	return out.String()
}
