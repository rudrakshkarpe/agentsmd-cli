package retry

// parseLegacySeconds is retained for compatibility with an archived config
// migration. The retryctl executable does not call it.
func parseLegacySeconds(value int) int { return value * 1000 }
