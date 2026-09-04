// Package cli exposes the Cobra command tree for embedding and testing.
package cli

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

const Version = "0.1.0-dev"

type app struct {
	root string
	out  io.Writer
	err  io.Writer
}

func New() *cobra.Command {
	state := &app{out: os.Stdout, err: os.Stderr}
	root := &cobra.Command{
		Use:           "agentsmd",
		Short:         "Version-controlled authoring and self-improvement for AGENTS.md",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(state.out)
	root.SetErr(state.err)
	root.PersistentFlags().StringVar(&state.root, "root", ".", "project directory")
	root.AddCommand(
		state.initCommand(),
		state.templateCommand(),
		state.editCommand(),
		state.renderCommand(),
		state.lintCommand(),
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
	return root
}
