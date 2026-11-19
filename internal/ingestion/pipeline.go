package ingestion

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/laymanmax/steam-localization-atlas/internal/database"
	"github.com/laymanmax/steam-localization-atlas/internal/logging"
	"github.com/laymanmax/steam-localization-atlas/internal/normalization"
)

// ImportStats captures summary metrics after an ingestion run.
type ImportStats struct {
	TotalRecords       int64         `json:"totalRecords"`
	GamesWithLanguages int64         `json:"gamesWithLanguages"`
	GamesWithAudio     int64         `json:"gamesWithAudio"`
	GamesWithGenres    int64         `json:"gamesWithGenres"`
	SuspectedPlaytests int64         `json:"suspectedPlaytests"`
	TotalIssuesLogged  int64         `json:"totalIssuesLogged"`
	DistinctLanguages  int64         `json:"distinctLanguages"`
	DistinctGenres     int64         `json:"distinctGenres"`
	Duration           time.Duration `json:"duration"`
	ImportID           string        `json:"importId"`
}

// IngestionProgressCallback is called periodically to report rows processed.
type IngestionProgressCallback func(stage string, processed, total int64)

// Pipeline manages the ingestion and normalization of Parquet files.
type Pipeline struct {
	db *database.DB
}

// NewPipeline creates a new ingestion pipeline.
func NewPipeline(db *database.DB) *Pipeline {
	return &Pipeline{db: db}
}

// InspectSchema inspects the Parquet schema and verifies required columns.
func (p *Pipeline) InspectSchema(parquetPath string) ([]string, error) {
	rows, err := p.db.Conn().Query(fmt.Sprintf("DESCRIBE SELECT * FROM read_parquet('%s')", parquetPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read parquet schema: %w", err)
	}
	defer rows.Close()

	var columns []string
	hasAppID := false

	for rows.Next() {
		var colName, colType, null, key, def, extra sql.NullString
		if err := rows.Scan(&colName, &colType, &null, &key, &def, &extra); err != nil {
			return nil, err
		}
		columns = append(columns, colName.String)
		if strings.EqualFold(colName.String, "appID") {
			hasAppID = true
		}
	}

	if !hasAppID {
		return nil, fmt.Errorf("required identifier column 'appID' missing from Parquet file")
	}

	return columns, nil
}

// IngestFile ingests and normalizes records from a Parquet file into DuckDB atomically.
func (p *Pipeline) IngestFile(parquetPath string, sourceType string, progress IngestionProgressCallback) (*ImportStats, error) {
	startTime := time.Now()
	importID := uuid.New().String()

	logging.Info("Starting ingestion pipeline for %s (Import ID: %s)", parquetPath, importID)

	// 1. Inspect schema
	if _, err := p.InspectSchema(parquetPath); err != nil {
		return nil, err
	}

	// 2. Count total rows in Parquet
	var totalRows int64
	err := p.db.Conn().QueryRow(fmt.Sprintf("SELECT count(*) FROM read_parquet('%s')", parquetPath)).Scan(&totalRows)
	if err != nil {
		return nil, fmt.Errorf("failed to count records in %s: %w", parquetPath, err)
	}
	logging.Info("Discovered %d records in Parquet", totalRows)

	// 3. Create staging tables
	if err := p.createStagingTables(); err != nil {
		return nil, fmt.Errorf("failed to create staging tables: %w", err)
	}
	defer p.dropStagingTables()

	// 4. Create temporary CSV/TSV files for bulk COPY
	tempDir, err := os.MkdirTemp("", "atlas_ingest_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	gamesFile, err := os.Create(filepath.Join(tempDir, "games.csv"))
	if err != nil {
		return nil, err
	}
	gameLangsFile, err := os.Create(filepath.Join(tempDir, "game_languages.csv"))
	if err != nil {
		return nil, err
	}
	gameGenresFile, err := os.Create(filepath.Join(tempDir, "game_genres.csv"))
	if err != nil {
		return nil, err
	}
	issuesFile, err := os.Create(filepath.Join(tempDir, "issues.csv"))
	if err != nil {
		return nil, err
	}

	gamesWriter := csv.NewWriter(gamesFile)
	gameLangsWriter := csv.NewWriter(gameLangsFile)
	gameGenresWriter := csv.NewWriter(gameGenresFile)
	issuesWriter := csv.NewWriter(issuesFile)

	// 5. Query parquet data using DuckDB
	if progress != nil {
		progress("Reading and normalizing records", 0, totalRows)
	}

	query := fmt.Sprintf(`
		SELECT 
			COALESCE(CAST(appID AS VARCHAR), ''),
			COALESCE(CAST(name AS VARCHAR), ''),
			COALESCE(CAST(release_date AS VARCHAR), ''),
			COALESCE(CAST(estimated_owners AS VARCHAR), ''),
			COALESCE(price, 0.0),
			COALESCE(windows, true),
			COALESCE(mac, false),
			COALESCE(linux, false),
			COALESCE(metacritic_score, 0),
			COALESCE(user_score, 0),
			COALESCE(positive, 0),
			COALESCE(negative, 0),
			COALESCE(CAST(header_image AS VARCHAR), ''),
			COALESCE(CAST(developers AS VARCHAR), ''),
			COALESCE(CAST(publishers AS VARCHAR), ''),
			COALESCE(CAST(supported_languages AS VARCHAR), '[]'),
			COALESCE(CAST(full_audio_languages AS VARCHAR), '[]'),
			COALESCE(CAST(genres AS VARCHAR), '[]')
		FROM read_parquet('%s')
	`, parquetPath)

	rows, err := p.db.Conn().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query parquet rows: %w", err)
	}
	defer rows.Close()

	var processedCount int64
	var withLangsCount, withAudioCount, withGenresCount, playtestsCount, issuesCount int64
	uniqueAppIDs := make(map[int64]bool)

	for rows.Next() {
		var (
			rawAppID, name, releaseDate, estOwners string
			price                                  float64
			windows, mac, linux                    bool
			metacritic, userScore, pos, neg        int64
			headerImg, dev, pub                    string
			supLangs, audioLangs, rawGenres        string
		)

		if err := rows.Scan(
			&rawAppID, &name, &releaseDate, &estOwners, &price,
			&windows, &mac, &linux, &metacritic, &userScore, &pos, &neg,
			&headerImg, &dev, &pub, &supLangs, &audioLangs, &rawGenres,
		); err != nil {
			logging.Warn("Error scanning parquet row: %v", err)
			continue
		}

		appID, _ := strconv.ParseInt(strings.TrimSpace(rawAppID), 10, 64)
		if appID == 0 {
			continue
		}

		// Quality check: duplicate AppID
		if uniqueAppIDs[appID] {
			issuesCount++
			_ = issuesWriter.Write([]string{
				uuid.New().String(), "duplicate_appid", "warning",
				strconv.FormatInt(appID, 10), name, "Duplicate App ID encountered in dataset",
			})
			continue
		}
		uniqueAppIDs[appID] = true

		// Normalization
		textLangs := normalization.ParseLanguagesFromString(supLangs)
		fullAudioLangs := normalization.ParseLanguagesFromString(audioLangs)
		genres := normalization.ParseGenresList([]string{rawGenres})
		ownership := normalization.ParseEstimatedOwners(estOwners)
		relYear, cleanRelDate, validDate := normalization.ParseReleaseDate(releaseDate)
		isPlaytest := normalization.IsSuspectedPlaytest(name)

		if isPlaytest {
			playtestsCount++
			issuesCount++
			_ = issuesWriter.Write([]string{
				uuid.New().String(), "suspected_playtest", "info",
				strconv.FormatInt(appID, 10), name, "Identified as suspected playtest/beta via title heuristics",
			})
		}

		hasUsableLangs := len(textLangs) > 0
		hasUsableGenres := len(genres) > 0

		if hasUsableLangs {
			withLangsCount++
		}
		if len(fullAudioLangs) > 0 {
			withAudioCount++
		}
		if hasUsableGenres {
			withGenresCount++
		}

		var qualityFlags []string
		if strings.TrimSpace(name) == "" {
			qualityFlags = append(qualityFlags, "missing_name")
			issuesCount++
			_ = issuesWriter.Write([]string{
				uuid.New().String(), "missing_name", "warning",
				strconv.FormatInt(appID, 10), name, "Game record has an empty name",
			})
		}
		if !hasUsableLangs {
			qualityFlags = append(qualityFlags, "missing_languages")
			issuesCount++
			_ = issuesWriter.Write([]string{
				uuid.New().String(), "missing_languages", "warning",
				strconv.FormatInt(appID, 10), name, "No usable supported language metadata",
			})
		}
		if !hasUsableGenres {
			qualityFlags = append(qualityFlags, "missing_genres")
			issuesCount++
			_ = issuesWriter.Write([]string{
				uuid.New().String(), "missing_genres", "warning",
				strconv.FormatInt(appID, 10), name, "No genres assigned to this game",
			})
		}
		if !validDate {
			qualityFlags = append(qualityFlags, "invalid_release_date")
		}
		if !ownership.IsValid {
			qualityFlags = append(qualityFlags, "invalid_ownership")
		}

		// Check audio languages without text listing
		textLangMap := make(map[string]bool)
		for _, tl := range textLangs {
			textLangMap[tl.ID] = true
		}
		for _, al := range fullAudioLangs {
			if !textLangMap[al.ID] {
				qualityFlags = append(qualityFlags, "audio_not_in_text")
				issuesCount++
				_ = issuesWriter.Write([]string{
					uuid.New().String(), "audio_not_in_text", "warning",
					strconv.FormatInt(appID, 10), name,
					fmt.Sprintf("Audio language %s listed but not found in supported_languages", al.DisplayName),
				})
			}
		}

		flagsStr := strings.Join(qualityFlags, ",")
		isFree := price == 0.0

		// Write to games.csv
		_ = gamesWriter.Write([]string{
			strconv.FormatInt(appID, 10),
			name,
			cleanRelDate,
			strconv.Itoa(relYear),
			ownership.NormalizedRange,
			strconv.FormatInt(ownership.MinOwners, 10),
			strconv.FormatInt(ownership.MaxOwners, 10),
			strconv.FormatInt(ownership.MidOwners, 10),
			fmt.Sprintf("%.2f", price),
			strconv.FormatBool(isFree),
			strconv.FormatBool(windows),
			strconv.FormatBool(mac),
			strconv.FormatBool(linux),
			strconv.FormatInt(metacritic, 10),
			strconv.FormatInt(userScore, 10),
			strconv.FormatInt(pos, 10),
			strconv.FormatInt(neg, 10),
			headerImg,
			dev,
			pub,
			supLangs,
			audioLangs,
			rawGenres,
			strconv.FormatBool(isPlaytest),
			strconv.FormatBool(hasUsableLangs),
			strconv.FormatBool(hasUsableGenres),
			strconv.Itoa(len(textLangs)),
			strconv.Itoa(len(fullAudioLangs)),
			flagsStr,
		})

		// Write to game_languages.csv
		langSupportMap := make(map[string]*struct {
			hasText  bool
			hasAudio bool
			rawLabel string
		})
		for _, tl := range textLangs {
			langSupportMap[tl.ID] = &struct {
				hasText  bool
				hasAudio bool
				rawLabel string
			}{hasText: true, hasAudio: false, rawLabel: tl.DisplayName}
		}
		for _, al := range fullAudioLangs {
			if entry, ok := langSupportMap[al.ID]; ok {
				entry.hasAudio = true
			} else {
				langSupportMap[al.ID] = &struct {
					hasText  bool
					hasAudio bool
					rawLabel string
				}{hasText: false, hasAudio: true, rawLabel: al.DisplayName}
			}
		}

		for langID, supp := range langSupportMap {
			_ = gameLangsWriter.Write([]string{
				strconv.FormatInt(appID, 10),
				langID,
				strconv.FormatBool(supp.hasText),
				strconv.FormatBool(supp.hasAudio),
				supp.rawLabel,
			})
		}

		// Write to game_genres.csv
		for _, g := range genres {
			_ = gameGenresWriter.Write([]string{
				strconv.FormatInt(appID, 10),
				g.ID,
				g.DisplayName,
			})
		}

		processedCount++
		if processedCount%10000 == 0 && progress != nil {
			progress("Normalizing records", processedCount, totalRows)
		}
	}

	gamesWriter.Flush()
	gameLangsWriter.Flush()
	gameGenresWriter.Flush()
	issuesWriter.Flush()

	gamesFile.Close()
	gameLangsFile.Close()
	gameGenresFile.Close()
	issuesFile.Close()

	// 6. Bulk COPY into staging tables
	if progress != nil {
		progress("Loading staging data into DuckDB", processedCount, totalRows)
	}

	copyCommands := []struct {
		table string
		file  string
	}{
		{"games_staging", filepath.Join(tempDir, "games.csv")},
		{"game_languages_staging", filepath.Join(tempDir, "game_languages.csv")},
		{"game_genres_staging", filepath.Join(tempDir, "game_genres.csv")},
		{"data_quality_issues_staging", filepath.Join(tempDir, "issues.csv")},
	}

	for _, cmd := range copyCommands {
		copySQL := fmt.Sprintf("COPY %s FROM '%s' (HEADER false)", cmd.table, cmd.file)
		if _, err := p.db.Conn().Exec(copySQL); err != nil {
			return nil, fmt.Errorf("failed to load %s into staging: %w", cmd.file, err)
		}
	}

	// 7. Populate languages_staging and genres_staging aggregates
	if progress != nil {
		progress("Computing language and genre aggregates", processedCount, totalRows)
	}

	buildLangSQL := `
		INSERT INTO languages_staging (id, display_name, total_text_games, total_audio_games)
		SELECT 
			gl.language_id,
			MAX(gl.raw_label),
			COUNT(DISTINCT CASE WHEN gl.has_text THEN gl.game_id END),
			COUNT(DISTINCT CASE WHEN gl.has_audio THEN gl.game_id END)
		FROM game_languages_staging gl
		GROUP BY gl.language_id;
	`
	if _, err := p.db.Conn().Exec(buildLangSQL); err != nil {
		return nil, fmt.Errorf("failed to aggregate languages: %w", err)
	}

	buildGenreSQL := `
		INSERT INTO genres_staging (id, display_name, total_games, total_eligible_games)
		SELECT 
			gg.genre_id,
			MAX(gg.display_name),
			COUNT(DISTINCT gg.game_id),
			COUNT(DISTINCT CASE WHEN g.has_usable_languages THEN gg.game_id END)
		FROM game_genres_staging gg
		JOIN games_staging g ON gg.game_id = g.app_id
		GROUP BY gg.genre_id;
	`
	if _, err := p.db.Conn().Exec(buildGenreSQL); err != nil {
		return nil, fmt.Errorf("failed to aggregate genres: %w", err)
	}

	// 8. Atomic table swap
	if progress != nil {
		progress("Applying atomic database update", processedCount, totalRows)
	}

	txSQL := `
		DROP TABLE IF EXISTS games;
		ALTER TABLE games_staging RENAME TO games;

		DROP TABLE IF EXISTS game_languages;
		ALTER TABLE game_languages_staging RENAME TO game_languages;

		DROP TABLE IF EXISTS game_genres;
		ALTER TABLE game_genres_staging RENAME TO game_genres;

		DROP TABLE IF EXISTS languages;
		ALTER TABLE languages_staging RENAME TO languages;

		DROP TABLE IF EXISTS genres;
		ALTER TABLE genres_staging RENAME TO genres;

		DROP TABLE IF EXISTS data_quality_issues;
		ALTER TABLE data_quality_issues_staging RENAME TO data_quality_issues;
	`
	if _, err := p.db.Conn().Exec(txSQL); err != nil {
		return nil, fmt.Errorf("atomic table swap failed: %w", err)
	}

	// 9. Count distinct languages and genres
	var distinctLangs, distinctGenres int64
	p.db.Conn().QueryRow("SELECT count(*) FROM languages").Scan(&distinctLangs)
	p.db.Conn().QueryRow("SELECT count(*) FROM genres").Scan(&distinctGenres)

	duration := time.Since(startTime)

	// 10. Record import history
	recordImportSQL := `
		INSERT INTO dataset_imports (
			id, imported_at, source_type, source_file, total_records,
			games_with_languages, games_with_genres, status, error_message
		) VALUES (?, CURRENT_TIMESTAMP, ?, ?, ?, ?, ?, 'success', '');
	`
	_, _ = p.db.Conn().Exec(recordImportSQL, importID, sourceType, filepath.Base(parquetPath),
		processedCount, withLangsCount, withGenresCount)

	logging.Info("Ingestion complete in %v! Processed %d games (%d text, %d audio, %d genres, %d issues)",
		duration, processedCount, withLangsCount, withAudioCount, withGenresCount, issuesCount)

	stats := &ImportStats{
		TotalRecords:       processedCount,
		GamesWithLanguages: withLangsCount,
		GamesWithAudio:     withAudioCount,
		GamesWithGenres:    withGenresCount,
		SuspectedPlaytests: playtestsCount,
		TotalIssuesLogged:  issuesCount,
		DistinctLanguages:  distinctLangs,
		DistinctGenres:     distinctGenres,
		Duration:           duration,
		ImportID:           importID,
	}

	return stats, nil
}

func (p *Pipeline) createStagingTables() error {
	queries := []string{
		`DROP TABLE IF EXISTS games_staging;
		CREATE TABLE games_staging AS SELECT * FROM games WHERE 1=0;`,

		`DROP TABLE IF EXISTS game_languages_staging;
		CREATE TABLE game_languages_staging AS SELECT * FROM game_languages WHERE 1=0;`,

		`DROP TABLE IF EXISTS game_genres_staging;
		CREATE TABLE game_genres_staging AS SELECT * FROM game_genres WHERE 1=0;`,

		`DROP TABLE IF EXISTS languages_staging;
		CREATE TABLE languages_staging AS SELECT * FROM languages WHERE 1=0;`,

		`DROP TABLE IF EXISTS genres_staging;
		CREATE TABLE genres_staging AS SELECT * FROM genres WHERE 1=0;`,

		`DROP TABLE IF EXISTS data_quality_issues_staging;
		CREATE TABLE data_quality_issues_staging AS SELECT * FROM data_quality_issues WHERE 1=0;`,
	}

	for _, q := range queries {
		if _, err := p.db.Conn().Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) dropStagingTables() {
	tables := []string{
		"games_staging", "game_languages_staging", "game_genres_staging",
		"languages_staging", "genres_staging", "data_quality_issues_staging",
	}
	for _, t := range tables {
		_, _ = p.db.Conn().Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", t))
	}
}
