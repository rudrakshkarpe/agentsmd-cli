// Package claude normalizes Claude Code hook transcripts.
package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rudrakshkarpe/agentsmd-cli/capture"
	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

type HookEvent struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
}

type Adapter struct {
	Event HookEvent
}

func (Adapter) Name() string { return "claude" }

func (Adapter) Capabilities() capture.Capabilities {
	return capture.Capabilities{Hooks: true, Transcript: "jsonl", Events: []string{"Stop", "SessionEnd", "PostToolUse"}}
}

func (a Adapter) Latest(ctx context.Context) (*schema.Trajectory, error) {
	if a.Event.TranscriptPath == "" {
		return nil, fmt.Errorf("Claude hook event has no transcript_path")
	}
	file, err := os.Open(a.Event.TranscriptPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return Normalize(ctx, file, a.Event)
}

type envelope struct {
	Type    string  `json:"type"`
	Message message `json:"message"`
}

type message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Usage   usage           `json:"usage"`
}

type usage struct {
	Input        int `json:"input_tokens"`
	Output       int `json:"output_tokens"`
	CacheRead    int `json:"cache_read_input_tokens"`
	CacheCreated int `json:"cache_creation_input_tokens"`
}

type block struct {
	Type      string         `json:"type"`
	Text      string         `json:"text"`
	Name      string         `json:"name"`
	ID        string         `json:"id"`
	ToolUseID string         `json:"tool_use_id"`
	Input     map[string]any `json:"input"`
	Content   any            `json:"content"`
}

func Normalize(ctx context.Context, source io.Reader, event HookEvent) (*schema.Trajectory, error) {
	result := &schema.Trajectory{
		SessionID: event.SessionID,
		Tool:      "claude",
		Steps:     []schema.Step{},
		ToolCalls: []schema.ToolCall{},
		Files:     []schema.FileTouch{},
		Commands:  []schema.Command{},
		Metadata:  map[string]string{"transcript_path": event.TranscriptPath},
	}
	toolIndexes := map[string]int{}
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var item envelope
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			return nil, fmt.Errorf("decode Claude transcript line: %w", err)
		}
		result.Tokens.Input += item.Message.Usage.Input
		result.Tokens.Output += item.Message.Usage.Output
		result.Tokens.Cached += item.Message.Usage.CacheRead + item.Message.Usage.CacheCreated
		blocks, err := contentBlocks(item.Message.Content)
		if err != nil {
			return nil, err
		}
		for _, value := range blocks {
			switch value.Type {
			case "text":
				if text := strings.TrimSpace(value.Text); text != "" && (item.Type == "assistant" || item.Message.Role == "assistant") {
					result.Steps = append(result.Steps, schema.Step{Role: "assistant", Summary: text})
				}
			case "tool_use":
				call := schema.ToolCall{Name: value.Name, Args: value.Input}
				result.ToolCalls = append(result.ToolCalls, call)
				toolIndexes[value.ID] = len(result.ToolCalls) - 1
				if value.Name == "Bash" || value.Name == "Shell" {
					if command, ok := value.Input["command"].(string); ok {
						result.Commands = append(result.Commands, schema.Command{Argv: strings.Fields(command), ExitCode: -1})
					}
				}
			case "tool_result":
				if index, ok := toolIndexes[value.ToolUseID]; ok {
					result.ToolCalls[index].Result = stringify(value.Content)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func contentBlocks(raw json.RawMessage) ([]block, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return []block{{Type: "text", Text: text}}, nil
	}
	var values []block
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("decode Claude message content: %w", err)
	}
	return values, nil
}

func stringify(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	data, _ := json.Marshal(value)
	return string(data)
}
