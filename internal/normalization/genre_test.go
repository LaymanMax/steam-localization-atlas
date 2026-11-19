package normalization

import (
	"testing"
)

func TestParseGenresList(t *testing.T) {
	input := []string{"Action", "Adventure", "Action", "Strategy, RPG"}
	results := ParseGenresList(input)

	if len(results) != 4 {
		t.Fatalf("expected 4 genres, got %d", len(results))
	}

	expectedIDs := []string{"action", "adventure", "strategy", "rpg"}
	for i, exp := range expectedIDs {
		if results[i].ID != exp {
			t.Errorf("index %d: expected %q, got %q", i, exp, results[i].ID)
		}
	}
}
