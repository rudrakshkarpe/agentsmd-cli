package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

const (
	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiDim       = "\033[2m"
	ansiGreen     = "\033[38;5;82m"
	ansiSoftGreen = "\033[38;5;114m"
)

type terminalUI struct {
	interactive bool
	color       bool
}

func uiFor(cmd *cobra.Command) terminalUI {
	interactive := isTerminal(cmd.OutOrStdout())
	return terminalUI{
		interactive: interactive,
		color:       interactive && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb",
	}
}

func isTerminal(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func (ui terminalUI) styled(code, value string) string {
	if !ui.color {
		return value
	}
	return code + value + ansiReset
}

func (ui terminalUI) brand(value string) string {
	return ui.styled(ansiBold+ansiGreen, value)
}

func (ui terminalUI) accent(value string) string {
	return ui.styled(ansiGreen, value)
}

func (ui terminalUI) soft(value string) string {
	return ui.styled(ansiSoftGreen, value)
}

func (ui terminalUI) muted(value string) string {
	return ui.styled(ansiDim+ansiSoftGreen, value)
}

func (ui terminalUI) icon(value string) string {
	if !ui.interactive {
		return ""
	}
	return value + " "
}

func writeSection(cmd *cobra.Command, icon, title string) {
	ui := uiFor(cmd)
	fmt.Fprintf(cmd.OutOrStdout(), "\n%s%s\n", ui.icon(icon), ui.brand(title))
}

func writeSuccess(cmd *cobra.Command, message string) {
	ui := uiFor(cmd)
	fmt.Fprintln(cmd.OutOrStdout(), ui.icon("✅")+ui.accent(message))
}

func writeInfo(cmd *cobra.Command, message string) {
	ui := uiFor(cmd)
	fmt.Fprintln(cmd.OutOrStdout(), ui.icon("ℹ️")+ui.soft(message))
}

func writeWarning(cmd *cobra.Command, message string) {
	ui := uiFor(cmd)
	fmt.Fprintln(cmd.OutOrStdout(), ui.icon("⚠️")+ui.soft(message))
}

func commandIcon(name string) string {
	icons := map[string]string{
		"init":      "🌱",
		"templates": "📚",
		"connect":   "🔌",
		"automate":  "🧠",
		"doctor":    "🩺",
		"update":    "🔄",
		"sessions":  "📡",
		"benchmark": "🧪",
		"learn":     "✨",
		"pending":   "⏳",
		"promote":   "✅",
		"reject":    "🗑️",
		"help":      "❓",
	}
	return icons[name]
}

func printHelp(cmd *cobra.Command) {
	ui := uiFor(cmd)
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, ui.brand("agentsmd-cli"))
	description := "Project-aware instructions that learn from agent sessions."
	if cmd.Parent() != nil && cmd.Short != "" {
		description = cmd.Short
	}
	fmt.Fprintln(out, ui.muted(description))
	if len(cmd.Aliases) > 0 {
		fmt.Fprintln(out, ui.muted("Aliases: agentsmd "+strings.Join(cmd.Aliases, ", agentsmd ")))
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, ui.brand("Usage"))
	fmt.Fprintf(out, "  %s\n", ui.accent(cmd.UseLine()))

	children := make([]*cobra.Command, 0)
	for _, child := range cmd.Commands() {
		if child.IsAvailableCommand() && child.Name() != "help" {
			children = append(children, child)
		}
	}
	if len(children) > 0 {
		order := map[string]int{
			"init": 0, "templates": 1, "connect": 2, "automate": 3, "doctor": 4,
			"update": 5, "sessions": 6, "benchmark": 7, "learn": 8, "pending": 9, "promote": 10, "reject": 11,
		}
		sort.SliceStable(children, func(i, j int) bool {
			left, leftOK := order[children[i].Name()]
			right, rightOK := order[children[j].Name()]
			if leftOK && rightOK {
				return left < right
			}
			if leftOK != rightOK {
				return leftOK
			}
			return children[i].Name() < children[j].Name()
		})
		fmt.Fprintln(out)
		fmt.Fprintln(out, ui.brand("Commands"))
		for _, child := range children {
			icon := ui.icon(commandIcon(child.Name()))
			fmt.Fprintf(out, "  %s%-12s %s\n", icon, ui.accent(child.Name()), ui.soft(child.Short))
		}
	}

	flags := strings.TrimRight(cmd.Flags().FlagUsages(), "\n")
	if flags != "" {
		fmt.Fprintln(out)
		fmt.Fprintln(out, ui.brand("Flags"))
		fmt.Fprintln(out, ui.soft(flags))
	}
	if cmd.HasAvailableSubCommands() {
		fmt.Fprintln(out)
		fmt.Fprintln(out, ui.muted("Run `agentsmd <command> --help` for command details."))
	}
}
