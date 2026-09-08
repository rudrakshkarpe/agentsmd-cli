package reflect

import (
	"context"
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

type recordingReflector struct{ trajectory schema.Trajectory }

func (r *recordingReflector) Reflect(_ context.Context, trajectory schema.Trajectory) (Result, error) {
	r.trajectory = trajectory
	return Result{Verdict: NotRelevant}, nil
}

func TestRedactingScrubsExternalCopyWithoutChangingStoredValue(t *testing.T) {
	secret := "token=sk-secret123"
	original := schema.Trajectory{
		SessionID: "user@example.com", Task: "private-ticket-42",
		Steps:     []schema.Step{{Role: "assistant", Summary: "used " + secret}},
		ToolCalls: []schema.ToolCall{{Name: "request", Args: map[string]any{"headers": map[string]any{"authorization": secret}, "items": []any{secret}}, Result: secret}},
		Files:     []schema.FileTouch{{Path: "/Users/alice/private.go", Diff: "+" + secret}},
		Commands:  []schema.Command{{Argv: []string{"curl", "-H", secret}}},
		FinalDiff: secret, Metadata: map[string]string{"email": "user@example.com"},
	}
	recorder := &recordingReflector{}
	_, err := (Redacting{Next: recorder, Patterns: []string{`sk-[A-Za-z0-9]+`, `[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+`, `/Users/[^/]+`}}).Reflect(context.Background(), original)
	if err != nil {
		t.Fatal(err)
	}
	encoded := strings.Join([]string{recorder.trajectory.SessionID, recorder.trajectory.Steps[0].Summary, recorder.trajectory.ToolCalls[0].Result, recorder.trajectory.Files[0].Path, recorder.trajectory.FinalDiff, recorder.trajectory.Metadata["email"]}, "\n")
	if strings.Contains(encoded, "secret123") || strings.Contains(encoded, "user@example.com") || strings.Contains(encoded, "/Users/alice") || !strings.Contains(encoded, redacted) {
		t.Fatalf("redacted trajectory=%+v", recorder.trajectory)
	}
	if !strings.Contains(original.Steps[0].Summary, "sk-secret123") || original.Metadata["email"] != "user@example.com" {
		t.Fatalf("original was changed: %+v", original)
	}
	nested := recorder.trajectory.ToolCalls[0].Args["headers"].(map[string]any)["authorization"].(string)
	if strings.Contains(nested, "sk-secret123") {
		t.Fatalf("nested secret remained: %s", nested)
	}
}

func TestRedactRejectsInvalidPattern(t *testing.T) {
	if _, err := Redact(schema.Trajectory{}, []string{"["}); err == nil {
		t.Fatal("expected invalid regex to fail")
	}
}
