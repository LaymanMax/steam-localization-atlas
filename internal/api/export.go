package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handleExportMatrix exports the visible matrix as a CSV file.
func (s *Server) handleExportMatrix(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	metric := q.Get("metric")
	if metric == "" {
		metric = "text"
	}
	minSample := 30
	if msStr := q.Get("minSample"); msStr != "" {
		if ms, err := strconv.Atoi(msStr); err == nil && ms > 0 {
			minSample = ms
		}
	}
	limit := 50
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 150 {
			limit = l
		}
	}

	filters := parseFilters(r)
	resp, err := s.engine.GetMatrix(metric, filters, q.Get("search"), nil, minSample, limit, q.Get("sortBy"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "EXPORT_ERROR", err.Error())
		return
	}

	filename := fmt.Sprintf("steam_localization_matrix_%s_%s.csv", metric, time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header row: Language, Overall Coverage %, [Genre1 Coverage %, Genre1 Numerator, Genre1 Denominator, ...]
	header := []string{"Language", "Overall Coverage %", "Text Games", "Audio Games"}
	for _, g := range resp.Genres {
		header = append(header, fmt.Sprintf("%s (Coverage %%)", g.Name))
		header = append(header, fmt.Sprintf("%s (Numerator)", g.Name))
		header = append(header, fmt.Sprintf("%s (Denominator)", g.Name))
	}
	_ = writer.Write(header)

	for _, row := range resp.Languages {
		record := []string{
			row.Name,
			fmt.Sprintf("%.2f%%", row.OverallCoverage),
			strconv.FormatInt(row.TotalTextGames, 10),
			strconv.FormatInt(row.TotalAudioGames, 10),
		}
		for _, g := range resp.Genres {
			cell := row.Cells[g.ID]
			record = append(record, fmt.Sprintf("%.2f%%", cell.Percentage))
			record = append(record, strconv.FormatInt(cell.Numerator, 10))
			record = append(record, strconv.FormatInt(cell.Denominator, 10))
		}
		_ = writer.Write(record)
	}
}

// handleExportLanguage exports games supporting a language.
func (s *Server) handleExportLanguage(w http.ResponseWriter, r *http.Request) {
	langID := r.PathValue("id")
	filters := parseFilters(r)
	whereClause, args := filters.BuildWhereClause("g")

	query := fmt.Sprintf(`
		SELECT 
			g.app_id, g.name, g.release_date, g.estimated_owners, g.price,
			gl.has_text, gl.has_audio
		FROM games g
		JOIN game_languages gl ON g.app_id = gl.game_id AND gl.language_id = ?
		WHERE %s
		ORDER BY g.app_id DESC
		LIMIT 10000;
	`, whereClause)

	allArgs := append([]any{langID}, args...)
	rows, err := s.db.Conn().Query(query, allArgs...)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "EXPORT_ERROR", err.Error())
		return
	}
	defer rows.Close()

	filename := fmt.Sprintf("games_language_%s_%s.csv", langID, time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{"App ID", "Game Name", "Release Date", "Estimated Owners", "Price", "Text Support", "Full Audio"})

	for rows.Next() {
		var appID int64
		var name, relDate, owners string
		var price float64
		var ht, ha bool
		if err := rows.Scan(&appID, &name, &relDate, &owners, &price, &ht, &ha); err == nil {
			_ = writer.Write([]string{
				strconv.FormatInt(appID, 10),
				name,
				relDate,
				owners,
				fmt.Sprintf("%.2f", price),
				strconv.FormatBool(ht),
				strconv.FormatBool(ha),
			})
		}
	}
}

// handleExportGenre exports language coverage rankings for a genre.
func (s *Server) handleExportGenre(w http.ResponseWriter, r *http.Request) {
	genreID := r.PathValue("id")
	filters := parseFilters(r)
	detail, err := s.engine.GetGenreDetail(genreID, filters)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "EXPORT_ERROR", err.Error())
		return
	}

	filename := fmt.Sprintf("genre_localization_%s_%s.csv", genreID, time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{"Language", "Text Games", "Audio Games", "Text Coverage %", "Audio Coverage %", "Gap (Percentage Points)", "Audio Conversion %"})

	for _, l := range detail.TopLanguages {
		_ = writer.Write([]string{
			l.Name,
			strconv.FormatInt(l.TextGames, 10),
			strconv.FormatInt(l.AudioGames, 10),
			fmt.Sprintf("%.2f%%", l.TextCoverage),
			fmt.Sprintf("%.2f%%", l.AudioCoverage),
			fmt.Sprintf("%.2f pt", l.PercentagePointGap),
			fmt.Sprintf("%.2f%%", l.AudioConversion),
		})
	}
}

// handleExportGaps exports audio gap rankings as CSV.
func (s *Server) handleExportGaps(w http.ResponseWriter, r *http.Request) {
	filters := parseFilters(r)
	gaps, err := s.engine.GetAudioGaps(filters, 100)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "EXPORT_ERROR", err.Error())
		return
	}

	filename := fmt.Sprintf("localization_gaps_%s.csv", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{"Language", "Text Games", "Audio Games", "Eligible Games", "Text Coverage %", "Audio Coverage %", "Gap (Percentage Points)", "Audio Conversion %"})

	for _, g := range gaps {
		_ = writer.Write([]string{
			g.LanguageName,
			strconv.FormatInt(g.TextGames, 10),
			strconv.FormatInt(g.AudioGames, 10),
			strconv.FormatInt(g.EligibleGames, 10),
			fmt.Sprintf("%.2f%%", g.TextCoverage),
			fmt.Sprintf("%.2f%%", g.AudioCoverage),
			fmt.Sprintf("%.2f pt", g.PercentagePointGap),
			fmt.Sprintf("%.2f%%", g.AudioConversion),
		})
	}
}

// handleExportGames exports filtered games list as CSV.
func (s *Server) handleExportGames(w http.ResponseWriter, r *http.Request) {
	filters := parseFilters(r)
	games, _, err := s.engine.GetGames(filters, 5000, 0)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "EXPORT_ERROR", err.Error())
		return
	}

	filename := fmt.Sprintf("steam_games_%s.csv", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{"App ID", "Name", "Release Date", "Estimated Owners", "Price", "Free", "Genres", "Text Languages", "Audio Languages", "Playtest"})

	for _, g := range games {
		_ = writer.Write([]string{
			strconv.FormatInt(g.AppID, 10),
			g.Name,
			g.ReleaseDate,
			g.EstimatedOwners,
			fmt.Sprintf("%.2f", g.Price),
			strconv.FormatBool(g.IsFree),
			strings.Join(g.Genres, ", "),
			strings.Join(g.TextLanguages, ", "),
			strings.Join(g.AudioLanguages, ", "),
			strconv.FormatBool(g.IsPlaytest),
		})
	}
}
