package repository

import (
	"fmt"
	"strings"
)

// FilterParams defines the global filtering parameters applicable to analytical queries.
type FilterParams struct {
	ReleaseYearMin   int      `json:"releaseYearMin"`
	ReleaseYearMax   int      `json:"releaseYearMax"`
	Genres           []string `json:"genres"`
	Ownership        []string `json:"ownership"`
	IsFree           *bool    `json:"isFree"`
	Platform         string   `json:"platform"` // "windows", "mac", "linux"
	ExcludePlaytests bool     `json:"excludePlaytests"`
	RequireLanguages bool     `json:"requireLanguages"`
	Search           string   `json:"search"`
}

// BuildWhereClause generates SQL conditions and parameter values for the games table.
// alias is the table alias for games, e.g. "g" or "".
func (f *FilterParams) BuildWhereClause(alias string) (string, []any) {
	var conds []string
	var args []any

	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}

	if f.RequireLanguages {
		conds = append(conds, prefix+"has_usable_languages = true")
	}

	if f.ExcludePlaytests {
		conds = append(conds, prefix+"is_playtest = false")
	}

	if f.ReleaseYearMin > 0 {
		conds = append(conds, prefix+"release_year >= ?")
		args = append(args, f.ReleaseYearMin)
	}

	if f.ReleaseYearMax > 0 {
		conds = append(conds, prefix+"release_year <= ?")
		args = append(args, f.ReleaseYearMax)
	}

	if len(f.Ownership) > 0 {
		placeholders := make([]string, len(f.Ownership))
		for i, o := range f.Ownership {
			placeholders[i] = "?"
			args = append(args, o)
		}
		conds = append(conds, fmt.Sprintf("%sestimated_owners IN (%s)", prefix, strings.Join(placeholders, ",")))
	}

	if f.IsFree != nil {
		conds = append(conds, prefix+"is_free = ?")
		args = append(args, *f.IsFree)
	}

	if f.Platform != "" {
		switch strings.ToLower(f.Platform) {
		case "windows":
			conds = append(conds, prefix+"windows = true")
		case "mac":
			conds = append(conds, prefix+"mac = true")
		case "linux":
			conds = append(conds, prefix+"linux = true")
		}
	}

	if f.Search != "" {
		conds = append(conds, fmt.Sprintf("LOWER(%sname) LIKE ?", prefix))
		args = append(args, "%"+strings.ToLower(f.Search)+"%")
	}

	if len(conds) == 0 {
		return "1=1", nil
	}

	return strings.Join(conds, " AND "), args
}
