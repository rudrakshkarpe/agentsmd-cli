package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestTerminalUIUsesGreenOnlyWhenEnabled(t *testing.T) {
	plain := terminalUI{}
	if got := plain.brand("agentsmd CLI"); got != "agentsmd CLI" {
		t.Fatalf("plain brand=%q", got)
	}
	colored := terminalUI{interactive: true, color: true}
	if got := colored.brand("agentsmd CLI"); !strings.Contains(got, ansiGreen) || !strings.Contains(got, ansiReset) {
		t.Fatalf("colored brand=%q", got)
	}
	if got := colored.icon("🌱"); got != "🌱 " {
		t.Fatalf("icon=%q", got)
	}
}

func TestHelpIdentifiesAgentsmdCLIWithoutANSIWhenPiped(t *testing.T) {
	command := New()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"--help"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "agentsmd CLI") {
		t.Fatalf("help=%q", output.String())
	}
	if strings.Contains(output.String(), "\033[") {
		t.Fatalf("piped help contains ANSI: %q", output.String())
	}
}

func TestVersionOutputRemainsScriptCompatible(t *testing.T) {
	command := New()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"--version"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "agentsmd version "+Version+"\n"; got != want {
		t.Fatalf("version=%q want %q", got, want)
	}
}
