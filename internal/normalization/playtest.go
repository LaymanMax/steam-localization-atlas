package normalization

import (
	"strings"
)

// IsSuspectedPlaytest identifies likely playtests using explicit title cues.
func IsSuspectedPlaytest(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "playtest") ||
		strings.Contains(lower, "play test") ||
		strings.Contains(lower, "closed beta") ||
		strings.Contains(lower, "open beta") ||
		strings.Contains(lower, "beta test") ||
		strings.Contains(lower, "server test") ||
		strings.Contains(lower, "stress test")
}
