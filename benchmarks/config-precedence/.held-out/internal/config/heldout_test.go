package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHeldOutEmptyEnvironmentOverridesFile(t *testing.T) {
	path := heldOutConfig(t)
	values, err := Load(path, func(string) (string, bool) { return "", true })
	if err != nil {
		t.Fatal(err)
	}
	if values.Region != "" {
		t.Fatalf("explicit empty APP_REGION must clear file value, got %q", values.Region)
	}
}

func TestHeldOutUnsetEnvironmentPreservesFile(t *testing.T) {
	path := heldOutConfig(t)
	values, err := Load(path, func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatal(err)
	}
	if values.Region != "ap-northeast-1" {
		t.Fatalf("unset APP_REGION must preserve file value, got %q", values.Region)
	}
}

func heldOutConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"region":"ap-northeast-1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
