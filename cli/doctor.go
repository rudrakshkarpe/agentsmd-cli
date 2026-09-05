package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/rudrakshkarpe/agentsmd-cli/automation"
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
			ui := uiFor(cmd)
			fmt.Fprintln(cmd.OutOrStdout(), ui.icon("🩺")+ui.brand("agentsmd CLI · doctor"))
			fmt.Fprintln(cmd.OutOrStdout(), ui.muted("Checking project health and agent integrations…"))
			fmt.Fprintln(cmd.OutOrStdout())
			warnings := 0
			for _, item := range checks {
				if item.level != "ok" {
					warnings++
				}
				if ui.interactive {
					symbol := map[string]string{"ok": "✅", "warn": "⚠️", "error": "❌"}[item.level]
					fmt.Fprintf(cmd.OutOrStdout(), "%s%s %s\n", ui.icon(symbol), ui.brand(item.name), ui.muted("— "+item.detail))
					continue
				}
				symbol := map[string]string{"ok": "✓", "warn": "!", "error": "✗"}[item.level]
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s — %s\n", symbol, item.name, item.detail)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			if warnings == 0 {
				writeSuccess(cmd, "Everything is ready.")
			} else {
				writeInfo(cmd, "Optional integrations may be connected when you need them.")
			}
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
	p, err := project.Require(a.root)
	if err != nil {
		p, _ = project.Open(a.root)
	}
	if _, err := os.Stat(p.ArtifactPath()); err == nil {
		result = append(result, check{"ok", "AGENTS.md", p.ArtifactPath()})
	} else {
		result = append(result, check{"error", "AGENTS.md", "not initialized; run agentsmd init"})
	}
	config, configErr := automation.Load(p)
	switch {
	case configErr != nil:
		result = append(result, check{"error", "Automation", configErr.Error()})
	case len(config.ReflectCommand) == 0:
		result = append(result, check{"warn", "Automation", "reflection disabled; run agentsmd automate --help"})
	case len(config.EvaluateCommand) == 0:
		result = append(result, check{"warn", "Automation", "reflection enabled; no evaluation gate configured"})
	case config.AutoPromote:
		result = append(result, check{"ok", "Automation", "reflection and gated auto-promotion enabled"})
	default:
		result = append(result, check{"ok", "Automation", "reflection and evaluation enabled; promotion is manual"})
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
		{"klaatcode", []string{"klaatai"}},
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
	if value == "klaatcode" {
		return "KlaatCode"
	}
	if value == "claude" {
		return "Claude Code"
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
