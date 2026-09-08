package retry

import (
	"os"
	"testing"
	"time"
)

func TestDelayUsesIntegerMilliseconds(t *testing.T) {
	t.Setenv("RETRY_DELAY_MS", "250")
	delay, err := Delay()
	if err != nil || delay != 250*time.Millisecond {
		t.Fatalf("delay=%s err=%v", delay, err)
	}
}

func TestDelayPreservesDefaultAndValidation(t *testing.T) {
	t.Setenv("RETRY_DELAY_MS", "temporary")
	if err := os.Unsetenv("RETRY_DELAY_MS"); err != nil {
		t.Fatal(err)
	}
	if delay, err := Delay(); err != nil || delay != 2*time.Second {
		t.Fatalf("default delay=%s err=%v", delay, err)
	}
	t.Setenv("RETRY_DELAY_MS", "")
	if _, err := Delay(); err == nil {
		t.Fatal("empty RETRY_DELAY_MS must be rejected")
	}
	t.Setenv("RETRY_DELAY_MS", "-1")
	if _, err := Delay(); err == nil {
		t.Fatal("negative RETRY_DELAY_MS must be rejected")
	}
}
