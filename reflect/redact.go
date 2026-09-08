package reflect

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

const redacted = "[REDACTED]"

// Redacting keeps the stored trajectory untouched and passes a scrubbed copy
// to an external reflector.
type Redacting struct {
	Next     Reflector
	Patterns []string
}

func (r Redacting) Reflect(ctx context.Context, trajectory schema.Trajectory) (Result, error) {
	if r.Next == nil {
		return Result{}, fmt.Errorf("redacting reflector requires a next reflector")
	}
	clean, err := Redact(trajectory, r.Patterns)
	if err != nil {
		return Result{}, err
	}
	return r.Next.Reflect(ctx, clean)
}

func Redact(trajectory schema.Trajectory, patterns []string) (schema.Trajectory, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if pattern == "" {
			return schema.Trajectory{}, fmt.Errorf("redaction pattern cannot be empty")
		}
		value, err := regexp.Compile(pattern)
		if err != nil {
			return schema.Trajectory{}, fmt.Errorf("compile redaction pattern %q: %w", pattern, err)
		}
		compiled = append(compiled, value)
	}
	data, err := json.Marshal(trajectory)
	if err != nil {
		return schema.Trajectory{}, err
	}
	var clean schema.Trajectory
	if err := json.Unmarshal(data, &clean); err != nil {
		return schema.Trajectory{}, err
	}
	replace := func(value string) string {
		for _, pattern := range compiled {
			value = pattern.ReplaceAllString(value, redacted)
		}
		return value
	}
	clean.SessionID = replace(clean.SessionID)
	clean.Task = replace(clean.Task)
	for index := range clean.Steps {
		clean.Steps[index].Summary = replace(clean.Steps[index].Summary)
	}
	for index := range clean.ToolCalls {
		clean.ToolCalls[index].Result = replace(clean.ToolCalls[index].Result)
		clean.ToolCalls[index].Args = redactMap(clean.ToolCalls[index].Args, replace)
	}
	for index := range clean.Files {
		clean.Files[index].Path = replace(clean.Files[index].Path)
		clean.Files[index].Diff = replace(clean.Files[index].Diff)
	}
	for index := range clean.Commands {
		for argument := range clean.Commands[index].Argv {
			clean.Commands[index].Argv[argument] = replace(clean.Commands[index].Argv[argument])
		}
	}
	clean.FinalDiff = replace(clean.FinalDiff)
	for key, value := range clean.Metadata {
		clean.Metadata[key] = replace(value)
	}
	return clean, nil
}

func redactMap(input map[string]any, replace func(string) string) map[string]any {
	for key, value := range input {
		input[key] = redactValue(value, replace)
	}
	return input
}

func redactValue(value any, replace func(string) string) any {
	switch typed := value.(type) {
	case string:
		return replace(typed)
	case map[string]any:
		return redactMap(typed, replace)
	case []any:
		for index := range typed {
			typed[index] = redactValue(typed[index], replace)
		}
		return typed
	default:
		return value
	}
}
