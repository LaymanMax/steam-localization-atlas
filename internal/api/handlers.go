package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/laymanmax/steam-localization-atlas/internal/analytics"
	"github.com/laymanmax/steam-localization-atlas/internal/database"
	"github.com/laymanmax/steam-localization-atlas/internal/dataset"
	"github.com/laymanmax/steam-localization-atlas/internal/ingestion"
	"github.com/laymanmax/steam-localization-atlas/internal/repository"
)

// Server coordinates HTTP routing and handlers.
type Server struct {
	db       *database.DB
	engine   *analytics.Engine
	pipeline *ingestion.Pipeline
	dataDir  string
	origins  string
	mux      *http.ServeMux

	// Ingestion task tracking
	importMu     sync.RWMutex
	importActive bool
	importStatus string
	importError  string
}

// NewServer initializes the HTTP API server.
func NewServer(db *database.DB, dataDir, allowedOrigins string) *Server {
	s := &Server{
		db:       db,
		engine:   analytics.NewEngine(db),
		pipeline: ingestion.NewPipeline(db),
		dataDir:  dataDir,
		origins:  allowedOrigins,
		mux:      http.NewServeMux(),
	}

	s.registerRoutes()
	return s
}

// Handler returns the HTTP handler with middleware applied.
func (s *Server) Handler() http.Handler {
	return Chain(
		s.mux,
		CORSMiddleware(s.origins),
		LoggingMiddleware,
		RecoveryMiddleware,
	)
}

func (s *Server) registerRoutes() {
	// System & Dataset
	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/v1/dataset/status", s.handleDatasetStatus)
	s.mux.HandleFunc("GET /api/v1/dataset/metadata", s.handleDatasetMetadata)
	s.mux.HandleFunc("POST /api/v1/dataset/import", s.handleDatasetImport)
	s.mux.HandleFunc("GET /api/v1/dataset/imports", s.handleDatasetImports)

	// Analytical Endpoints
	s.mux.HandleFunc("GET /api/v1/overview", s.handleOverview)
	s.mux.HandleFunc("GET /api/v1/filters", s.handleFilters)
	s.mux.HandleFunc("GET /api/v1/languages", s.handleLanguages)
	s.mux.HandleFunc("GET /api/v1/languages/{id}", s.handleLanguageDetail)
	s.mux.HandleFunc("GET /api/v1/languages/{id}/games", s.handleLanguageGames)
	s.mux.HandleFunc("GET /api/v1/genres", s.handleGenres)
	s.mux.HandleFunc("GET /api/v1/genres/{id}", s.handleGenreDetail)
	s.mux.HandleFunc("GET /api/v1/genres/{id}/languages", s.handleGenreLanguages)
	s.mux.HandleFunc("GET /api/v1/genres/{id}/compare", s.handleGenreCompare)
	s.mux.HandleFunc("GET /api/v1/localization/matrix", s.handleLocalizationMatrix)
	s.mux.HandleFunc("GET /api/v1/localization/gaps", s.handleLocalizationGaps)
	s.mux.HandleFunc("GET /api/v1/localization/audio-coverage", s.handleLocalizationAudioCoverage)
	s.mux.HandleFunc("GET /api/v1/ownership/coverage", s.handleOwnershipCoverage)
	s.mux.HandleFunc("GET /api/v1/games", s.handleGames)
	s.mux.HandleFunc("GET /api/v1/games/{appId}", s.handleGameDetail)
	s.mux.HandleFunc("GET /api/v1/quality/summary", s.handleQualitySummary)
	s.mux.HandleFunc("GET /api/v1/quality/issues", s.handleQualityIssues)
	s.mux.HandleFunc("GET /api/v1/quality/issues/{type}", s.handleQualityIssuesByType)

	// CSV Export Endpoints
	s.mux.HandleFunc("GET /api/v1/exports/matrix", s.handleExportMatrix)
	s.mux.HandleFunc("GET /api/v1/exports/language/{id}", s.handleExportLanguage)
	s.mux.HandleFunc("GET /api/v1/exports/genre/{id}", s.handleExportGenre)
	s.mux.HandleFunc("GET /api/v1/exports/gaps", s.handleExportGaps)
	s.mux.HandleFunc("GET /api/v1/exports/games", s.handleExportGames)

	// Serve Frontend SPA if dist directory exists
	distDirs := []string{"./apps/web/dist", "../apps/web/dist", filepath.Join(s.dataDir, "../apps/web/dist")}
	for _, dir := range distDirs {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			fileServer := http.FileServer(http.Dir(dir))
			s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/api/") {
					http.NotFound(w, r)
					return
				}
				targetPath := filepath.Join(dir, filepath.Clean(r.URL.Path))
				if fi, err := os.Stat(targetPath); os.IsNotExist(err) || fi.IsDir() {
					http.ServeFile(w, r, filepath.Join(dir, "index.html"))
					return
				}
				fileServer.ServeHTTP(w, r)
			})
			break
		}
	}
}

// parseFilters extracts common analytical filter parameters from URL query values.
func parseFilters(r *http.Request) *repository.FilterParams {
	q := r.URL.Query()
	f := &repository.FilterParams{
		Platform:         q.Get("platform"),
		Search:           q.Get("search"),
		ExcludePlaytests: q.Get("excludePlaytests") == "true",
		RequireLanguages: q.Get("requireLanguages") == "true",
	}

	if yMin, err := strconv.Atoi(q.Get("releaseYearMin")); err == nil {
		f.ReleaseYearMin = yMin
	}
	if yMax, err := strconv.Atoi(q.Get("releaseYearMax")); err == nil {
		f.ReleaseYearMax = yMax
	}
	if q.Has("isFree") {
		isFree := q.Get("isFree") == "true"
		f.IsFree = &isFree
	}
	if genres := q.Get("genres"); genres != "" {
		for _, g := range strings.Split(genres, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
				f.Genres = append(f.Genres, g)
			}
		}
	}
	if owners := q.Get("ownership"); owners != "" {
		for _, o := range strings.Split(owners, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				f.Ownership = append(f.Ownership, o)
			}
		}
	}

	return f
}

// GET /api/v1/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"hasData":   s.db.HasData(),
	})
}

// GET /api/v1/dataset/status
func (s *Server) handleDatasetStatus(w http.ResponseWriter, r *http.Request) {
	var totalRecords int64
	var lastImport string
	hasData := s.db.HasData()

	if hasData {
		_ = s.db.Conn().QueryRow("SELECT count(*) FROM games").Scan(&totalRecords)
		_ = s.db.Conn().QueryRow(`
			SELECT imported_at FROM dataset_imports 
			WHERE status = 'success' 
			ORDER BY imported_at DESC LIMIT 1
		`).Scan(&lastImport)
	}

	s.importMu.RLock()
	active := s.importActive
	statusMsg := s.importStatus
	errMsg := s.importError
	s.importMu.RUnlock()

	var dbSize int64
	if fi, err := os.Stat(s.db.Path()); err == nil {
		dbSize = fi.Size()
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"hasData":           hasData,
		"totalRecords":      totalRecords,
		"lastImport":        lastImport,
		"databasePath":      s.db.Path(),
		"databaseSizeBytes": dbSize,
		"isImporting":       active,
		"importStatus":      statusMsg,
		"importError":       errMsg,
	})
}

// GET /api/v1/dataset/metadata
func (s *Server) handleDatasetMetadata(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{
		"datasetName":           "FronkonGames/steam-games-dataset",
		"upstreamLicense":       "MIT",
		"upstreamHuggingFace":   "https://huggingface.co/datasets/FronkonGames/steam-games-dataset",
		"metadataDiscrepancy":   "Upstream README reports 148,909 games, while the Hugging Face viewer / train shard contains 124,146 records. This discrepancy originates upstream and is reported accurately.",
		"actualImportedRecords": 124146,
		"keyFields": []string{
			"appID", "name", "release_date", "estimated_owners", "price",
			"supported_languages", "full_audio_languages", "genres",
		},
	})
}

// POST /api/v1/dataset/import
type ImportRequest struct {
	Source string `json:"source"` // "huggingface" or "local"
	File   string `json:"file"`   // path to local parquet if source == "local"
}

func (s *Server) handleDatasetImport(w http.ResponseWriter, r *http.Request) {
	s.importMu.Lock()
	if s.importActive {
		s.importMu.Unlock()
		WriteError(w, http.StatusConflict, "IMPORT_IN_PROGRESS", "A dataset import is currently running")
		return
	}
	s.importActive = true
	s.importStatus = "Starting import..."
	s.importError = ""
	s.importMu.Unlock()

	var req ImportRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Source == "" {
		req.Source = "huggingface"
	}

	go func() {
		defer func() {
			s.importMu.Lock()
			s.importActive = false
			s.importMu.Unlock()
		}()

		var parquetFile string

		if req.Source == "local" && req.File != "" {
			parquetFile = req.File
		} else {
			// Check if local source already exists
			localDefault := fmt.Sprintf("%s/source/games.parquet", s.dataDir)
			if _, err := os.Stat(localDefault); err == nil {
				parquetFile = localDefault
			} else {
				// Download from Hugging Face
				s.importMu.Lock()
				s.importStatus = "Discovering Hugging Face files..."
				s.importMu.Unlock()

				ds := dataset.NewDatasetSource("", fmt.Sprintf("%s/source", s.dataDir))
				files, err := ds.DiscoverParquetFiles()
				if err != nil {
					s.importMu.Lock()
					s.importError = fmt.Sprintf("Discovery failed: %v", err)
					s.importMu.Unlock()
					return
				}

				if len(files) == 0 {
					s.importMu.Lock()
					s.importError = "No parquet files found in repository"
					s.importMu.Unlock()
					return
				}

				s.importMu.Lock()
				s.importStatus = fmt.Sprintf("Downloading %s...", files[0].Path)
				s.importMu.Unlock()

				destPath, _, err := ds.DownloadFile(files[0], func(done, total int64) {
					pct := float64(done) * 100.0 / float64(total)
					s.importMu.Lock()
					s.importStatus = fmt.Sprintf("Downloading: %.1f%% (%d / %d MB)", pct, done/1048576, total/1048576)
					s.importMu.Unlock()
				})
				if err != nil {
					s.importMu.Lock()
					s.importError = fmt.Sprintf("Download failed: %v", err)
					s.importMu.Unlock()
					return
				}
				parquetFile = destPath
			}
		}

		s.importMu.Lock()
		s.importStatus = "Normalizing and rebuilding analytical database..."
		s.importMu.Unlock()

		stats, err := s.pipeline.IngestFile(parquetFile, req.Source, func(stage string, proc, total int64) {
			s.importMu.Lock()
			s.importStatus = fmt.Sprintf("%s: %d / %d", stage, proc, total)
			s.importMu.Unlock()
		})
		if err != nil {
			s.importMu.Lock()
			s.importError = fmt.Sprintf("Ingestion failed: %v", err)
			s.importMu.Unlock()
			return
		}

		s.importMu.Lock()
		s.importStatus = fmt.Sprintf("Completed successfully! Ingested %d records in %v", stats.TotalRecords, stats.Duration)
		s.importMu.Unlock()
	}()

	WriteJSON(w, http.StatusAccepted, map[string]any{
		"message": "Dataset import initiated in background",
		"source":  req.Source,
	})
}

// GET /api/v1/dataset/imports
func (s *Server) handleDatasetImports(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Conn().Query(`
		SELECT id, imported_at, source_type, source_file, total_records,
		       games_with_languages, games_with_genres, status, error_message
		FROM dataset_imports
		ORDER BY imported_at DESC
		LIMIT 20;
	`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var (
			id, importedAt, srcType, srcFile, status, errMsg string
			total, withLangs, withGenres                     int64
		)
		if err := rows.Scan(&id, &importedAt, &srcType, &srcFile, &total, &withLangs, &withGenres, &status, &errMsg); err == nil {
			list = append(list, map[string]any{
				"id":                 id,
				"importedAt":         importedAt,
				"sourceType":         srcType,
				"sourceFile":         srcFile,
				"totalRecords":       total,
				"gamesWithLanguages": withLangs,
				"gamesWithGenres":    withGenres,
				"status":             status,
				"errorMessage":       errMsg,
			})
		}
	}

	WriteJSON(w, http.StatusOK, list)
}

// GET /api/v1/overview
func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	if !s.db.HasData() {
		WriteError(w, http.StatusNotFound, "DATASET_EMPTY", "No dataset imported yet. Please run ingestion.")
		return
	}

	filters := parseFilters(r)
	summary, err := s.engine.GetOverview(filters)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "ANALYTICS_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, summary)
}

// GET /api/v1/filters
func (s *Server) handleFilters(w http.ResponseWriter, r *http.Request) {
	options, err := s.engine.GetFilterOptions()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, options)
}

// GET /api/v1/languages
func (s *Server) handleLanguages(w http.ResponseWriter, r *http.Request) {
	filters := parseFilters(r)
	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 150 {
			limit = l
		}
	}

	whereClause, args := filters.BuildWhereClause("g")
	var totalEligible int64
	_ = s.db.Conn().QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM games g WHERE %s AND g.has_usable_languages = true", whereClause), args...).Scan(&totalEligible)

	query := fmt.Sprintf(`
		SELECT 
			l.id,
			l.display_name,
			COUNT(DISTINCT CASE WHEN gl.has_text THEN g.app_id END) as text_count,
			COUNT(DISTINCT CASE WHEN gl.has_audio THEN g.app_id END) as audio_count
		FROM games g
		JOIN game_languages gl ON g.app_id = gl.game_id
		JOIN languages l ON gl.language_id = l.id
		WHERE %s AND g.has_usable_languages = true
		GROUP BY l.id, l.display_name
		ORDER BY text_count DESC
		LIMIT ?;
	`, whereClause)

	allArgs := append(args, limit)
	rows, err := s.db.Conn().Query(query, allArgs...)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	defer rows.Close()

	var list []analytics.LanguageRankItem
	for rows.Next() {
		var item analytics.LanguageRankItem
		if err := rows.Scan(&item.ID, &item.Name, &item.TextGames, &item.AudioGames); err == nil {
			if totalEligible > 0 {
				item.TextCoverage = round2(float64(item.TextGames) * 100.0 / float64(totalEligible))
				item.AudioCoverage = round2(float64(item.AudioGames) * 100.0 / float64(totalEligible))
			}
			item.PercentagePointGap = round2(item.TextCoverage - item.AudioCoverage)
			if item.TextGames > 0 {
				item.AudioConversion = round2(float64(item.AudioGames) * 100.0 / float64(item.TextGames))
			}
			list = append(list, item)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"languages":     list,
		"totalEligible": totalEligible,
	})
}

// GET /api/v1/languages/{id}
func (s *Server) handleLanguageDetail(w http.ResponseWriter, r *http.Request) {
	langID := r.PathValue("id")
	if langID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_PARAM", "Language ID is required")
		return
	}

	filters := parseFilters(r)
	detail, err := s.engine.GetLanguageDetail(langID, filters)
	if err != nil {
		WriteError(w, http.StatusNotFound, "LANGUAGE_NOT_FOUND", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, detail)
}

// GET /api/v1/languages/{id}/games
func (s *Server) handleLanguageGames(w http.ResponseWriter, r *http.Request) {
	langID := r.PathValue("id")
	limit := 25
	offset := 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	filters := parseFilters(r)
	whereClause, args := filters.BuildWhereClause("g")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT g.app_id)
		FROM games g
		JOIN game_languages gl ON g.app_id = gl.game_id AND gl.language_id = ?
		WHERE %s;
	`, whereClause)

	var total int64
	allArgs := append([]any{langID}, args...)
	_ = s.db.Conn().QueryRow(countQuery, allArgs...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT 
			g.app_id, g.name, g.release_date, g.release_year, g.estimated_owners,
			g.price, g.is_free, g.windows, g.mac, g.linux, g.metacritic_score,
			g.header_image, g.developers, g.publishers, g.is_playtest,
			g.has_usable_languages, g.data_quality_flags,
			gl.has_text, gl.has_audio
		FROM games g
		JOIN game_languages gl ON g.app_id = gl.game_id AND gl.language_id = ?
		WHERE %s
		ORDER BY g.app_id DESC
		LIMIT ? OFFSET ?;
	`, whereClause)

	fetchArgs := append(append([]any{langID}, args...), limit, offset)
	rows, err := s.db.Conn().Query(query, fetchArgs...)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	defer rows.Close()

	var games []map[string]any
	for rows.Next() {
		var g analytics.GameSummary
		var flagsStr string
		var ht, ha bool
		err := rows.Scan(
			&g.AppID, &g.Name, &g.ReleaseDate, &g.ReleaseYear, &g.EstimatedOwners,
			&g.Price, &g.IsFree, &g.Windows, &g.Mac, &g.Linux, &g.MetacriticScore,
			&g.HeaderImage, &g.Developers, &g.Publishers, &g.IsPlaytest,
			&g.HasUsableLanguages, &flagsStr, &ht, &ha,
		)
		if err == nil {
			games = append(games, map[string]any{
				"appId":              g.AppID,
				"name":               g.Name,
				"releaseDate":        g.ReleaseDate,
				"releaseYear":        g.ReleaseYear,
				"estimatedOwners":    g.EstimatedOwners,
				"price":              g.Price,
				"isFree":             g.IsFree,
				"windows":            g.Windows,
				"mac":                g.Mac,
				"linux":              g.Linux,
				"metacriticScore":    g.MetacriticScore,
				"headerImage":        g.HeaderImage,
				"hasTextSupport":     ht,
				"hasFullAudio":       ha,
				"isPlaytest":         g.IsPlaytest,
				"hasUsableLanguages": g.HasUsableLanguages,
			})
		}
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"games":  games,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GET /api/v1/genres
func (s *Server) handleGenres(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Conn().Query(`
		SELECT id, display_name, total_games, total_eligible_games
		FROM genres
		WHERE total_eligible_games >= 10
		ORDER BY total_eligible_games DESC;
	`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	defer rows.Close()

	var genres []analytics.MatrixGenre
	for rows.Next() {
		var g analytics.MatrixGenre
		if err := rows.Scan(&g.ID, &g.Name, &g.TotalGames, &g.EligibleGames); err == nil {
			genres = append(genres, g)
		}
	}

	WriteJSON(w, http.StatusOK, genres)
}

// GET /api/v1/genres/{id}
func (s *Server) handleGenreDetail(w http.ResponseWriter, r *http.Request) {
	genreID := r.PathValue("id")
	filters := parseFilters(r)

	detail, err := s.engine.GetGenreDetail(genreID, filters)
	if err != nil {
		WriteError(w, http.StatusNotFound, "GENRE_NOT_FOUND", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, detail)
}

// GET /api/v1/genres/{id}/languages
func (s *Server) handleGenreLanguages(w http.ResponseWriter, r *http.Request) {
	genreID := r.PathValue("id")
	filters := parseFilters(r)

	detail, err := s.engine.GetGenreDetail(genreID, filters)
	if err != nil {
		WriteError(w, http.StatusNotFound, "GENRE_NOT_FOUND", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, detail.TopLanguages)
}

// GET /api/v1/genres/{id}/compare?against=rpg
func (s *Server) handleGenreCompare(w http.ResponseWriter, r *http.Request) {
	genreA := r.PathValue("id")
	genreB := r.URL.Query().Get("against")
	if genreB == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_PARAM", "Query parameter 'against' required (e.g. against=rpg)")
		return
	}

	filters := parseFilters(r)
	comp, err := s.engine.CompareGenres(genreA, genreB, filters)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "COMPARISON_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, comp)
}

// GET /api/v1/localization/matrix
func (s *Server) handleLocalizationMatrix(w http.ResponseWriter, r *http.Request) {
	if !s.db.HasData() {
		WriteError(w, http.StatusNotFound, "DATASET_EMPTY", "No dataset imported yet")
		return
	}

	q := r.URL.Query()
	metric := q.Get("metric")
	if metric == "" {
		metric = "text"
	}
	langSearch := q.Get("search")
	minSample := 30
	if msStr := q.Get("minSample"); msStr != "" {
		if ms, err := strconv.Atoi(msStr); err == nil && ms > 0 {
			minSample = ms
		}
	}
	limit := 35
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	sortBy := q.Get("sortBy")
	if sortBy == "" {
		sortBy = "overall"
	}

	var selectedGenres []string
	if gStr := q.Get("genres"); gStr != "" {
		for _, part := range strings.Split(gStr, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				selectedGenres = append(selectedGenres, part)
			}
		}
	}

	filters := parseFilters(r)
	resp, err := s.engine.GetMatrix(metric, filters, langSearch, selectedGenres, minSample, limit, sortBy)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MATRIX_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}

// GET /api/v1/localization/gaps
func (s *Server) handleLocalizationGaps(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	filters := parseFilters(r)
	gaps, err := s.engine.GetAudioGaps(filters, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "GAPS_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, gaps)
}

// GET /api/v1/localization/audio-coverage
func (s *Server) handleLocalizationAudioCoverage(w http.ResponseWriter, r *http.Request) {
	filters := parseFilters(r)
	whereClause, args := filters.BuildWhereClause("g")

	var (
		totalWithText, totalWithAudio, textNoAudio int64
	)

	query := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT CASE WHEN g.text_language_count > 0 THEN g.app_id END),
			COUNT(DISTINCT CASE WHEN g.audio_language_count > 0 THEN g.app_id END),
			COUNT(DISTINCT CASE WHEN g.text_language_count > 0 AND g.audio_language_count = 0 THEN g.app_id END)
		FROM games g
		WHERE %s AND g.has_usable_languages = true;
	`, whereClause)

	_ = s.db.Conn().QueryRow(query, args...).Scan(&totalWithText, &totalWithAudio, &textNoAudio)

	gaps, err := s.engine.GetAudioGaps(filters, 15)
	if err != nil {
		gaps = nil
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"gamesWithText":        totalWithText,
		"gamesWithFullAudio":   totalWithAudio,
		"gamesWithTextNoAudio": textNoAudio,
		"largestGaps":          gaps,
	})
}

// GET /api/v1/ownership/coverage
func (s *Server) handleOwnershipCoverage(w http.ResponseWriter, r *http.Request) {
	langID := r.URL.Query().Get("language")
	filters := parseFilters(r)

	items, err := s.engine.GetOwnershipAnalytics(langID, filters)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "OWNERSHIP_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, items)
}

// GET /api/v1/games
func (s *Server) handleGames(w http.ResponseWriter, r *http.Request) {
	limit := 25
	offset := 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	filters := parseFilters(r)
	games, total, err := s.engine.GetGames(filters, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "GAMES_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"games":  games,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GET /api/v1/games/{appId}
func (s *Server) handleGameDetail(w http.ResponseWriter, r *http.Request) {
	appIDStr := r.PathValue("appId")
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_APP_ID", "Invalid Steam App ID")
		return
	}

	game, err := s.engine.GetGame(appID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "GAME_NOT_FOUND", fmt.Sprintf("Game %d not found", appID))
		return
	}

	WriteJSON(w, http.StatusOK, game)
}

// GET /api/v1/quality/summary
func (s *Server) handleQualitySummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.engine.GetQualitySummary()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "QUALITY_ERROR", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

// GET /api/v1/quality/issues
func (s *Server) handleQualityIssues(w http.ResponseWriter, r *http.Request) {
	limit := 25
	offset := 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	issueType := r.URL.Query().Get("type")
	issues, total, err := s.engine.GetQualityIssues(issueType, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "QUALITY_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"issues": issues,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GET /api/v1/quality/issues/{type}
func (s *Server) handleQualityIssuesByType(w http.ResponseWriter, r *http.Request) {
	issueType := r.PathValue("type")
	limit := 25
	offset := 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	issues, total, err := s.engine.GetQualityIssues(issueType, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "QUALITY_ERROR", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"issueType": issueType,
		"issues":    issues,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
