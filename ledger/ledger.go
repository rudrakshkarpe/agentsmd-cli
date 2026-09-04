// Package ledger implements rule operations and AGENTS.md rendering.
package ledger

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/schema"
)

const DuplicateThreshold = 0.6

var wordPattern = regexp.MustCompile(`[a-z0-9]+`)

func Overlap(a, b string) float64 {
	left, right := words(a), words(b)
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	intersection := 0
	union := make(map[string]struct{}, len(left)+len(right))
	for word := range left {
		union[word] = struct{}{}
		if _, ok := right[word]; ok {
			intersection++
		}
	}
	for word := range right {
		union[word] = struct{}{}
	}
	return float64(intersection) / float64(len(union))
}

func words(value string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, word := range wordPattern.FindAllString(strings.ToLower(value), -1) {
		result[word] = struct{}{}
	}
	return result
}

func FindDuplicate(value schema.Ledger, text string) *schema.Rule {
	for i := range value.Rules {
		if Overlap(value.Rules[i].Text, text) >= DuplicateThreshold {
			return &value.Rules[i]
		}
	}
	return nil
}

func Add(value *schema.Ledger, text string, origin schema.Origin) (*schema.Rule, *schema.Rule, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil, fmt.Errorf("rule text cannot be empty")
	}
	if duplicate := FindDuplicate(*value, text); duplicate != nil {
		return nil, duplicate, nil
	}
	rule := schema.Rule{
		ID:     nextID(value.Rules),
		Text:   text,
		Origin: origin,
		Born:   time.Now().UTC().Format(time.DateOnly),
	}
	value.Rules = append(value.Rules, rule)
	return &value.Rules[len(value.Rules)-1], nil, nil
}

func nextID(rules []schema.Rule) string {
	max := -1
	for _, rule := range rules {
		n, err := strconv.Atoi(strings.TrimPrefix(rule.ID, "r"))
		if err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("r%03d", max+1)
}

func Render(value schema.Ledger) string {
	var output strings.Builder
	output.WriteString("# AGENTS.md\n\n## Lessons\n")
	output.WriteString("<!-- managed by agentsmd; agents read these first -->\n\n")
	for _, rule := range value.Rules {
		fmt.Fprintf(&output, "- [%s] %s  (cited: %d)\n", rule.ID, rule.Text, rule.Cited)
	}
	return output.String()
}

type Issue struct {
	Kind    string
	RuleID  string
	OtherID string
	Message string
}

func Lint(value schema.Ledger) []Issue {
	issues := []Issue{}
	for i, rule := range value.Rules {
		if rule.Cited == 0 {
			issues = append(issues, Issue{Kind: "unused", RuleID: rule.ID, Message: "rule has never fired"})
		}
		for j := 0; j < i; j++ {
			if Overlap(rule.Text, value.Rules[j].Text) >= DuplicateThreshold {
				issues = append(issues, Issue{Kind: "duplicate", RuleID: rule.ID, OtherID: value.Rules[j].ID, Message: "rules overlap"})
			}
		}
	}
	sort.SliceStable(issues, func(i, j int) bool { return issues[i].RuleID < issues[j].RuleID })
	return issues
}
