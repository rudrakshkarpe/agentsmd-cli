package learning_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/learning"
	"github.com/rudrakshkarpe/agentsmd-cli/project"
	reflector "github.com/rudrakshkarpe/agentsmd-cli/reflect"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
	"github.com/rudrakshkarpe/agentsmd-cli/version"
)

type fakeReflector struct {
	result reflector.Result
}

func (f fakeReflector) Reflect(context.Context, schema.Trajectory) (reflector.Result, error) {
	return f.result, nil
}

func TestLearnAllowsNoOpAndPersistsTargetedRule(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	service := learning.New(p)
	trajectory := schema.Trajectory{SessionID: "session-1", Task: "demo"}
	proposal, result, err := service.Learn(context.Background(), trajectory, fakeReflector{result: reflector.Result{Verdict: reflector.NotRelevant}})
	if err != nil || proposal != nil || result.Verdict != reflector.NotRelevant {
		t.Fatalf("proposal=%v result=%v err=%v", proposal, result, err)
	}
	proposal, _, err = service.Learn(context.Background(), trajectory, fakeReflector{result: reflector.Result{Verdict: reflector.MissingRule, Rule: "Read the fixture before editing the parser.", Confidence: 0.9}})
	if err != nil || proposal == nil {
		t.Fatalf("proposal=%v err=%v", proposal, err)
	}
	if proposal.Origin.Run != "session-1" || proposal.Origin.Task != "demo" {
		t.Fatalf("origin=%+v", proposal.Origin)
	}
}

func TestProposalPromotionIsGatedAndVersioned(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	service := learning.New(p)
	service.Now = func() time.Time { return time.Date(2026, 9, 5, 1, 2, 3, 4, time.UTC) }
	proposal, err := service.Propose("Read SPEC.md before changing stored data.", schema.Origin{Run: "s1", Task: "schema"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p.ArtifactPath()); !os.IsNotExist(err) {
		t.Fatal("proposing a rule must not modify AGENTS.md")
	}
	rule, duplicate, err := service.Promote(proposal.ID)
	if err != nil || duplicate != nil || rule.ID != "r000" {
		t.Fatalf("rule=%v duplicate=%v err=%v", rule, duplicate, err)
	}
	data, _ := os.ReadFile(p.ArtifactPath())
	if !strings.Contains(string(data), "Read SPEC.md") {
		t.Fatal("promoted rule was not rendered")
	}
	items, err := version.New(p).Log()
	if err != nil || len(items) != 1 || items[0].Reason != "learned" {
		t.Fatalf("versions=%v err=%v", items, err)
	}
}

func TestSavings(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	value := schema.Ledger{Rules: []schema.Rule{}, Runs: map[string][]int{"demo": {1000, 700}}}
	if err := p.SaveLedger(value); err != nil {
		t.Fatal(err)
	}
	result, err := learning.New(p).Savings("demo")
	if err != nil || result.Percent != 30 {
		t.Fatalf("result=%v err=%v", result, err)
	}
}
