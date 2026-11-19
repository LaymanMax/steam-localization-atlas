package normalization

import (
	"strings"
)

// GenreInfo holds normalized genre data.
type GenreInfo struct {
	ID          string // normalized slug, e.g. "action", "role_playing_rpg"
	DisplayName string // original clean name, e.g. "Action", "RPG"
	RawLabel    string
}

// ParseGenreEntry normalizes a single genre label.
func ParseGenreEntry(raw string) (GenreInfo, bool) {
	trimmed := strings.Trim(strings.TrimSpace(raw), "[] \t\r\n'\"")
	if trimmed == "" {
		return GenreInfo{}, false
	}

	id := toSnakeCase(trimmed)
	if id == "" {
		return GenreInfo{}, false
	}

	return GenreInfo{
		ID:          id,
		DisplayName: trimmed,
		RawLabel:    raw,
	}, true
}

// ParseGenresList normalizes and deduplicates a slice of genres for a game.
func ParseGenresList(genres []string) []GenreInfo {
	if len(genres) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var result []GenreInfo

	for _, g := range genres {
		cleaned := strings.Trim(g, "[]")
		for _, part := range strings.Split(cleaned, ",") {
			info, ok := ParseGenreEntry(part)
			if !ok {
				continue
			}
			if !seen[info.ID] {
				seen[info.ID] = true
				result = append(result, info)
			}
		}
	}

	return result
}
