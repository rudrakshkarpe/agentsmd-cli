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
	Stacks   []string
	Commands []Command
	Signals  []string
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
		case has("yarn.lock"):
			manager, install = "yarn", "yarn install"
		case has("bun.lock"), has("bun.lockb"):
			manager, install = "bun", "bun install"
		case has("package-lock.json"):
			install = "npm ci"
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
	if has(".github/workflows") {
		profile.Signals = append(profile.Signals, ".github/workflows")
	}

	sort.Strings(profile.Signals)
	return profile, nil
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
	output.WriteString("## Project profile\n\n")
	if len(p.Stacks) == 0 {
		output.WriteString("- No supported project manifest was detected. Add repository-specific setup and verification commands.\n")
	} else {
		fmt.Fprintf(&output, "- Detected stack: %s.\n", strings.Join(p.Stacks, ", "))
		fmt.Fprintf(&output, "- Detection evidence: `%s`.\n", strings.Join(p.Signals, "`, `"))
	}
	if len(p.Commands) > 0 {
		output.WriteString("\n## Commands\n\n")
		for _, command := range p.Commands {
			fmt.Fprintf(&output, "- %s: `%s`\n", command.Purpose, command.Value)
		}
	}
	output.WriteString("\n## Working agreements\n\n")
	output.WriteString("- Prefer focused validation before running the complete suite.\n")
	output.WriteString("- Keep changes scoped and preserve existing project conventions.\n")
	return output.String()
}
