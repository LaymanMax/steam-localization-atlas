package normalization

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	yearRegex   = regexp.MustCompile(`\b(19\d\d|20\d\d)\b`)
	dateFormats = []string{
		"Jan 2, 2006",
		"January 2, 2006",
		"2 Jan, 2006",
		"2006-01-02",
		"Jan 2006",
		"2006",
	}
)

// ParseReleaseDate extracts the year and formatted date if possible.
func ParseReleaseDate(raw string) (year int, formatted string, isValid bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, "", false
	}

	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, trimmed); err == nil {
			return t.Year(), t.Format("2006-01-02"), true
		}
	}

	// Fallback: search for 4-digit year (1970 - 2035)
	if match := yearRegex.FindString(trimmed); match != "" {
		if y, err := strconv.Atoi(match); err == nil {
			return y, match, true
		}
	}

	return 0, trimmed, false
}
