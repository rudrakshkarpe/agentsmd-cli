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
	root.PersistentFlags().StringVar(&state.root, "root", ".", "project directory")
	root.AddCommand(
		state.initCommand(),
		state.connectCommand(),
		state.hookCommand(),
		state.doctorCommand(),
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

const dotBanner = `·●●· ●●●● ●●●● ●··● ●●●● ●●●● ●···● ●●●·
●··● ●··· ●··· ●●·● ··●· ●··· ●●·●● ●··●
●●●● ●·●● ●●●· ●·●● ··●· ●●●● ●·●·● ●··●
●··● ●··● ●··· ●··● ··●· ···● ●···● ●··●
●··● ●●●● ●●●● ●··● ··●· ●●●● ●···● ●●●·`

func printWelcome(cmd *cobra.Command) {
	green, reset := "", ""
	if os.Getenv("NO_COLOR") == "" {
		if file, ok := cmd.OutOrStdout().(*os.File); ok {
			if info, err := file.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
				green, reset = "\033[92m", "\033[0m"
			}
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s\n\n", green, dotBanner, reset)
	fmt.Fprintln(cmd.OutOrStdout(), "Project-aware AGENTS.md setup and a reviewable self-improvement loop.")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "  agentsmd init                 detect this project and create AGENTS.md")
	fmt.Fprintln(cmd.OutOrStdout(), "  agentsmd templates            browse reusable starting points")
	fmt.Fprintln(cmd.OutOrStdout(), "  agentsmd connect <cli>        connect codex, claude, goose, or cursor")
	fmt.Fprintln(cmd.OutOrStdout(), "  agentsmd doctor               check your project and installed CLIs")
	fmt.Fprintln(cmd.OutOrStdout(), "  agentsmd --help               show all primary commands")
}
