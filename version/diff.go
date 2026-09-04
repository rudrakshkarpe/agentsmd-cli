package version

import (
	"fmt"
	"strings"
)

// Diff returns a compact line diff suitable for CLI review.
func (s *Store) Diff(from, to string) (string, error) {
	leftData, err := s.Content(from)
	if err != nil {
		return "", err
	}
	rightData, err := s.Content(to)
	if err != nil {
		return "", err
	}
	left := strings.Split(strings.TrimSuffix(string(leftData), "\n"), "\n")
	right := strings.Split(strings.TrimSuffix(string(rightData), "\n"), "\n")
	operations := lineDiff(left, right)
	var output strings.Builder
	fmt.Fprintf(&output, "--- %s\n+++ %s\n", from, to)
	for _, operation := range operations {
		fmt.Fprintf(&output, "%c%s\n", operation.kind, operation.line)
	}
	return output.String(), nil
}

type diffLine struct {
	kind rune
	line string
}

func lineDiff(left, right []string) []diffLine {
	table := make([][]int, len(left)+1)
	for i := range table {
		table[i] = make([]int, len(right)+1)
	}
	for i := len(left) - 1; i >= 0; i-- {
		for j := len(right) - 1; j >= 0; j-- {
			if left[i] == right[j] {
				table[i][j] = table[i+1][j+1] + 1
			} else if table[i+1][j] >= table[i][j+1] {
				table[i][j] = table[i+1][j]
			} else {
				table[i][j] = table[i][j+1]
			}
		}
	}
	result := []diffLine{}
	for i, j := 0, 0; i < len(left) || j < len(right); {
		switch {
		case i < len(left) && j < len(right) && left[i] == right[j]:
			result = append(result, diffLine{' ', left[i]})
			i++
			j++
		case j < len(right) && (i == len(left) || table[i][j+1] > table[i+1][j]):
			result = append(result, diffLine{'+', right[j]})
			j++
		default:
			result = append(result, diffLine{'-', left[i]})
			i++
		}
	}
	return result
}
