package template_test

import (
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/template"
)

func TestTemplatesProvideActionableStartingPoints(t *testing.T) {
	names, err := template.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		content, err := template.Load(name)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(content, "# AGENTS.md") {
			t.Fatalf("template %s has no AGENTS.md heading", name)
		}
		if strings.Count(content, "## ") < 3 {
			t.Fatalf("template %s is too generic:\n%s", name, content)
		}
	}
}
