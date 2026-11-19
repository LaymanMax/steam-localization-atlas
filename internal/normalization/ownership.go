package normalization

import (
	"regexp"
	"strconv"
	"strings"
)

// OwnershipInfo holds normalized ownership data.
type OwnershipInfo struct {
	OriginalBracket string // e.g. "0 - 20000" or "20,000 - 50,000"
	NormalizedRange string // standardized bracket label e.g. "0 - 20,000"
	MinOwners       int64  // lower bound
	MaxOwners       int64  // upper bound
	MidOwners       int64  // midpoint for approximate sorting
	IsValid         bool
}

var ownerRangeRegex = regexp.MustCompile(`^([\d,\.]+)\s*-\s*([\d,\.]+)$`)

// ParseEstimatedOwners parses an estimated owners string into structured bounds.
func ParseEstimatedOwners(raw string) OwnershipInfo {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return OwnershipInfo{
			OriginalBracket: raw,
			NormalizedRange: "Unknown",
			IsValid:         false,
		}
	}

	matches := ownerRangeRegex.FindStringSubmatch(trimmed)
	if len(matches) != 3 {
		return OwnershipInfo{
			OriginalBracket: raw,
			NormalizedRange: trimmed,
			IsValid:         false,
		}
	}

	minStr := strings.ReplaceAll(strings.ReplaceAll(matches[1], ",", ""), ".", "")
	maxStr := strings.ReplaceAll(strings.ReplaceAll(matches[2], ",", ""), ".", "")

	minVal, err1 := strconv.ParseInt(minStr, 10, 64)
	maxVal, err2 := strconv.ParseInt(maxStr, 10, 64)

	if err1 != nil || err2 != nil {
		return OwnershipInfo{
			OriginalBracket: raw,
			NormalizedRange: trimmed,
			IsValid:         false,
		}
	}

	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	midVal := (minVal + maxVal) / 2

	// Format a clean normalized range label with commas
	normLabel := formatNumber(minVal) + " - " + formatNumber(maxVal)

	return OwnershipInfo{
		OriginalBracket: raw,
		NormalizedRange: normLabel,
		MinOwners:       minVal,
		MaxOwners:       maxVal,
		MidOwners:       midVal,
		IsValid:         true,
	}
}

func formatNumber(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}

	var res []byte
	l := len(s)
	for i := 0; i < l; i++ {
		if i > 0 && (l-i)%3 == 0 {
			res = append(res, ',')
		}
		res = append(res, s[i])
	}
	return string(res)
}
