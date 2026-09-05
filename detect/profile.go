// Package detect inspects a repository and derives a conservative AGENTS.md baseline.
package detect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Command struct {
	Purpose string
	Value   string
}

type Profile struct {
	Stacks         []string
	Commands       []Command
	Signals        []string
	PackageManager string
	Workspace      bool
}

func Inspect(root string) (Profile, error) {
	profile := Profile{}
	has := func(name string) bool {
		_, err := os.Stat(filepath.Join(root, name))
		return err == nil
	}
	addStack := func(stack, signal string) {
		profile.Stacks = append(profile.Stacks, stack)
		profile.Signals = append(profile.Signals, signal)
	}
	addCommand := func(purpose, value string) {
		for _, item := range profile.Commands {
			if item.Value == value {
				return
			}
		}
		profile.Commands = append(profile.Commands, Command{Purpose: purpose, Value: value})
	}

	if has("go.mod") {
		addStack("Go", "go.mod")
		addCommand("Test", "go test ./...")
		addCommand("Static analysis", "go vet ./...")
	}
	if has("package.json") {
		addStack("Node.js", "package.json")
		manager := "npm"
		install := "npm install"
		switch {
		case has("pnpm-lock.yaml"):
			manager, install = "pnpm", "pnpm install"
			profile.Signals = append(profile.Signals, "pnpm-lock.yaml")
		case has("yarn.lock"):
			manager, install = "yarn", "yarn install"
			profile.Signals = append(profile.Signals, "yarn.lock")
		case has("bun.lock"), has("bun.lockb"):
			manager, install = "bun", "bun install"
			profile.Signals = append(profile.Signals, firstExisting(has, "bun.lock", "bun.lockb"))
		case has("package-lock.json"):
			install = "npm ci"
			profile.Signals = append(profile.Signals, "package-lock.json")
		}
		profile.PackageManager = manager
		profile.Workspace = has("pnpm-workspace.yaml") || has("turbo.json") || has("nx.json") || packageJSONHasWorkspaces(root)
		if workspaceSignal := firstExisting(has, "pnpm-workspace.yaml", "turbo.json", "nx.json"); workspaceSignal != "" {
			profile.Signals = append(profile.Signals, workspaceSignal)
		}
		addCommand("Install", install)
		data, err := os.ReadFile(filepath.Join(root, "package.json"))
		if err != nil {
			return Profile{}, err
		}
		var manifest struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(data, &manifest) == nil {
			for _, name := range []string{"dev", "test", "lint", "typecheck", "check", "build"} {
				if _, ok := manifest.Scripts[name]; ok {
					purpose := strings.ToUpper(name[:1]) + name[1:]
					addCommand(purpose, manager+" run "+name)
				}
			}
		}
	}
	if has("pyproject.toml") || has("requirements.txt") {
		addStack("Python", firstExisting(has, "pyproject.toml", "requirements.txt"))
		switch {
		case has("uv.lock"):
			addCommand("Install", "uv sync")
		case has("poetry.lock"):
			addCommand("Install", "poetry install")
		case has("requirements.txt"):
			addCommand("Install", "python -m pip install -r requirements.txt")
		default:
			addCommand("Install", "python -m pip install -e .")
		}
		addCommand("Test", "python -m pytest")
	}
	if has("Cargo.toml") {
		addStack("Rust", "Cargo.toml")
		addCommand("Test", "cargo test")
		addCommand("Lint", "cargo clippy --all-targets --all-features")
		addCommand("Format check", "cargo fmt --all -- --check")
	}
	if has("pubspec.yaml") {
		addStack("Flutter/Dart", "pubspec.yaml")
		addCommand("Install", "flutter pub get")
		addCommand("Analyze", "flutter analyze")
		addCommand("Test", "flutter test")
	}
	if has("pom.xml") {
		addStack("Java/Maven", "pom.xml")
		addCommand("Test", "./mvnw test")
	}
	if has("build.gradle") || has("build.gradle.kts") {
		addStack("Java/Gradle", firstExisting(has, "build.gradle.kts", "build.gradle"))
		addCommand("Test", "./gradlew test")
	}
	if has("Makefile") {
		profile.Signals = append(profile.Signals, "Makefile")
	}
	if has("go.work") {
		profile.Workspace = true
		profile.Signals = append(profile.Signals, "go.work")
	}
	if has(".github/workflows") {
		profile.Signals = append(profile.Signals, ".github/workflows")
	}

	sort.Strings(profile.Signals)
	return profile, nil
}

func packageJSONHasWorkspaces(root string) bool {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return false
	}
	var manifest map[string]json.RawMessage
	if json.Unmarshal(data, &manifest) != nil {
		return false
	}
	_, ok := manifest["workspaces"]
	return ok
}

func firstExisting(has func(string) bool, names ...string) string {
	for _, name := range names {
		if has(name) {
			return name
		}
	}
	return ""
}

func (p Profile) Render() string {
	var output strings.Builder
	output.WriteString("# AGENTS.md\n\n")
	output.WriteString("> Repository guidance for coding agents. Keep this file concise, concrete, and current.\n\n")
	output.WriteString("## Project overview\n\n")
	if len(p.Stacks) == 0 {
		output.WriteString("- No supported manifest was detected yet. Inspect the repository before choosing tools or commands.\n")
	} else {
		fmt.Fprintf(&output, "- Detected stack: %s.\n", strings.Join(p.Stacks, ", "))
		fmt.Fprintf(&output, "- Detection evidence: `%s`.\n", strings.Join(p.Signals, "`, `"))
	}
	if p.Workspace {
		output.WriteString("- This is a workspace or monorepo. Identify the owning package before editing.\n")
	}

	output.WriteString("\n## Development environment\n\n")
	if p.PackageManager != "" {
		fmt.Fprintf(&output, "- Use `%s` for JavaScript dependencies; do not introduce a second package manager.\n", p.PackageManager)
	}
	if command, ok := p.command("Install"); ok {
		fmt.Fprintf(&output, "- Bootstrap dependencies with `%s`.\n", command)
	} else {
		output.WriteString("- Use the repository's existing toolchain and lockfiles. Do not guess setup commands.\n")
	}
	output.WriteString("- Read nearby documentation and configuration before changing build or tooling behavior.\n")
	if p.Workspace {
		output.WriteString("- Run dependency commands from the workspace root unless package documentation says otherwise.\n")
	}

	validationCommands := p.validationCommands()
	if len(validationCommands) > 0 {
		output.WriteString("\n## Commands and validation\n\n")
		for _, command := range validationCommands {
			fmt.Fprintf(&output, "- %s: `%s`\n", command.Purpose, command.Value)
		}
	}

	output.WriteString("\n## Testing instructions\n\n")
	output.WriteString("- Start with the narrowest relevant test or package, then run the broader checks listed above.\n")
	output.WriteString("- Add or update tests beside the behavior being changed. Cover failures and edge cases, not only the happy path.\n")
	output.WriteString("- Treat an existing failure as evidence to investigate; do not weaken tests merely to make them pass.\n")

	output.WriteString("\n## Code and dependency conventions\n\n")
	for _, rule := range p.conventions() {
		fmt.Fprintf(&output, "- %s\n", rule)
	}
	output.WriteString("- Follow nearby naming, structure, error-handling, and documentation patterns.\n")
	output.WriteString("- Keep dependency changes intentional and update the existing lockfile in the same change.\n")
	output.WriteString("- Do not commit secrets, credentials, generated local state, or unrelated formatting churn.\n")

	output.WriteString("\n## Agent working agreement\n\n")
	output.WriteString("- Inspect relevant files and current git changes before editing; preserve user work already in progress.\n")
	output.WriteString("- Prefer the smallest complete change. Avoid broad rewrites unless the task requires one.\n")
	output.WriteString("- State assumptions, surface blockers early, and never claim a check passed unless it was run.\n")
	output.WriteString("- At handoff, summarize changed behavior, validation performed, and any remaining risks.\n")
	return output.String()
}

func (p Profile) validationCommands() []Command {
	commands := make([]Command, 0, len(p.Commands))
	for _, command := range p.Commands {
		if command.Purpose != "Install" {
			commands = append(commands, command)
		}
	}
	return commands
}

func (p Profile) command(purpose string) (string, bool) {
	for _, command := range p.Commands {
		if command.Purpose == purpose {
			return command.Value, true
		}
	}
	return "", false
}

func (p Profile) conventions() []string {
	rules := make([]string, 0, len(p.Stacks))
	for _, stack := range p.Stacks {
		switch stack {
		case "Go":
			rules = append(rules, "Format Go changes with `gofmt`; keep packages focused and errors explicit.")
		case "Node.js":
			rules = append(rules, "Respect the configured module system, formatter, linter, and type checker; avoid bypassing them.")
		case "Python":
			rules = append(rules, "Use the configured virtual environment and project tooling; keep imports and public types explicit.")
		case "Rust":
			rules = append(rules, "Keep `rustfmt` clean and address Clippy findings without blanket suppressions.")
		case "Flutter/Dart":
			rules = append(rules, "Keep Dart formatting and analyzer output clean; follow existing widget and state-management patterns.")
		case "Java/Maven", "Java/Gradle":
			rules = append(rules, "Preserve the configured Java version, package layout, and build-tool conventions.")
		}
	}
	return rules
}
