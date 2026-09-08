package retry

import (
	"fmt"
	"os"
	"time"
)

const defaultDelay = 2 * time.Second

func Delay() (time.Duration, error) {
	raw, ok := os.LookupEnv("RETRY_DELAY_MS")
	if !ok {
		return defaultDelay, nil
	}
	delay, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse RETRY_DELAY_MS: %w", err)
	}
	if delay < 0 {
		return 0, fmt.Errorf("RETRY_DELAY_MS must not be negative")
	}
	return delay, nil
}
