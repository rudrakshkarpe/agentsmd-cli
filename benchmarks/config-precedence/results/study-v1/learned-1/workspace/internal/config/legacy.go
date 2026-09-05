package config

import "os"

// LegacyRegion is retained for an old integration and is not used by envmerge.
func LegacyRegion(fileValue string) string {
	if value := os.Getenv("APP_REGION"); value != "" {
		return value
	}
	return fileValue
}
