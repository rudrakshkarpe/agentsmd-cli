package ledger_test

import (
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/ledger"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

func TestAddRejectsDuplicateAndRendersProvenanceRule(t *testing.T) {
	value := schema.Ledger{Rules: []schema.Rule{}, Runs: map[string][]int{}}
	rule, duplicate, err := ledger.Add(&value, "Run focused tests before the full suite.", schema.Origin{Run: "s1"})
	if err != nil || duplicate != nil {
		t.Fatalf("first add: rule=%v duplicate=%v err=%v", rule, duplicate, err)
	}
	second, duplicate, err := ledger.Add(&value, "Run the focused test suite before full tests.", schema.Origin{})
	if err != nil || second != nil || duplicate == nil {
		t.Fatalf("duplicate add: rule=%v duplicate=%v err=%v", second, duplicate, err)
	}
	if !strings.Contains(ledger.Render(value), "[r000] Run focused tests") {
		t.Fatal("rendered document is missing the rule")
	}
}

func TestIDsAreNotReusedAfterPruning(t *testing.T) {
	value := schema.Ledger{Rules: []schema.Rule{{ID: "r004", Text: "existing"}}, Runs: map[string][]int{}}
	rule, _, err := ledger.Add(&value, "a distinct new instruction", schema.Origin{})
	if err != nil {
		t.Fatal(err)
	}
	if rule.ID != "r005" {
		t.Fatalf("got %s, want r005", rule.ID)
	}
}
