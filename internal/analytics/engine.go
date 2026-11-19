package analytics

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/laymanmax/steam-localization-atlas/internal/database"
	"github.com/laymanmax/steam-localization-atlas/internal/repository"
)

// Engine performs authoritative analytical queries on the DuckDB database.
type Engine struct {
	db *database.DB
}

// NewEngine creates an analytical engine instance.
func NewEngine(db *database.DB) *Engine {
	return &Engine{db: db}
}

// GetOverview computes headline stats and distribution charts for the dashboard.
func (e *Engine) GetOverview(f *repository.FilterParams) (*OverviewSummary, error) {
	whereClause, args := f.BuildWhereClause("g")

	summary := &OverviewSummary{}

	// Basic counts
	statsQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*),
			COUNT(DISTINCT CASE WHEN g.has_usable_languages THEN g.app_id END),
			COUNT(DISTINCT CASE WHEN g.audio_language_count > 0 THEN g.app_id END),
			COUNT(DISTINCT CASE WHEN g.has_usable_genres THEN g.app_id END),
			COALESCE(MEDIAN(CASE WHEN g.has_usable_languages THEN g.text_language_count END), 0),
			COALESCE(AVG(CASE WHEN g.has_usable_languages THEN g.text_language_count END), 0),
			COALESCE(MEDIAN(CASE WHEN g.has_usable_languages THEN g.audio_language_count END), 0),
			COALESCE(AVG(CASE WHEN g.has_usable_languages THEN g.audio_language_count END), 0)
		FROM games g
		WHERE %s;
	`, whereClause)

	err := e.db.Conn().QueryRow(statsQuery, args...).Scan(
		&summary.TotalGames,
		&summary.GamesWithLanguages,
		&summary.GamesWithAudio,
		&summary.GamesWithGenres,
		&summary.MedianTextLanguages,
		&summary.AverageTextLanguages,
		&summary.MedianAudioLanguages,
		&summary.AverageAudioLanguages,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query overview stats: %w", err)
	}

	summary.AverageTextLanguages = math.Round(summary.AverageTextLanguages*100) / 100
	summary.AverageAudioLanguages = math.Round(summary.AverageAudioLanguages*100) / 100

	// Count unique languages & genres in filtered scope
	countQuery := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT gl.language_id),
			COUNT(DISTINCT gg.genre_id),
			COUNT(*)
		FROM games g
		LEFT JOIN game_languages gl ON g.app_id = gl.game_id
		LEFT JOIN game_genres gg ON g.app_id = gg.game_id
		WHERE %s;
	`, whereClause)
	_ = e.db.Conn().QueryRow(countQuery, args...).Scan(
		&summary.UniqueLanguages,
		&summary.UniqueGenres,
		&summary.TotalRelationships,
	)

	// Last import date
	_ = e.db.Conn().QueryRow(`
		SELECT imported_at FROM dataset_imports 
		WHERE status = 'success' 
		ORDER BY imported_at DESC LIMIT 1
	`).Scan(&summary.LastImportDate)

	// Top supported languages (Rankings)
	topLangsQuery := fmt.Sprintf(`
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
		LIMIT 15;
	`, whereClause)

	rows, err := e.db.Conn().Query(topLangsQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item LanguageRankItem
			if err := rows.Scan(&item.ID, &item.Name, &item.TextGames, &item.AudioGames); err == nil {
				if summary.GamesWithLanguages > 0 {
					item.TextCoverage = math.Round(float64(item.TextGames)*10000.0/float64(summary.GamesWithLanguages)) / 100.0
					item.AudioCoverage = math.Round(float64(item.AudioGames)*10000.0/float64(summary.GamesWithLanguages)) / 100.0
				}
				if item.TextGames > 0 {
					item.AudioConversion = math.Round(float64(item.AudioGames)*10000.0/float64(item.TextGames)) / 100.0
				}
				item.PercentagePointGap = math.Round((item.TextCoverage-item.AudioCoverage)*100) / 100
				summary.TopLanguages = append(summary.TopLanguages, item)
			}
		}
	}

	// Text Language distribution histogram bins
	textDistQuery := fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN text_language_count = 1 THEN '1 Language'
				WHEN text_language_count BETWEEN 2 AND 5 THEN '2–5 Languages'
				WHEN text_language_count BETWEEN 6 AND 10 THEN '6–10 Languages'
				WHEN text_language_count BETWEEN 11 AND 20 THEN '11–20 Languages'
				ELSE '21+ Languages'
			END AS bin_label,
			COUNT(*) AS cnt
		FROM games g
		WHERE %s AND g.has_usable_languages = true
		GROUP BY bin_label
		ORDER BY MIN(text_language_count);
	`, whereClause)

	distRows, err := e.db.Conn().Query(textDistQuery, args...)
	if err == nil {
		defer distRows.Close()
		for distRows.Next() {
			var bin DistributionBin
			if err := distRows.Scan(&bin.BinLabel, &bin.Count); err == nil {
				if summary.GamesWithLanguages > 0 {
					bin.Percent = math.Round(float64(bin.Count)*10000.0/float64(summary.GamesWithLanguages)) / 100.0
				}
				summary.LanguageDistribution = append(summary.LanguageDistribution, bin)
			}
		}
	}

	// Audio language distribution histogram bins
	audioDistQuery := fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN audio_language_count = 0 THEN '0 (No Audio)'
				WHEN audio_language_count = 1 THEN '1 Language'
				WHEN audio_language_count BETWEEN 2 AND 5 THEN '2–5 Languages'
				WHEN audio_language_count BETWEEN 6 AND 10 THEN '6–10 Languages'
				ELSE '11+ Languages'
			END AS bin_label,
			COUNT(*) AS cnt
		FROM games g
		WHERE %s AND g.has_usable_languages = true
		GROUP BY bin_label
		ORDER BY MIN(audio_language_count);
	`, whereClause)

	aDistRows, err := e.db.Conn().Query(audioDistQuery, args...)
	if err == nil {
		defer aDistRows.Close()
		for aDistRows.Next() {
			var bin DistributionBin
			if err := aDistRows.Scan(&bin.BinLabel, &bin.Count); err == nil {
				if summary.GamesWithLanguages > 0 {
					bin.Percent = math.Round(float64(bin.Count)*10000.0/float64(summary.GamesWithLanguages)) / 100.0
				}
				summary.AudioDistribution = append(summary.AudioDistribution, bin)
			}
		}
	}

	// Top genres ranked by broad localization coverage
	genreRankQuery := fmt.Sprintf(`
		SELECT 
			gg.genre_id,
			MAX(gg.display_name),
			COUNT(DISTINCT g.app_id) as total_games,
			COUNT(DISTINCT CASE WHEN g.has_usable_languages THEN g.app_id END) as eligible_games,
			COALESCE(MEDIAN(CASE WHEN g.has_usable_languages THEN g.text_language_count END), 0) as med_langs,
			COALESCE(AVG(CASE WHEN g.has_usable_languages THEN g.text_language_count END), 0) as avg_langs,
			COALESCE(AVG(CASE WHEN g.has_usable_languages THEN g.audio_language_count END), 0) as avg_audio
		FROM games g
		JOIN game_genres gg ON g.app_id = gg.game_id
		WHERE %s
		GROUP BY gg.genre_id
		HAVING eligible_games >= 100
		ORDER BY med_langs DESC, avg_langs DESC
		LIMIT 10;
	`, whereClause)

	genreRows, err := e.db.Conn().Query(genreRankQuery, args...)
	if err == nil {
		defer genreRows.Close()
		for genreRows.Next() {
			var item GenreRankItem
			if err := genreRows.Scan(
				&item.ID, &item.Name, &item.TotalGames, &item.EligibleGames,
				&item.MedianTextLanguages, &item.AverageTextLanguages, &item.AverageAudioLanguages,
			); err == nil {
				item.AverageTextLanguages = math.Round(item.AverageTextLanguages*100) / 100
				item.AverageAudioLanguages = math.Round(item.AverageAudioLanguages*100) / 100
				summary.TopGenresCoverage = append(summary.TopGenresCoverage, item)
			}
		}
	}

	// Localization gaps (top languages with largest text - audio coverage gap)
	gaps, err := e.GetAudioGaps(f, 10)
	if err == nil {
		summary.LocalizationGaps = gaps
	}

	return summary, nil
}

// GetMatrix computes the full coverage matrix across languages and genres.
func (e *Engine) GetMatrix(
	metric string,
	f *repository.FilterParams,
	langSearch string,
	selectedGenres []string,
	minSample int,
	limit int,
	sortBy string,
) (*MatrixResponse, error) {
	if minSample <= 0 {
		minSample = 30
	}
	if limit <= 0 || limit > 100 {
		limit = 35
	}
	if metric == "" {
		metric = "text"
	}

	whereClause, args := f.BuildWhereClause("g")

	// 1. Fetch eligible genres
	var genreFilterSQL string
	var genreArgs []any
	if len(selectedGenres) > 0 {
		placeholders := make([]string, len(selectedGenres))
		for i, sg := range selectedGenres {
			placeholders[i] = "?"
			genreArgs = append(genreArgs, sg)
		}
		genreFilterSQL = fmt.Sprintf("AND gg.genre_id IN (%s)", strings.Join(placeholders, ","))
	}

	genresQuery := fmt.Sprintf(`
		SELECT 
			gg.genre_id,
			MAX(gg.display_name) as display_name,
			COUNT(DISTINCT g.app_id) as total_games,
			COUNT(DISTINCT CASE WHEN g.has_usable_languages THEN g.app_id END) as eligible_games
		FROM games g
		JOIN game_genres gg ON g.app_id = gg.game_id
		WHERE %s %s
		GROUP BY gg.genre_id
		HAVING eligible_games >= ?
		ORDER BY eligible_games DESC
		LIMIT 20;
	`, whereClause, genreFilterSQL)

	allGenreArgs := append(append(args, genreArgs...), minSample)
	gRows, err := e.db.Conn().Query(genresQuery, allGenreArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query matrix genres: %w", err)
	}
	defer gRows.Close()

	var genres []MatrixGenre
	genreMap := make(map[string]MatrixGenre)
	for gRows.Next() {
		var g MatrixGenre
		if err := gRows.Scan(&g.ID, &g.Name, &g.TotalGames, &g.EligibleGames); err == nil {
			genres = append(genres, g)
			genreMap[g.ID] = g
		}
	}

	if len(genres) == 0 {
		return &MatrixResponse{
			Metric:            metric,
			Genres:            nil,
			Languages:         nil,
			MinimumSampleSize: minSample,
		}, nil
	}

	// 2. Fetch top languages matching search
	var langSearchSQL string
	var langArgs []any
	if langSearch != "" {
		langSearchSQL = "AND (LOWER(l.display_name) LIKE ? OR LOWER(l.id) LIKE ?)"
		langArgs = append(langArgs, "%"+strings.ToLower(langSearch)+"%", "%"+strings.ToLower(langSearch)+"%")
	}

	langsQuery := fmt.Sprintf(`
		SELECT 
			l.id,
			l.display_name,
			COUNT(DISTINCT CASE WHEN gl.has_text THEN g.app_id END) as total_text_games,
			COUNT(DISTINCT CASE WHEN gl.has_audio THEN g.app_id END) as total_audio_games
		FROM games g
		JOIN game_languages gl ON g.app_id = gl.game_id
		JOIN languages l ON gl.language_id = l.id
		WHERE %s AND g.has_usable_languages = true %s
		GROUP BY l.id, l.display_name
		ORDER BY total_text_games DESC
		LIMIT ?;
	`, whereClause, langSearchSQL)

	allLangArgs := append(append(args, langArgs...), limit)
	lRows, err := e.db.Conn().Query(langsQuery, allLangArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query matrix languages: %w", err)
	}
	defer lRows.Close()

	var langIDs []string
	rowsMap := make(map[string]*MatrixLanguageRow)
	var totalEligibleGames int64
	e.db.Conn().QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM games g WHERE %s AND g.has_usable_languages = true", whereClause), args...).Scan(&totalEligibleGames)

	for lRows.Next() {
		var row MatrixLanguageRow
		if err := lRows.Scan(&row.ID, &row.Name, &row.TotalTextGames, &row.TotalAudioGames); err == nil {
			if totalEligibleGames > 0 {
				row.OverallCoverage = math.Round(float64(row.TotalTextGames)*10000.0/float64(totalEligibleGames)) / 100.0
			}
			row.Cells = make(map[string]MatrixCell)
			rowsMap[row.ID] = &row
			langIDs = append(langIDs, row.ID)
		}
	}

	if len(langIDs) == 0 {
		return &MatrixResponse{
			Metric:            metric,
			Genres:            genres,
			Languages:         nil,
			MinimumSampleSize: minSample,
			TotalEligible:     totalEligibleGames,
		}, nil
	}

	// 3. Compute cell counts
	genrePlaceholders := make([]string, len(genres))
	for i, g := range genres {
		genrePlaceholders[i] = "'" + g.ID + "'"
	}
	langPlaceholders := make([]string, len(langIDs))
	for i, l := range langIDs {
		langPlaceholders[i] = "'" + l + "'"
	}

	cellsQuery := fmt.Sprintf(`
		SELECT 
			gl.language_id,
			gg.genre_id,
			COUNT(DISTINCT CASE WHEN gl.has_text THEN gl.game_id END) as text_count,
			COUNT(DISTINCT CASE WHEN gl.has_audio THEN gl.game_id END) as audio_count,
			COUNT(DISTINCT CASE WHEN gl.has_text AND gl.has_audio THEN gl.game_id END) as conv_count
		FROM games g
		JOIN game_languages gl ON g.app_id = gl.game_id
		JOIN game_genres gg ON g.app_id = gg.game_id
		WHERE %s 
		  AND gl.language_id IN (%s)
		  AND gg.genre_id IN (%s)
		  AND g.has_usable_languages = true
		GROUP BY gl.language_id, gg.genre_id;
	`, whereClause, strings.Join(langPlaceholders, ","), strings.Join(genrePlaceholders, ","))

	cRows, err := e.db.Conn().Query(cellsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query matrix cells: %w", err)
	}
	defer cRows.Close()

	for cRows.Next() {
		var langID, genreID string
		var textCount, audioCount, convCount int64
		if err := cRows.Scan(&langID, &genreID, &textCount, &audioCount, &convCount); err == nil {
			if row, ok := rowsMap[langID]; ok {
				genreMeta := genreMap[genreID]

				var num, denom int64
				switch metric {
				case "audio":
					num = audioCount
					denom = genreMeta.EligibleGames
				case "conversion":
					num = convCount
					denom = textCount // conditional on text support
				default: // "text"
					num = textCount
					denom = genreMeta.EligibleGames
				}

				pct := 0.0
				if denom > 0 {
					pct = math.Round(float64(num)*10000.0/float64(denom)) / 100.0
				}

				row.Cells[genreID] = MatrixCell{
					Numerator:   num,
					Denominator: denom,
					Percentage:  pct,
					IsLowSample: denom < int64(minSample),
				}
			}
		}
	}

	// Ensure all cells exist (0% for missing combinations)
	var languages []MatrixLanguageRow
	for _, lID := range langIDs {
		row := rowsMap[lID]
		for _, g := range genres {
			if _, exists := row.Cells[g.ID]; !exists {
				var denom int64
				if metric == "conversion" {
					denom = 0
				} else {
					denom = g.EligibleGames
				}
				row.Cells[g.ID] = MatrixCell{
					Numerator:   0,
					Denominator: denom,
					Percentage:  0.0,
					IsLowSample: denom < int64(minSample),
				}
			}
		}
		languages = append(languages, *row)
	}

	// Sorting
	switch sortBy {
	case "alphabetical":
		sort.Slice(languages, func(i, j int) bool {
			return languages[i].Name < languages[j].Name
		})
	case "gap":
		sort.Slice(languages, func(i, j int) bool {
			gapI := float64(languages[i].TotalTextGames - languages[i].TotalAudioGames)
			gapJ := float64(languages[j].TotalTextGames - languages[j].TotalAudioGames)
			return gapI > gapJ
		})
	default: // overall coverage desc
		sort.Slice(languages, func(i, j int) bool {
			return languages[i].OverallCoverage > languages[j].OverallCoverage
		})
	}

	var totalDatasetGames int64
	e.db.Conn().QueryRow("SELECT count(*) FROM games").Scan(&totalDatasetGames)

	return &MatrixResponse{
		Metric:            metric,
		Genres:            genres,
		Languages:         languages,
		MinimumSampleSize: minSample,
		TotalEligible:     totalEligibleGames,
		TotalDatasetGames: totalDatasetGames,
	}, nil
}

// GetAudioGaps computes the ranking of languages with the largest text-to-audio gap.
func (e *Engine) GetAudioGaps(f *repository.FilterParams, limit int) ([]LocalizationGapItem, error) {
	if limit <= 0 {
		limit = 20
	}
	whereClause, args := f.BuildWhereClause("g")

	var totalEligible int64
	e.db.Conn().QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM games g WHERE %s AND g.has_usable_languages = true", whereClause), args...).Scan(&totalEligible)

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
		HAVING text_count >= 100
		ORDER BY (text_count - audio_count) DESC
		LIMIT ?;
	`, whereClause)

	allArgs := append(args, limit)
	rows, err := e.db.Conn().Query(query, allArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audio gaps: %w", err)
	}
	defer rows.Close()

	var gaps []LocalizationGapItem
	for rows.Next() {
		var item LocalizationGapItem
		if err := rows.Scan(&item.LanguageID, &item.LanguageName, &item.TextGames, &item.AudioGames); err == nil {
			item.EligibleGames = totalEligible
			if totalEligible > 0 {
				item.TextCoverage = math.Round(float64(item.TextGames)*10000.0/float64(totalEligible)) / 100.0
				item.AudioCoverage = math.Round(float64(item.AudioGames)*10000.0/float64(totalEligible)) / 100.0
			}
			item.PercentagePointGap = math.Round((item.TextCoverage-item.AudioCoverage)*100) / 100
			if item.TextGames > 0 {
				item.AudioConversion = math.Round(float64(item.AudioGames)*10000.0/float64(item.TextGames)) / 100.0
			}
			gaps = append(gaps, item)
		}
	}

	return gaps, nil
}

// GetLanguageDetail retrieves detailed breakdowns for a single language.
func (e *Engine) GetLanguageDetail(langID string, f *repository.FilterParams) (*LanguageDetail, error) {
	whereClause, args := f.BuildWhereClause("g")

	var detail LanguageDetail
	detail.ID = langID

	// Display name & basic counts
	err := e.db.Conn().QueryRow(`SELECT display_name FROM languages WHERE id = ?`, langID).Scan(&detail.Name)
	if err != nil {
		return nil, fmt.Errorf("language %s not found: %w", langID, err)
	}

	var totalEligible int64
	e.db.Conn().QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM games g WHERE %s AND g.has_usable_languages = true", whereClause), args...).Scan(&totalEligible)
	detail.EligibleGames = totalEligible

	countQuery := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT CASE WHEN gl.has_text THEN g.app_id END),
			COUNT(DISTINCT CASE WHEN gl.has_audio THEN g.app_id END)
		FROM games g
		JOIN game_languages gl ON g.app_id = gl.game_id
		WHERE %s AND gl.language_id = ? AND g.has_usable_languages = true;
	`, whereClause)

	allArgs := append(args, langID)
	_ = e.db.Conn().QueryRow(countQuery, allArgs...).Scan(&detail.TextGames, &detail.AudioGames)

	if totalEligible > 0 {
		detail.TextCoverage = math.Round(float64(detail.TextGames)*10000.0/float64(totalEligible)) / 100.0
		detail.AudioCoverage = math.Round(float64(detail.AudioGames)*10000.0/float64(totalEligible)) / 100.0
	}
	detail.PercentagePointGap = math.Round((detail.TextCoverage-detail.AudioCoverage)*100) / 100
	if detail.TextGames > 0 {
		detail.AudioConversion = math.Round(float64(detail.AudioGames)*10000.0/float64(detail.TextGames)) / 100.0
	}

	// Genre breakdowns
	genreQuery := fmt.Sprintf(`
		SELECT 
			gg.genre_id,
			MAX(gg.display_name),
			COUNT(DISTINCT g.app_id) as eligible_in_genre,
			COUNT(DISTINCT CASE WHEN gl.has_text THEN g.app_id END) as text_count,
			COUNT(DISTINCT CASE WHEN gl.has_audio THEN g.app_id END) as audio_count
		FROM games g
		JOIN game_genres gg ON g.app_id = gg.game_id
		LEFT JOIN game_languages gl ON g.app_id = gl.game_id AND gl.language_id = ?
		WHERE %s AND g.has_usable_languages = true
		GROUP BY gg.genre_id
		HAVING eligible_in_genre >= 50
		ORDER BY text_count * 1.0 / eligible_in_genre DESC;
	`, whereClause)

	genreArgs := append([]any{langID}, args...)
	rows, err := e.db.Conn().Query(genreQuery, genreArgs...)
	if err == nil {
		defer rows.Close()
		var allGenreItems []GenreCoverageItem
		for rows.Next() {
			var gi GenreCoverageItem
			if err := rows.Scan(&gi.GenreID, &gi.GenreName, &gi.EligibleGames, &gi.TextGames, &gi.AudioGames); err == nil {
				if gi.EligibleGames > 0 {
					gi.TextCoverage = math.Round(float64(gi.TextGames)*10000.0/float64(gi.EligibleGames)) / 100.0
					gi.AudioCoverage = math.Round(float64(gi.AudioGames)*10000.0/float64(gi.EligibleGames)) / 100.0
				}
				gi.PercentagePointGap = math.Round((gi.TextCoverage-gi.AudioCoverage)*100) / 100
				if gi.TextGames > 0 {
					gi.AudioConversion = math.Round(float64(gi.AudioGames)*10000.0/float64(gi.TextGames)) / 100.0
				}
				allGenreItems = append(allGenreItems, gi)
			}
		}

		if len(allGenreItems) > 0 {
			if len(allGenreItems) <= 5 {
				detail.TopGenres = allGenreItems
			} else {
				detail.TopGenres = allGenreItems[:5]
				// Lowest genres
				reverse := make([]GenreCoverageItem, len(allGenreItems))
				copy(reverse, allGenreItems)
				sort.Slice(reverse, func(i, j int) bool {
					return reverse[i].TextCoverage < reverse[j].TextCoverage
				})
				detail.LowestGenres = reverse[:5]
			}
		}
	}

	// Ownership breakdown
	detail.OwnershipBreakdown, _ = e.GetOwnershipAnalytics(langID, f)

	// Yearly trend
	trendQuery := fmt.Sprintf(`
		SELECT 
			g.release_year,
			COUNT(DISTINCT g.app_id) as total_year_games,
			COUNT(DISTINCT CASE WHEN gl.has_text THEN g.app_id END) as text_count,
			COUNT(DISTINCT CASE WHEN gl.has_audio THEN g.app_id END) as audio_count
		FROM games g
		LEFT JOIN game_languages gl ON g.app_id = gl.game_id AND gl.language_id = ?
		WHERE %s AND g.has_usable_languages = true AND g.release_year BETWEEN 2005 AND 2025
		GROUP BY g.release_year
		HAVING total_year_games >= 20
		ORDER BY g.release_year ASC;
	`, whereClause)

	trendArgs := append([]any{langID}, args...)
	tRows, err := e.db.Conn().Query(trendQuery, trendArgs...)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var ti YearlyTrendItem
			if err := tRows.Scan(&ti.Year, &ti.TotalGames, &ti.TextGames, &ti.AudioGames); err == nil {
				if ti.TotalGames > 0 {
					ti.TextCoverage = math.Round(float64(ti.TextGames)*10000.0/float64(ti.TotalGames)) / 100.0
					ti.AudioCoverage = math.Round(float64(ti.AudioGames)*10000.0/float64(ti.TotalGames)) / 100.0
				}
				ti.PercentagePointGap = math.Round((ti.TextCoverage-ti.AudioCoverage)*100) / 100
				detail.YearlyTrends = append(detail.YearlyTrends, ti)
			}
		}
	}

	return &detail, nil
}

// GetOwnershipAnalytics computes coverage across estimated ownership brackets.
func (e *Engine) GetOwnershipAnalytics(langID string, f *repository.FilterParams) ([]OwnershipCoverageItem, error) {
	whereClause, args := f.BuildWhereClause("g")

	var query string
	var qArgs []any

	if langID != "" {
		query = fmt.Sprintf(`
			SELECT 
				g.estimated_owners,
				COUNT(DISTINCT g.app_id) as total_games,
				COUNT(DISTINCT CASE WHEN gl.has_text THEN g.app_id END) as text_count,
				COUNT(DISTINCT CASE WHEN gl.has_audio THEN g.app_id END) as audio_count,
				AVG(g.owners_mid) as sort_mid
			FROM games g
			LEFT JOIN game_languages gl ON g.app_id = gl.game_id AND gl.language_id = ?
			WHERE %s AND g.has_usable_languages = true AND g.estimated_owners != 'Unknown'
			GROUP BY g.estimated_owners
			ORDER BY sort_mid ASC;
		`, whereClause)
		qArgs = append([]any{langID}, args...)
	} else {
		// Overall ownership brackets
		query = fmt.Sprintf(`
			SELECT 
				g.estimated_owners,
				COUNT(DISTINCT g.app_id) as total_games,
				COUNT(DISTINCT CASE WHEN g.text_language_count > 0 THEN g.app_id END) as text_count,
				COUNT(DISTINCT CASE WHEN g.audio_language_count > 0 THEN g.app_id END) as audio_count,
				AVG(g.owners_mid) as sort_mid
			FROM games g
			WHERE %s AND g.has_usable_languages = true AND g.estimated_owners != 'Unknown'
			GROUP BY g.estimated_owners
			ORDER BY sort_mid ASC;
		`, whereClause)
		qArgs = args
	}

	rows, err := e.db.Conn().Query(query, qArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query ownership brackets: %w", err)
	}
	defer rows.Close()

	var items []OwnershipCoverageItem
	for rows.Next() {
		var item OwnershipCoverageItem
		var sortMid float64
		if err := rows.Scan(&item.Bracket, &item.TotalGames, &item.TextGames, &item.AudioGames, &sortMid); err == nil {
			if item.TotalGames > 0 {
				item.TextCoverage = math.Round(float64(item.TextGames)*10000.0/float64(item.TotalGames)) / 100.0
				item.AudioCoverage = math.Round(float64(item.AudioGames)*10000.0/float64(item.TotalGames)) / 100.0
			}
			item.PercentagePointGap = math.Round((item.TextCoverage-item.AudioCoverage)*100) / 100
			if item.TextGames > 0 {
				item.AudioConversion = math.Round(float64(item.AudioGames)*10000.0/float64(item.TextGames)) / 100.0
			}
			items = append(items, item)
		}
	}

	return items, nil
}

// GetGenreDetail returns in-depth analytics for a single genre.
func (e *Engine) GetGenreDetail(genreID string, f *repository.FilterParams) (*GenreDetail, error) {
	whereClause, args := f.BuildWhereClause("g")

	var detail GenreDetail
	detail.ID = genreID

	err := e.db.Conn().QueryRow(`SELECT display_name FROM genres WHERE id = ?`, genreID).Scan(&detail.Name)
	if err != nil {
		detail.Name = strings.ToUpper(genreID[:1]) + genreID[1:]
	}

	statsQuery := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT g.app_id),
			COUNT(DISTINCT CASE WHEN g.has_usable_languages THEN g.app_id END),
			COALESCE(MEDIAN(CASE WHEN g.has_usable_languages THEN g.text_language_count END), 0),
			COALESCE(AVG(CASE WHEN g.has_usable_languages THEN g.text_language_count END), 0)
		FROM games g
		JOIN game_genres gg ON g.app_id = gg.game_id
		WHERE %s AND gg.genre_id = ?;
	`, whereClause)

	allArgs := append(args, genreID)
	_ = e.db.Conn().QueryRow(statsQuery, allArgs...).Scan(
		&detail.TotalGames,
		&detail.EligibleGames,
		&detail.MedianLanguages,
		&detail.AverageLanguages,
	)
	detail.AverageLanguages = math.Round(detail.AverageLanguages*100) / 100

	// Languages in this genre
	langQuery := fmt.Sprintf(`
		SELECT 
			l.id,
			l.display_name,
			COUNT(DISTINCT CASE WHEN gl.has_text THEN g.app_id END) as text_count,
			COUNT(DISTINCT CASE WHEN gl.has_audio THEN g.app_id END) as audio_count
		FROM games g
		JOIN game_genres gg ON g.app_id = gg.game_id
		JOIN game_languages gl ON g.app_id = gl.game_id
		JOIN languages l ON gl.language_id = l.id
		WHERE %s AND gg.genre_id = ? AND g.has_usable_languages = true
		GROUP BY l.id, l.display_name
		ORDER BY text_count DESC;
	`, whereClause)

	rows, err := e.db.Conn().Query(langQuery, allArgs...)
	if err == nil {
		defer rows.Close()
		var allLangs []LanguageRankItem
		for rows.Next() {
			var li LanguageRankItem
			if err := rows.Scan(&li.ID, &li.Name, &li.TextGames, &li.AudioGames); err == nil {
				if detail.EligibleGames > 0 {
					li.TextCoverage = math.Round(float64(li.TextGames)*10000.0/float64(detail.EligibleGames)) / 100.0
					li.AudioCoverage = math.Round(float64(li.AudioGames)*10000.0/float64(detail.EligibleGames)) / 100.0
				}
				li.PercentagePointGap = math.Round((li.TextCoverage-li.AudioCoverage)*100) / 100
				if li.TextGames > 0 {
					li.AudioConversion = math.Round(float64(li.AudioGames)*10000.0/float64(li.TextGames)) / 100.0
				}
				allLangs = append(allLangs, li)
			}
		}

		if len(allLangs) > 0 {
			if len(allLangs) <= 10 {
				detail.TopLanguages = allLangs
			} else {
				detail.TopLanguages = allLangs[:10]
				// Least supported meeting sample count
				var candidates []LanguageRankItem
				for _, l := range allLangs {
					if l.TextGames >= 5 {
						candidates = append(candidates, l)
					}
				}
				sort.Slice(candidates, func(i, j int) bool {
					return candidates[i].TextGames < candidates[j].TextGames
				})
				if len(candidates) > 10 {
					detail.LowestLanguages = candidates[:10]
				} else {
					detail.LowestLanguages = candidates
				}
			}
		}
	}

	return &detail, nil
}

// CompareGenres compares language representation between Genre A and Genre B.
func (e *Engine) CompareGenres(genreA, genreB string, f *repository.FilterParams) (*GenreComparison, error) {
	whereClause, args := f.BuildWhereClause("g")

	var metaA, metaB MatrixGenre
	metaA.ID = genreA
	metaB.ID = genreB

	e.db.Conn().QueryRow(`SELECT display_name FROM genres WHERE id = ?`, genreA).Scan(&metaA.Name)
	e.db.Conn().QueryRow(`SELECT display_name FROM genres WHERE id = ?`, genreB).Scan(&metaB.Name)

	e.db.Conn().QueryRow(fmt.Sprintf(`
		SELECT COUNT(DISTINCT g.app_id), COUNT(DISTINCT CASE WHEN g.has_usable_languages THEN g.app_id END)
		FROM games g JOIN game_genres gg ON g.app_id = gg.game_id
		WHERE %s AND gg.genre_id = ?
	`, whereClause), append(args, genreA)...).Scan(&metaA.TotalGames, &metaA.EligibleGames)

	e.db.Conn().QueryRow(fmt.Sprintf(`
		SELECT COUNT(DISTINCT g.app_id), COUNT(DISTINCT CASE WHEN g.has_usable_languages THEN g.app_id END)
		FROM games g JOIN game_genres gg ON g.app_id = gg.game_id
		WHERE %s AND gg.genre_id = ?
	`, whereClause), append(args, genreB)...).Scan(&metaB.TotalGames, &metaB.EligibleGames)

	query := fmt.Sprintf(`
		WITH a_stats AS (
			SELECT gl.language_id, l.display_name, COUNT(DISTINCT g.app_id) as games_a
			FROM games g
			JOIN game_genres gg ON g.app_id = gg.game_id AND gg.genre_id = ?
			JOIN game_languages gl ON g.app_id = gl.game_id AND gl.has_text = true
			JOIN languages l ON gl.language_id = l.id
			WHERE %s AND g.has_usable_languages = true
			GROUP BY gl.language_id, l.display_name
		),
		b_stats AS (
			SELECT gl.language_id, l.display_name, COUNT(DISTINCT g.app_id) as games_b
			FROM games g
			JOIN game_genres gg ON g.app_id = gg.game_id AND gg.genre_id = ?
			JOIN game_languages gl ON g.app_id = gl.game_id AND gl.has_text = true
			JOIN languages l ON gl.language_id = l.id
			WHERE %s AND g.has_usable_languages = true
			GROUP BY gl.language_id, l.display_name
		)
		SELECT 
			COALESCE(a.language_id, b.language_id) as lang_id,
			COALESCE(a.display_name, b.display_name) as lang_name,
			COALESCE(a.games_a, 0) as count_a,
			COALESCE(b.games_b, 0) as count_b
		FROM a_stats a
		FULL OUTER JOIN b_stats b ON a.language_id = b.language_id
		ORDER BY count_a + count_b DESC
		LIMIT 30;
	`, whereClause, whereClause)

	allArgs := append([]any{genreA}, args...)
	allArgs = append(allArgs, genreB)
	allArgs = append(allArgs, args...)

	rows, err := e.db.Conn().Query(query, allArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to compare genres: %w", err)
	}
	defer rows.Close()

	var compItems []GenreCompareItem
	for rows.Next() {
		var item GenreCompareItem
		if err := rows.Scan(&item.LanguageID, &item.LanguageName, &item.GamesA, &item.GamesB); err == nil {
			if metaA.EligibleGames > 0 {
				item.CoverageA = math.Round(float64(item.GamesA)*10000.0/float64(metaA.EligibleGames)) / 100.0
			}
			if metaB.EligibleGames > 0 {
				item.CoverageB = math.Round(float64(item.GamesB)*10000.0/float64(metaB.EligibleGames)) / 100.0
			}
			item.DifferencePt = math.Round((item.CoverageA-item.CoverageB)*100) / 100
			compItems = append(compItems, item)
		}
	}

	return &GenreComparison{
		GenreA:     metaA,
		GenreB:     metaB,
		Comparison: compItems,
	}, nil
}

// GetGames returns a paginated list of games matching search and filters.
func (e *Engine) GetGames(f *repository.FilterParams, limit, offset int) ([]GameSummary, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}

	whereClause, args := f.BuildWhereClause("g")

	var totalCount int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM games g WHERE %s", whereClause)
	_ = e.db.Conn().QueryRow(countQuery, args...).Scan(&totalCount)

	query := fmt.Sprintf(`
		SELECT 
			g.app_id, g.name, g.release_date, g.release_year, g.estimated_owners,
			g.price, g.is_free, g.windows, g.mac, g.linux, g.metacritic_score,
			g.header_image, g.developers, g.publishers, g.is_playtest,
			g.has_usable_languages, g.data_quality_flags
		FROM games g
		WHERE %s
		ORDER BY g.app_id DESC
		LIMIT ? OFFSET ?;
	`, whereClause)

	allArgs := append(args, limit, offset)
	rows, err := e.db.Conn().Query(query, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var games []GameSummary
	for rows.Next() {
		var g GameSummary
		var flagsStr string
		err := rows.Scan(
			&g.AppID, &g.Name, &g.ReleaseDate, &g.ReleaseYear, &g.EstimatedOwners,
			&g.Price, &g.IsFree, &g.Windows, &g.Mac, &g.Linux, &g.MetacriticScore,
			&g.HeaderImage, &g.Developers, &g.Publishers, &g.IsPlaytest,
			&g.HasUsableLanguages, &flagsStr,
		)
		if err == nil {
			if flagsStr != "" {
				g.QualityFlags = strings.Split(flagsStr, ",")
			}
			games = append(games, g)
		}
	}

	// Fetch genres and languages for the page
	for i := range games {
		appID := games[i].AppID

		// Genres
		gRows, err := e.db.Conn().Query("SELECT display_name FROM game_genres WHERE game_id = ?", appID)
		if err == nil {
			for gRows.Next() {
				var gn string
				_ = gRows.Scan(&gn)
				games[i].Genres = append(games[i].Genres, gn)
			}
			gRows.Close()
		}

		// Text and audio languages
		lRows, err := e.db.Conn().Query("SELECT raw_label, has_text, has_audio FROM game_languages WHERE game_id = ?", appID)
		if err == nil {
			for lRows.Next() {
				var lbl string
				var ht, ha bool
				if err := lRows.Scan(&lbl, &ht, &ha); err == nil {
					if ht {
						games[i].TextLanguages = append(games[i].TextLanguages, lbl)
					}
					if ha {
						games[i].AudioLanguages = append(games[i].AudioLanguages, lbl)
					}
				}
			}
			lRows.Close()
		}
	}

	return games, totalCount, nil
}

// GetGame retrieves single game details by appID.
func (e *Engine) GetGame(appID int64) (*GameSummary, error) {
	var g GameSummary
	var flagsStr string

	query := `
		SELECT 
			app_id, name, release_date, release_year, estimated_owners,
			price, is_free, windows, mac, linux, metacritic_score,
			header_image, developers, publishers, is_playtest,
			has_usable_languages, data_quality_flags
		FROM games
		WHERE app_id = ?;
	`
	err := e.db.Conn().QueryRow(query, appID).Scan(
		&g.AppID, &g.Name, &g.ReleaseDate, &g.ReleaseYear, &g.EstimatedOwners,
		&g.Price, &g.IsFree, &g.Windows, &g.Mac, &g.Linux, &g.MetacriticScore,
		&g.HeaderImage, &g.Developers, &g.Publishers, &g.IsPlaytest,
		&g.HasUsableLanguages, &flagsStr,
	)
	if err != nil {
		return nil, fmt.Errorf("game %d not found: %w", appID, err)
	}

	if flagsStr != "" {
		g.QualityFlags = strings.Split(flagsStr, ",")
	}

	// Genres
	gRows, err := e.db.Conn().Query("SELECT display_name FROM game_genres WHERE game_id = ?", appID)
	if err == nil {
		for gRows.Next() {
			var gn string
			_ = gRows.Scan(&gn)
			g.Genres = append(g.Genres, gn)
		}
		gRows.Close()
	}

	// Languages
	lRows, err := e.db.Conn().Query("SELECT raw_label, has_text, has_audio FROM game_languages WHERE game_id = ?", appID)
	if err == nil {
		for lRows.Next() {
			var lbl string
			var ht, ha bool
			if err := lRows.Scan(&lbl, &ht, &ha); err == nil {
				if ht {
					g.TextLanguages = append(g.TextLanguages, lbl)
				}
				if ha {
					g.AudioLanguages = append(g.AudioLanguages, lbl)
				}
			}
		}
		lRows.Close()
	}

	return &g, nil
}

// GetQualitySummary returns dataset health diagnostics.
func (e *Engine) GetQualitySummary() (*QualitySummary, error) {
	var s QualitySummary
	s.IssueCounts = make(map[string]int64)

	e.db.Conn().QueryRow(`
		SELECT 
			COUNT(*),
			COUNT(DISTINCT app_id),
			COUNT(DISTINCT CASE WHEN name IS NULL OR name = '' THEN app_id END),
			COUNT(DISTINCT CASE WHEN has_usable_genres = false THEN app_id END),
			COUNT(DISTINCT CASE WHEN release_year = 0 THEN app_id END),
			COUNT(DISTINCT CASE WHEN has_usable_languages = false THEN app_id END),
			COUNT(DISTINCT CASE WHEN is_playtest = true THEN app_id END)
		FROM games;
	`).Scan(
		&s.TotalRecords,
		&s.UniqueAppIDs,
		&s.MissingNames,
		&s.MissingGenres,
		&s.MissingReleaseDates,
		&s.MissingLanguages,
		&s.SuspectedPlaytests,
	)

	s.DuplicateAppIDs = s.TotalRecords - s.UniqueAppIDs
	if s.TotalRecords > 0 {
		s.UsableLanguagePct = math.Round(float64(s.TotalRecords-s.MissingLanguages)*10000.0/float64(s.TotalRecords)) / 100.0
		s.UsableGenrePct = math.Round(float64(s.TotalRecords-s.MissingGenres)*10000.0/float64(s.TotalRecords)) / 100.0
	}

	// Count issues from data_quality_issues table
	rows, err := e.db.Conn().Query(`
		SELECT issue_type, count(*) 
		FROM data_quality_issues 
		GROUP BY issue_type
		ORDER BY count(*) DESC;
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t string
			var c int64
			if err := rows.Scan(&t, &c); err == nil {
				s.IssueCounts[t] = c
			}
		}
	}

	// Sample issues
	sampleRows, err := e.db.Conn().Query(`
		SELECT id, issue_type, severity, app_id, game_name, description
		FROM data_quality_issues
		ORDER BY id DESC
		LIMIT 20;
	`)
	if err == nil {
		defer sampleRows.Close()
		for sampleRows.Next() {
			var item QualityIssueItem
			if err := sampleRows.Scan(&item.ID, &item.IssueType, &item.Severity, &item.AppID, &item.GameName, &item.Description); err == nil {
				s.SampleIssues = append(s.SampleIssues, item)
			}
		}
	}

	return &s, nil
}

// GetQualityIssues returns paginated issues of a given type.
func (e *Engine) GetQualityIssues(issueType string, limit, offset int) ([]QualityIssueItem, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}

	var whereSQL string
	var args []any
	if issueType != "" && issueType != "all" {
		whereSQL = "WHERE issue_type = ?"
		args = append(args, issueType)
	}

	var total int64
	e.db.Conn().QueryRow(fmt.Sprintf("SELECT count(*) FROM data_quality_issues %s", whereSQL), args...).Scan(&total)

	query := fmt.Sprintf(`
		SELECT id, issue_type, severity, app_id, game_name, description
		FROM data_quality_issues
		%s
		ORDER BY app_id ASC
		LIMIT ? OFFSET ?;
	`, whereSQL)

	allArgs := append(args, limit, offset)
	rows, err := e.db.Conn().Query(query, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []QualityIssueItem
	for rows.Next() {
		var item QualityIssueItem
		if err := rows.Scan(&item.ID, &item.IssueType, &item.Severity, &item.AppID, &item.GameName, &item.Description); err == nil {
			items = append(items, item)
		}
	}

	return items, total, nil
}

// GetFilterOptions returns metadata lists for frontend filter selects.
func (e *Engine) GetFilterOptions() (map[string]any, error) {
	res := make(map[string]any)

	// Available genres
	var genres []MatrixGenre
	gRows, err := e.db.Conn().Query(`
		SELECT id, display_name, total_games, total_eligible_games
		FROM genres
		WHERE total_eligible_games >= 50
		ORDER BY total_eligible_games DESC;
	`)
	if err == nil {
		for gRows.Next() {
			var g MatrixGenre
			if err := gRows.Scan(&g.ID, &g.Name, &g.TotalGames, &g.EligibleGames); err == nil {
				genres = append(genres, g)
			}
		}
		gRows.Close()
	}
	res["genres"] = genres

	// Available ownership brackets
	var brackets []string
	oRows, err := e.db.Conn().Query(`
		SELECT estimated_owners, AVG(owners_mid) as mid
		FROM games
		WHERE estimated_owners != 'Unknown'
		GROUP BY estimated_owners
		ORDER BY mid ASC;
	`)
	if err == nil {
		for oRows.Next() {
			var b string
			var mid float64
			if err := oRows.Scan(&b, &mid); err == nil {
				brackets = append(brackets, b)
			}
		}
		oRows.Close()
	}
	res["ownershipBrackets"] = brackets

	// Release year range
	var minYear, maxYear int
	_ = e.db.Conn().QueryRow(`
		SELECT MIN(release_year), MAX(release_year)
		FROM games
		WHERE release_year >= 1995 AND release_year <= 2030;
	`).Scan(&minYear, &maxYear)
	res["releaseYearMin"] = minYear
	res["releaseYearMax"] = maxYear

	return res, nil
}
