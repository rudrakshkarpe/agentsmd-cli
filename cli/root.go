// Package cli exposes the Cobra command tree for embedding and testing.
package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// Version is replaced at release time with -ldflags. Keeping it a variable
// lets source builds remain honest about not being an official release.
var Version = "dev"

type app struct {
	root     string
	out      io.Writer
	err      io.Writer
	lookPath func(string) (string, error)
}

func New() *cobra.Command {
	state := &app{out: os.Stdout, err: os.Stderr, lookPath: exec.LookPath}
	root := &cobra.Command{
		Use:           "agentsmd",
		Short:         "Version-controlled authoring and self-improvement for AGENTS.md",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run:           func(cmd *cobra.Command, _ []string) { printWelcome(cmd) },
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetOut(state.out)
	root.SetErr(state.err)
	root.SetHelpFunc(func(cmd *cobra.Command, _ []string) { printHelp(cmd) })
	// Keep the version line stable for installers and shell scripts. The
	// interactive landing screen and help identify the product as agentsmd CLI.
	root.SetVersionTemplate("agentsmd version {{.Version}}\n")
	root.PersistentFlags().StringVar(&state.root, "root", ".", "project directory")
	root.AddCommand(
		state.initCommand(),
		state.connectCommand(),
		state.automateCommand(),
		state.hookCommand(),
		state.ingestCommand(),
		state.processCommand(),
		state.doctorCommand(),
		state.updateCommand(),
		state.sessionsCommand(),
		state.templateCommand(),
		state.editCommand(),
		state.renderCommand(),
		state.lintCommand(),
		state.captureCommand(),
		state.logCommand(),
		state.diffCommand(),
		state.commitCommand(),
		state.revertCommand(),
		state.tagCommand(),
		state.blameCommand(),
		state.statusCommand(),
		state.learnCommand(),
		state.pendingCommand(),
		state.promoteCommand(),
		state.rejectCommand(),
		state.recordCommand(),
		state.pruneCommand(),
		state.savingsCommand(),
	)
	for _, name := range []string{"edit", "render", "lint", "capture", "log", "diff", "commit", "revert", "tag", "blame", "status", "record", "prune", "savings"} {
		if command, _, err := root.Find([]string{name}); err == nil {
			command.Hidden = true
		}
	}
	return root
}

const solidBanner = ` ███   ████ █████ █   █ █████  ████ █   █ ████         ████ █     █████
█   █ █     █     ██  █   █   █     ██ ██ █   █       █     █       █
█████ █  ██ ████  █ █ █   █    ███  █ █ █ █   █ █████ █     █       █
█   █ █   █ █     █  ██   █       █ █   █ █   █       █     █       █
█   █  ███  █████ █   █   █   ████  █   █ ████         ████ █████ █████`

func printWelcome(cmd *cobra.Command) {
	ui := uiFor(cmd)
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s\n\n", ui.brand(solidBanner))
	fmt.Fprintln(out, ui.soft("Project-aware guidance for coding agents"), ui.muted("·"), ui.muted("agentsmd-cli"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, ui.brand("Start"))
	fmt.Fprintf(out, "  %s  %s\n", ui.accent(fmt.Sprintf("%-24s", "agentsmd init")), ui.muted("detect the project and create AGENTS.md"))
	fmt.Fprintf(out, "  %s  %s\n", ui.accent(fmt.Sprintf("%-24s", "agentsmd templates")), ui.muted("browse curated starting points"))
	fmt.Fprintf(out, "  %s  %s\n", ui.accent(fmt.Sprintf("%-24s", "agentsmd connect <cli>")), ui.muted("connect a coding-agent CLI"))
	fmt.Fprintf(out, "  %s  %s\n", ui.accent(fmt.Sprintf("%-24s", "agentsmd doctor")), ui.muted("inspect project and integrations"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, ui.muted("Run `agentsmd --help` for all commands."))
}
