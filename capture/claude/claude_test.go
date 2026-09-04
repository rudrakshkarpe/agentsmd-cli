package claude_test

import (
	"context"
	"strings"
	"testing"

	"github.com/rudrakshkarpe/agentsmd-cli/capture/claude"
)

func TestNormalizeAssistantToolsResultsAndTokens(t *testing.T) {
	transcript := strings.Join([]string{
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"I will run the focused test."},{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"go test ./ledger"}}],"usage":{"input_tokens":100,"output_tokens":30,"cache_read_input_tokens":20}}}`,
		`{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"ok"}]}}`,
	}, "\n")
	result, err := claude.Normalize(context.Background(), strings.NewReader(transcript), claude.HookEvent{SessionID: "session-1", TranscriptPath: "/tmp/transcript.jsonl"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Tool != "claude" || result.SessionID != "session-1" {
		t.Fatalf("identity=%+v", result)
	}
	if len(result.ToolCalls) != 1 || result.ToolCalls[0].Result != "ok" {
		t.Fatalf("tool calls=%+v", result.ToolCalls)
	}
	if len(result.Commands) != 1 || result.Commands[0].Argv[0] != "go" {
		t.Fatalf("commands=%+v", result.Commands)
	}
	if result.Tokens.Input != 100 || result.Tokens.Output != 30 || result.Tokens.Cached != 20 {
		t.Fatalf("tokens=%+v", result.Tokens)
	}
}
