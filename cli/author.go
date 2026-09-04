package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rudrakshkarpe/agentsmd-cli/detect"
	"github.com/rudrakshkarpe/agentsmd-cli/ledger"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/rudrakshkarpe/agentsmd-cli/template"
	"github.com/rudrakshkarpe/agentsmd-cli/version"
	"github.com/spf13/cobra"
)

func (a *app) initCommand() *cobra.Command {
	var templateName string
	var scratch bool
	var force bool
	command := &cobra.Command{
		Use:   "init",
		Short: "Create AGENTS.md and the .agentsmd store",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := project.Open(a.root)
			if err != nil {
				return err
			}
			if err := p.Scaffold(); err != nil {
				return err
			}
			content := "# AGENTS.md\n"
			reason := "manual"
			meta := map[string]any{}
			if templateName != "" {
				content, err = template.Load(templateName)
				if err != nil {
					return err
				}
				reason = "template"
				meta["template"] = templateName
			} else if !scratch {
				profile, inspectErr := detect.Inspect(p.Root)
				if inspectErr != nil {
					return inspectErr
				}
				content = profile.Render()
				reason = "auto-detected"
				meta["stacks"] = profile.Stacks
			}
			content = ledger.Merge(content, structuredEmptyLedger())
			if err := writeArtifact(p, content, force); err != nil {
				return err
			}
			if err := p.SaveLedger(structuredEmptyLedger()); err != nil {
				return err
			}
			item, err := version.New(p).Commit("init", reason, meta)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "initialized %s (%s)\n", p.Root, item.ID)
			return nil
		},
	}
	command.Flags().StringVar(&templateName, "template", "", "built-in template name")
	command.Flags().BoolVar(&scratch, "scratch", false, "start with a minimal file")
	command.Flags().BoolVar(&force, "force", false, "replace an existing AGENTS.md")
	return command
}

func (a *app) editCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open AGENTS.md in $EDITOR",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			parts := strings.Fields(os.Getenv("EDITOR"))
			if len(parts) == 0 {
				parts = []string{"vi"}
			}
			process := exec.Command(parts[0], append(parts[1:], p.ArtifactPath())...)
			process.Stdin = os.Stdin
			process.Stdout = cmd.OutOrStdout()
			process.Stderr = cmd.ErrOrStderr()
			return process.Run()
		},
	}
}

func structuredEmptyLedger() schema.Ledger {
	return schema.Ledger{Rules: []schema.Rule{}, Runs: map[string][]int{}}
}

func (a *app) templateCommand() *cobra.Command {
	list := func(cmd *cobra.Command) error {
		names, err := template.List()
		if err != nil {
			return err
		}
		descriptions := map[string]string{"minimal": "small universal baseline", "team": "shared review and collaboration rules", "monorepo": "multi-package repository workflow", "python-lib": "Python library conventions", "benchmark-kit": "reproducible evaluation projects"}
		for _, name := range names {
			fmt.Fprintf(cmd.OutOrStdout(), "%-16s %s\n", name, descriptions[name])
		}
		return nil
	}
	command := &cobra.Command{Use: "templates", Aliases: []string{"template"}, Short: "Browse or apply AGENTS.md templates", RunE: func(cmd *cobra.Command, _ []string) error { return list(cmd) }}
	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List embedded templates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return list(cmd)
		},
	})
	command.AddCommand(&cobra.Command{
		Use:   "use NAME",
		Args:  cobra.ExactArgs(1),
		Short: "Apply an embedded template",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			content, err := template.Load(args[0])
			if err != nil {
				return err
			}
			if err := project.AtomicWrite(p.ArtifactPath(), []byte(content), 0o644); err != nil {
				return err
			}
			item, err := version.New(p).Commit("template: "+args[0], "template", map[string]any{"template": args[0]})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "applied %s (%s)\n", args[0], item.ID)
			return nil
		},
	})
	return command
}

func (a *app) renderCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "render",
		Short: "Render AGENTS.md from the rule ledger",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			value, err := p.LoadLedger()
			if err != nil {
				return err
			}
			existing, err := os.ReadFile(p.ArtifactPath())
			if err != nil {
				return err
			}
			if err := project.AtomicWrite(p.ArtifactPath(), []byte(ledger.Merge(string(existing), value)), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "rendered %d rules\n", len(value.Rules))
			return nil
		},
	}
}

func (a *app) lintCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "lint",
		Short: "Find duplicate and unused rules",
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.requireProject()
			if err != nil {
				return err
			}
			value, err := p.LoadLedger()
			if err != nil {
				return err
			}
			issues := ledger.Lint(value)
			for _, issue := range issues {
				fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] %s\n", issue.Kind, issue.RuleID, issue.Message)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d issue(s)\n", len(issues))
			return nil
		},
	}
}
