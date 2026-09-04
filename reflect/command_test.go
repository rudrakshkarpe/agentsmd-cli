package reflect

import "testing"

func TestValidateAcceptsRuleAndNoOpVerdicts(t *testing.T) {
	if err := Validate(Result{Verdict: MissingRule, Rule: "Run focused tests first.", Confidence: 0.8}); err != nil {
		t.Fatal(err)
	}
	if err := Validate(Result{Verdict: NotRelevant, Confidence: 0.7}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejectsMalformedResults(t *testing.T) {
	tests := []Result{
		{Verdict: MissingRule, Confidence: 0.8},
		{Verdict: NotRelevant, Rule: "should not exist", Confidence: 0.7},
		{Verdict: Verdict("surprise"), Confidence: 0.5},
		{Verdict: MissingRule, Rule: "rule", Confidence: 1.1},
	}
	for _, test := range tests {
		if err := Validate(test); err == nil {
			t.Fatalf("expected error for %+v", test)
		}
	}
}
