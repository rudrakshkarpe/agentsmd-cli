package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvironmentPrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"region":"ap-northeast-1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	lookup := func(string) (string, bool) { return "eu-west-1", true }
	values, err := Load(path, lookup)
	if err != nil {
		t.Fatal(err)
	}
	if values.Region != "eu-west-1" {
		t.Fatalf("region=%q", values.Region)
	}
}
