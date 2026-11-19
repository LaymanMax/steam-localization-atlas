package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/marcboeker/go-duckdb"
)

// DB wraps the DuckDB connection and handles schema initialization.
type DB struct {
	mu     sync.RWMutex
	conn   *sql.DB
	dbPath string
}

// Open creates or opens a DuckDB database at the specified path.
func Open(dbPath string) (*DB, error) {
	if dbPath != "" && dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory %s: %w", dir, err)
		}
	}

	conn, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb at %s: %w", dbPath, err)
	}

	d := &DB{
		conn:   conn,
		dbPath: dbPath,
	}

	if err := d.InitSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return d, nil
}

// Close closes the underlying DuckDB connection.
func (d *DB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// Conn returns the active sql.DB instance.
func (d *DB) Conn() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.conn
}

// Path returns the database file path.
func (d *DB) Path() string {
	return d.dbPath
}

// InitSchema creates the tables and views if they do not exist.
func (d *DB) InitSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS dataset_imports (
			id VARCHAR PRIMARY KEY,
			imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			source_type VARCHAR,
			source_file VARCHAR,
			total_records BIGINT,
			games_with_languages BIGINT,
			games_with_genres BIGINT,
			status VARCHAR,
			error_message VARCHAR
		);`,

		`CREATE TABLE IF NOT EXISTS games (
			app_id BIGINT PRIMARY KEY,
			name VARCHAR,
			release_date VARCHAR,
			release_year INT,
			estimated_owners VARCHAR,
			owners_min BIGINT,
			owners_max BIGINT,
			owners_mid BIGINT,
			price DOUBLE,
			is_free BOOLEAN,
			windows BOOLEAN,
			mac BOOLEAN,
			linux BOOLEAN,
			metacritic_score INT,
			user_score INT,
			positive INT,
			negative INT,
			header_image VARCHAR,
			developers VARCHAR,
			publishers VARCHAR,
			raw_supported_languages VARCHAR,
			raw_full_audio_languages VARCHAR,
			raw_genres VARCHAR,
			is_playtest BOOLEAN,
			has_usable_languages BOOLEAN,
			has_usable_genres BOOLEAN,
			text_language_count INT,
			audio_language_count INT,
			data_quality_flags VARCHAR
		);`,

		`CREATE TABLE IF NOT EXISTS languages (
			id VARCHAR PRIMARY KEY,
			display_name VARCHAR,
			total_text_games BIGINT DEFAULT 0,
			total_audio_games BIGINT DEFAULT 0
		);`,

		`CREATE TABLE IF NOT EXISTS genres (
			id VARCHAR PRIMARY KEY,
			display_name VARCHAR,
			total_games BIGINT DEFAULT 0,
			total_eligible_games BIGINT DEFAULT 0
		);`,

		`CREATE TABLE IF NOT EXISTS game_languages (
			game_id BIGINT,
			language_id VARCHAR,
			has_text BOOLEAN,
			has_audio BOOLEAN,
			raw_label VARCHAR,
			PRIMARY KEY (game_id, language_id)
		);`,

		`CREATE TABLE IF NOT EXISTS game_genres (
			game_id BIGINT,
			genre_id VARCHAR,
			display_name VARCHAR,
			PRIMARY KEY (game_id, genre_id)
		);`,

		`CREATE TABLE IF NOT EXISTS data_quality_issues (
			id VARCHAR PRIMARY KEY,
			issue_type VARCHAR,
			severity VARCHAR,
			app_id BIGINT,
			game_name VARCHAR,
			description VARCHAR
		);`,
	}

	for _, q := range queries {
		if _, err := d.conn.Exec(q); err != nil {
			return fmt.Errorf("schema error executing %s: %w", q, err)
		}
	}

	return nil
}

// HasData returns true if the games table has records.
func (d *DB) HasData() bool {
	var count int
	err := d.conn.QueryRow("SELECT count(*) FROM games").Scan(&count)
	return err == nil && count > 0
}
