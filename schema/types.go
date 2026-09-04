// Package schema defines the implementation-independent agentsmd data model.
package schema

import "time"

type Trajectory struct {
	SessionID   string            `json:"session_id"`
	Tool        string            `json:"tool"`
	Task        string            `json:"task,omitempty"`
	Steps       []Step            `json:"steps"`
	ToolCalls   []ToolCall        `json:"tool_calls"`
	Files       []FileTouch       `json:"files_touched"`
	Commands    []Command         `json:"commands"`
	TestResults TestResults       `json:"test_results"`
	Tokens      Tokens            `json:"tokens"`
	WallTimeS   float64           `json:"wall_time_s"`
	FinalDiff   string            `json:"final_diff"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type Step struct {
	Role    string `json:"role"`
	Summary string `json:"summary"`
}

type ToolCall struct {
	Name   string         `json:"name"`
	Args   map[string]any `json:"args"`
	Result string         `json:"result"`
}

type FileTouch struct {
	Path string `json:"path"`
	Diff string `json:"diff"`
}

type Command struct {
	Argv     []string `json:"argv"`
	ExitCode int      `json:"exit_code"`
}

type TestResults struct {
	Passed int `json:"passed"`
	Failed int `json:"failed"`
	Errors int `json:"errors"`
}

type Tokens struct {
	Input  int `json:"in"`
	Output int `json:"out"`
	Cached int `json:"cached"`
}

func (t Tokens) Total() int { return t.Input + t.Output }

type Origin struct {
	Run     string `json:"run,omitempty"`
	Task    string `json:"task,omitempty"`
	Version string `json:"version,omitempty"`
}

type Rule struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Origin Origin `json:"origin"`
	Cited  int    `json:"cited"`
	Born   string `json:"born"`
}

type Ledger struct {
	Rules []Rule           `json:"rules"`
	Runs  map[string][]int `json:"runs"`
}

type Version struct {
	ID      string         `json:"id"`
	Parent  *string        `json:"parent"`
	Time    time.Time      `json:"ts"`
	Reason  string         `json:"reason"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta"`
}

type Proposal struct {
	ID       string    `json:"id"`
	Text     string    `json:"text"`
	Origin   Origin    `json:"origin"`
	Proposed time.Time `json:"proposed"`
}
