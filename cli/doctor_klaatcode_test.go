package cli

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/integration"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

func TestDoctorFindsKlaatCodeExecutable(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	a := &app{root: p.Root, lookPath: func(name string) (string, error) {
		if name == "klaatai" {
			return "/tools/klaatai", nil
		}
		return "", exec.ErrNotFound
	}}
	for _, connected := range []bool{false, true} {
		if connected {
			if _, err := integration.Connect(p, "klaatcode"); err != nil {
				t.Fatal(err)
			}
		}
		found := false
		for _, item := range a.diagnose() {
			if item.name != "KlaatCode" {
				continue
			}
			found = true
			if connected {
				if item.level != "ok" || !strings.Contains(item.detail, "connected") {
					t.Fatalf("check=%+v", item)
				}
			} else if !strings.Contains(item.detail, "agentsmd connect klaatcode") {
				t.Fatalf("check=%+v", item)
			}
		}
		if !found {
			t.Fatal("KlaatCode diagnostic missing")
		}
	}
}

func TestDoctorReportsKlaatCodeRuleShadowing(t *testing.T) {
	for _, tc := range []struct {
		name, content       string
		exists, wantWarning bool
	}{
		{"missing", "", false, false},
		{"empty", " \n\t", true, false},
		{"nonempty", "Use these instructions", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := project.Open(t.TempDir())
			if err := p.Scaffold(); err != nil {
				t.Fatal(err)
			}
			if _, err := integration.Connect(p, "klaatcode"); err != nil {
				t.Fatal(err)
			}
			if tc.exists {
				if err := project.AtomicWrite(filepath.Join(p.Root, ".klaatai", "rules.md"), []byte(tc.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			a := &app{root: p.Root, lookPath: func(name string) (string, error) { return "/tools/" + name, nil }}
			found := false
			for _, item := range a.diagnose() {
				if item.name == "KlaatCode instructions" && item.level == "warn" && strings.Contains(item.detail, "shadows AGENTS.md") {
					found = true
				}
			}
			if found != tc.wantWarning {
				t.Fatalf("warning=%v want=%v", found, tc.wantWarning)
			}
		})
	}
}
