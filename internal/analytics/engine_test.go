package analytics

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/laymanmax/steam-localization-atlas/internal/database"
	"github.com/laymanmax/steam-localization-atlas/internal/repository"
	_ "github.com/marcboeker/go-duckdb"
)

func setupTestDB(t *testing.T) *database.DB {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "analytics_test.duckdb")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Insert test fixture:
	// 4 games:
	// Game 1 (app_id=1, name="Game One", year=2021, genre=action, RPG; lang: english text+audio, french text)
	// Game 2 (app_id=2, name="Game Two", year=2022, genre=action; lang: english text, german text+audio)
	// Game 3 (app_id=3, name="Game Three", year=2023, genre=RPG; lang: english text, french text+audio)
	// Game 4 (app_id=4, name="Game Four Playtest", year=2023, is_playtest=true, genre=action; lang: english text)
	// Game 5 (app_id=5, name="Game Five No Lang", year=2020, genre=action; has_usable_languages=false)
	setupSQL := `
		INSERT INTO games VALUES
		(1, 'Game One', '2021-01-01', 2021, '0 - 20,000', 0, 20000, 10000, 9.99, false, true, false, false, 80, 80, 100, 10, 'img1', 'Dev', 'Pub', 'English, French', 'English', 'Action, RPG', false, true, true, 2, 1, ''),
		(2, 'Game Two', '2022-01-01', 2022, '0 - 20,000', 0, 20000, 10000, 0.0, true, true, true, false, 75, 70, 200, 20, 'img2', 'Dev', 'Pub', 'English, German', 'German', 'Action', false, true, true, 2, 1, ''),
		(3, 'Game Three', '2023-01-01', 2023, '20,000 - 50,000', 20000, 50000, 35000, 19.99, false, true, false, false, 90, 85, 500, 30, 'img3', 'Dev', 'Pub', 'English, French', 'French', 'RPG', false, true, true, 2, 1, ''),
		(4, 'Game Four Playtest', '2023-01-01', 2023, '0 - 0', 0, 0, 0, 0.0, true, true, false, false, 0, 0, 0, 0, '', 'Dev', 'Pub', 'English', '', 'Action', true, true, true, 1, 0, 'suspected_playtest'),
		(5, 'Game Five No Lang', '2020-01-01', 2020, '0 - 20,000', 0, 20000, 10000, 4.99, false, true, false, false, 60, 60, 50, 5, '', 'Dev', 'Pub', '', '', 'Action', false, false, true, 0, 0, 'missing_languages');

		INSERT INTO genres VALUES
		('action', 'Action', 4, 3),
		('rpg', 'RPG', 2, 2);

		INSERT INTO languages VALUES
		('english', 'English', 4, 1),
		('french', 'French', 2, 1),
		('german', 'German', 1, 1);

		INSERT INTO game_genres VALUES
		(1, 'action', 'Action'), (1, 'rpg', 'RPG'),
		(2, 'action', 'Action'),
		(3, 'rpg', 'RPG'),
		(4, 'action', 'Action'),
		(5, 'action', 'Action');

		INSERT INTO game_languages VALUES
		(1, 'english', true, true, 'English'),
		(1, 'french', true, false, 'French'),
		(2, 'english', true, false, 'English'),
		(2, 'german', true, true, 'German'),
		(3, 'english', true, false, 'English'),
		(3, 'french', true, true, 'French'),
		(4, 'english', true, false, 'English');
	`

	if _, err := db.Conn().Exec(setupSQL); err != nil {
		t.Fatalf("failed to insert test fixture: %v", err)
	}

	return db
}

func TestAnalyticsOverview(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	engine := NewEngine(db)
	filters := &repository.FilterParams{
		ExcludePlaytests: true,
	}

	summary, err := engine.GetOverview(filters)
	if err != nil {
		t.Fatalf("GetOverview failed: %v", err)
	}

	// 5 games total minus 1 playtest = 4 games
	if summary.TotalGames != 4 {
		t.Errorf("expected 4 total games, got %d", summary.TotalGames)
	}

	// Games with usable languages: Game 1, 2, 3 = 3 games (Game 5 has no languages)
	if summary.GamesWithLanguages != 3 {
		t.Errorf("expected 3 games with languages, got %d", summary.GamesWithLanguages)
	}
}

func TestAnalyticsMatrix(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	engine := NewEngine(db)
	filters := &repository.FilterParams{
		ExcludePlaytests: true,
	}

	// Test Text Coverage
	resp, err := engine.GetMatrix("text", filters, "", nil, 1, 10, "overall")
	if err != nil {
		t.Fatalf("GetMatrix failed: %v", err)
	}

	if len(resp.Genres) == 0 {
		t.Fatalf("expected genres in matrix")
	}

	// In Action genre (eligible games = 2: Game 1, Game 2):
	// English is in Game 1 & 2 -> 2/2 = 100%
	// French is in Game 1 -> 1/2 = 50%
	// German is in Game 2 -> 1/2 = 50%
	foundFrench := false
	for _, l := range resp.Languages {
		if l.ID == "french" {
			foundFrench = true
			cell, ok := l.Cells["action"]
			if !ok {
				t.Fatalf("expected action cell for French")
			}
			if cell.Numerator != 1 || cell.Denominator != 2 || cell.Percentage != 50.0 {
				t.Errorf("expected French in Action to be 1/2 (50.0%%), got %d/%d (%.2f%%)",
					cell.Numerator, cell.Denominator, cell.Percentage)
			}
		}
	}
	if !foundFrench {
		t.Errorf("expected French in languages list")
	}

	// Test Audio Conversion Metric
	convResp, err := engine.GetMatrix("conversion", filters, "", nil, 1, 10, "overall")
	if err != nil {
		t.Fatalf("GetMatrix conversion failed: %v", err)
	}

	// In RPG genre for French: Game 1 (French text only), Game 3 (French text + audio).
	// Audio conversion = 1 audio / 2 text = 50%
	for _, l := range convResp.Languages {
		if l.ID == "french" {
			cell := l.Cells["rpg"]
			if cell.Numerator != 1 || cell.Denominator != 2 || cell.Percentage != 50.0 {
				t.Errorf("expected French RPG audio conversion to be 1/2 (50.0%%), got %d/%d (%.2f%%)",
					cell.Numerator, cell.Denominator, cell.Percentage)
			}
		}
	}
}

func TestCompareGenres(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	engine := NewEngine(db)
	filters := &repository.FilterParams{
		ExcludePlaytests: true,
	}

	comp, err := engine.CompareGenres("action", "rpg", filters)
	if err != nil {
		t.Fatalf("CompareGenres failed: %v", err)
	}

	// Action eligible: 2 games (1, 2)
	// RPG eligible: 2 games (1, 3)
	// German in Action: 1/2 (50%), in RPG: 0/2 (0%) -> Difference: +50 percentage points!
	var foundGerman bool
	for _, item := range comp.Comparison {
		if item.LanguageID == "german" {
			foundGerman = true
			if item.CoverageA != 50.0 || item.CoverageB != 0.0 || item.DifferencePt != 50.0 {
				t.Errorf("expected German delta +50.0, got A=%.1f, B=%.1f, diff=%.1f",
					item.CoverageA, item.CoverageB, item.DifferencePt)
			}
		}
	}
	if !foundGerman {
		t.Errorf("expected German in comparison")
	}
}

// Silence unused package warning
var _ = sql.ErrNoRows
