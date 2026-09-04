package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/rudrakshkarpe/agentsmd-cli/integration"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/spf13/cobra"
)

type check struct {
	level  string
	name   string
	detail string
}

func (a *app) doctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check project setup and supported coding CLIs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			checks := a.diagnose()
			fmt.Fprintln(cmd.OutOrStdout(), "agentsmd doctor")
			for _, item := range checks {
				symbol := map[string]string{"ok": "✓", "warn": "!", "error": "✗"}[item.level]
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s — %s\n", symbol, item.name, item.detail)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nWarnings are optional integrations; connect only the CLIs you use.")
			return nil
		},
	}
}

func (a *app) diagnose() []check {
	result := []check{}
	if path, err := a.lookPath("git"); err == nil {
		result = append(result, check{"ok", "Git", path})
	} else {
		result = append(result, check{"error", "Git", "not found; install Git to version project changes"})
	}
	p, _ := project.Open(a.root)
	if _, err := os.Stat(p.ArtifactPath()); err == nil {
		result = append(result, check{"ok", "AGENTS.md", p.ArtifactPath()})
	} else {
		result = append(result, check{"error", "AGENTS.md", "not initialized; run agentsmd init"})
	}
	records, _ := integration.Load(p)
	connected := map[string]bool{}
	for _, record := range records {
		connected[record.Provider] = true
	}
	tools := []struct {
		provider string
		commands []string
	}{
		{"codex", []string{"codex"}}, {"claude", []string{"claude"}}, {"goose", []string{"goose"}}, {"cursor", []string{"agent", "cursor"}},
	}
	for _, tool := range tools {
		path := ""
		for _, command := range tool.commands {
			if candidate, err := a.lookPath(command); err == nil {
				path = candidate
				break
			}
		}
		if tool.provider == "cursor" && path == "" && runtime.GOOS == "darwin" {
			if _, err := os.Stat("/Applications/Cursor.app"); err == nil {
				path = "/Applications/Cursor.app"
			}
		}
		if path == "" {
			result = append(result, check{"warn", displayProvider(tool.provider), "not installed (optional)"})
			continue
		}
		if connected[tool.provider] {
			result = append(result, check{"ok", displayProvider(tool.provider), path + "; connected"})
		} else {
			result = append(result, check{"warn", displayProvider(tool.provider), path + "; run agentsmd connect " + tool.provider})
		}
	}
	return result
}

func displayProvider(value string) string {
	if value == "claude" {
		return "Claude Code"
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
