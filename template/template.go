// Package template exposes embedded AGENTS.md starting points.
package template

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed files/*.md
var files embed.FS

func List() ([]string, error) {
	entries, err := fs.ReadDir(files, "files")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			names = append(names, strings.TrimSuffix(entry.Name(), ".md"))
		}
	}
	sort.Strings(names)
	return names, nil
}

func Load(name string) (string, error) {
	if strings.ContainsAny(name, `/\\`) || name == "" {
		return "", fmt.Errorf("invalid template name %q", name)
	}
	data, err := files.ReadFile("files/" + name + ".md")
	if err != nil {
		return "", fmt.Errorf("unknown template %q: %w", name, err)
	}
	return string(data), nil
}
