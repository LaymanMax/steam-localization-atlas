package ingestion

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/laymanmax/steam-localization-atlas/internal/database"
	_ "github.com/marcboeker/go-duckdb"
)

func createTestParquet(t *testing.T, destPath string) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatalf("failed to open duckdb: %v", err)
	}
	defer db.Close()

	// Create test games table with various edge cases:
	// 1: Normal RPG with English & French text, English audio
	// 2: Strategy with English text & audio, German text
	// 3: Playtest title with zero owners and no languages
	// 4: Missing genres
	// 5: Duplicate AppID (handled/skipped)
	// 6: Audio without text (audio: Italian, text: English)
	createSQL := `
		CREATE TABLE test_source (
			appID VARCHAR,
			name VARCHAR,
			release_date VARCHAR,
			estimated_owners VARCHAR,
			price DOUBLE,
			windows BOOLEAN,
			mac BOOLEAN,
			linux BOOLEAN,
			metacritic_score BIGINT,
			user_score BIGINT,
			positive BIGINT,
			negative BIGINT,
			header_image VARCHAR,
			developers VARCHAR[],
			publishers VARCHAR[],
			supported_languages VARCHAR[],
			full_audio_languages VARCHAR[],
			genres VARCHAR[]
		);

		INSERT INTO test_source VALUES
		('101', 'Kingdom Quest', '2022-05-10', '20,000 - 50,000', 19.99, true, true, false, 85, 80, 500, 50, 'http://img1', ['Dev1'], ['Pub1'], ['English', 'French'], ['English'], ['RPG', 'Action']),
		('102', 'Empire Builder', '2021-03-15', '50,000 - 100,000', 29.99, true, false, true, 78, 75, 1200, 150, 'http://img2', ['Dev2'], ['Pub2'], ['English', 'German'], ['English'], ['Strategy']),
		('103', 'Space Fighter Playtest', '2023-01-01', '0 - 0', 0.0, true, false, false, 0, 0, 0, 0, '', [], [], [], [], []),
		('104', 'Mystery Mansion', '2020-10-31', '0 - 20000', 4.99, true, true, true, 60, 65, 100, 20, 'http://img4', ['Dev3'], ['Pub3'], ['English', 'Spanish - Spain'], ['English', 'Italian'], []),
		('101', 'Duplicate Title', '2022-05-10', '0 - 20000', 9.99, true, false, false, 0, 0, 0, 0, '', [], [], ['English'], [], ['Indie']);
	`
	if _, err := db.Exec(createSQL); err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	copySQL := fmt.Sprintf("COPY test_source TO '%s' (FORMAT PARQUET)", destPath)
	if _, err := db.Exec(copySQL); err != nil {
		t.Fatalf("failed to write parquet: %v", err)
	}
}

func TestPipelineIngest(t *testing.T) {
	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test_games.parquet")
	createTestParquet(t, parquetPath)

	dbPath := filepath.Join(tempDir, "test.duckdb")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	pipeline := NewPipeline(db)

	stats, err := pipeline.IngestFile(parquetPath, "test", nil)
	if err != nil {
		t.Fatalf("ingestion failed: %v", err)
	}

	if stats.TotalRecords != 4 { // 5 rows inserted, 1 duplicate skipped -> 4 records
		t.Errorf("expected 4 total records, got %d", stats.TotalRecords)
	}
	if stats.SuspectedPlaytests != 1 {
		t.Errorf("expected 1 suspected playtest, got %d", stats.SuspectedPlaytests)
	}
	if stats.GamesWithLanguages != 3 {
		t.Errorf("expected 3 games with languages, got %d", stats.GamesWithLanguages)
	}

	// Verify games table in database
	var count int
	if err := db.Conn().QueryRow("SELECT count(*) FROM games").Scan(&count); err != nil {
		t.Fatalf("failed to query games: %v", err)
	}
	if count != 4 {
		t.Errorf("expected 4 games in db, got %d", count)
	}

	// Verify languages table
	var langCount int
	if err := db.Conn().QueryRow("SELECT count(*) FROM languages").Scan(&langCount); err != nil {
		t.Fatalf("failed to query languages: %v", err)
	}
	if langCount == 0 {
		t.Errorf("expected languages to be populated")
	}

	// Verify atomic preservation on failed import:
	// If a broken file is ingested, previous games table must remain intact!
	brokenParquet := filepath.Join(tempDir, "broken.parquet")
	os.WriteFile(brokenParquet, []byte("not a parquet file"), 0644)

	_, err = pipeline.IngestFile(brokenParquet, "broken", nil)
	if err == nil {
		t.Errorf("expected error for broken parquet, got nil")
	}

	// Previous games count must still be 4
	var countAfter int
	if err := db.Conn().QueryRow("SELECT count(*) FROM games").Scan(&countAfter); err != nil {
		t.Fatalf("failed to query games after failed ingest: %v", err)
	}
	if countAfter != 4 {
		t.Errorf("expected 4 games preserved after failed import, got %d", countAfter)
	}
}
