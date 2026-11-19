package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/laymanmax/steam-localization-atlas/internal/database"
	_ "github.com/marcboeker/go-duckdb"
)

func setupTestServer(t *testing.T) *Server {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "api_test.duckdb")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Populate fixture
	setupSQL := `
		INSERT INTO games VALUES
		(1, 'Cyber Realm', '2022-01-01', 2022, '20,000 - 50,000', 20000, 50000, 35000, 29.99, false, true, true, false, 85, 80, 500, 50, 'http://img1', 'Dev1', 'Pub1', 'English, French, German', 'English', 'Action, RPG', false, true, true, 3, 1, ''),
		(2, 'Solar Tactics', '2023-05-10', 2023, '50,000 - 100,000', 50000, 100000, 75000, 14.99, false, true, false, true, 90, 88, 1200, 100, 'http://img2', 'Dev2', 'Pub2', 'English, Japanese', 'English, Japanese', 'Strategy', false, true, true, 2, 2, ''),
		(3, 'Empty Game', '2020-01-01', 2020, '0 - 0', 0, 0, 0, 0.0, true, true, false, false, 0, 0, 0, 0, '', 'Dev3', 'Pub3', '', '', '', false, false, false, 0, 0, 'missing_languages,missing_genres');

		INSERT INTO genres VALUES
		('action', 'Action', 1, 1),
		('rpg', 'RPG', 1, 1),
		('strategy', 'Strategy', 1, 1);

		INSERT INTO languages VALUES
		('english', 'English', 2, 2),
		('french', 'French', 1, 0),
		('german', 'German', 1, 0),
		('japanese', 'Japanese', 1, 1);

		INSERT INTO game_genres VALUES
		(1, 'action', 'Action'), (1, 'rpg', 'RPG'),
		(2, 'strategy', 'Strategy');

		INSERT INTO game_languages VALUES
		(1, 'english', true, true, 'English'),
		(1, 'french', true, false, 'French'),
		(1, 'german', true, false, 'German'),
		(2, 'english', true, true, 'English'),
		(2, 'japanese', true, true, 'Japanese');

		INSERT INTO data_quality_issues VALUES
		('iss-1', 'missing_languages', 'warning', 3, 'Empty Game', 'No languages'),
		('iss-2', 'missing_genres', 'warning', 3, 'Empty Game', 'No genres');
	`
	if _, err := db.Conn().Exec(setupSQL); err != nil {
		t.Fatalf("failed to insert api test fixture: %v", err)
	}

	return NewServer(db, tempDir, "*")
}

func TestHealthEndpoint(t *testing.T) {
	s := setupTestServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("expected healthy, got %v", resp["status"])
	}
	if resp["hasData"] != true {
		t.Errorf("expected hasData=true, got %v", resp["hasData"])
	}
}

func TestOverviewEndpoint(t *testing.T) {
	s := setupTestServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/v1/overview", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["totalGames"].(float64) != 3 {
		t.Errorf("expected 3 games, got %v", resp["totalGames"])
	}
	if resp["gamesWithLanguages"].(float64) != 2 {
		t.Errorf("expected 2 games with languages, got %v", resp["gamesWithLanguages"])
	}
}

func TestLocalizationMatrixEndpoint(t *testing.T) {
	s := setupTestServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/v1/localization/matrix?metric=text&minSample=1", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode matrix response: %v", err)
	}

	if resp["metric"] != "text" {
		t.Errorf("expected metric text, got %v", resp["metric"])
	}

	genres := resp["genres"].([]any)
	if len(genres) == 0 {
		t.Fatalf("expected genres in matrix")
	}

	langs := resp["languages"].([]any)
	if len(langs) == 0 {
		t.Fatalf("expected languages in matrix")
	}
}

func TestExportMatrixEndpoint(t *testing.T) {
	s := setupTestServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/v1/exports/matrix?metric=text&minSample=1", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/csv; charset=utf-8" {
		t.Errorf("expected CSV content type, got %s", contentType)
	}

	body := rr.Body.String()
	if len(body) == 0 {
		t.Errorf("expected non-empty CSV body")
	}
}

func TestGameDetailEndpoint(t *testing.T) {
	s := setupTestServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/v1/games/1", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var g map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&g); err != nil {
		t.Fatalf("failed to decode game: %v", err)
	}

	if g["name"] != "Cyber Realm" {
		t.Errorf("expected 'Cyber Realm', got %v", g["name"])
	}
}

func TestQualitySummaryEndpoint(t *testing.T) {
	s := setupTestServer(t)
	defer s.db.Close()

	req := httptest.NewRequest("GET", "/api/v1/quality/summary", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var q map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("failed to decode quality summary: %v", err)
	}

	if q["missingLanguages"].(float64) != 1 {
		t.Errorf("expected 1 missing languages, got %v", q["missingLanguages"])
	}
}
