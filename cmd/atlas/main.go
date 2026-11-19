package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/laymanmax/steam-localization-atlas/internal/analytics"
	"github.com/laymanmax/steam-localization-atlas/internal/api"
	"github.com/laymanmax/steam-localization-atlas/internal/config"
	"github.com/laymanmax/steam-localization-atlas/internal/database"
	"github.com/laymanmax/steam-localization-atlas/internal/dataset"
	"github.com/laymanmax/steam-localization-atlas/internal/ingestion"
	"github.com/laymanmax/steam-localization-atlas/internal/logging"
	"github.com/laymanmax/steam-localization-atlas/internal/repository"
)

func main() {
	cfg := config.Load()
	logging.SetDefaultLevel(cfg.LogLevel)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		cmdServe(cfg, os.Args[2:])
	case "ingest":
		cmdIngest(cfg, os.Args[2:])
	case "inspect":
		cmdInspect(cfg, os.Args[2:])
	case "validate":
		cmdValidate(cfg, os.Args[2:])
	case "export":
		cmdExport(cfg, os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Steam Localization Gap Atlas CLI

Usage:
  atlas <command> [arguments]

Available Commands:
  serve       Start the HTTP API backend
  ingest      Download and/or ingest Parquet dataset into DuckDB
  inspect     Display summary statistics and health of the imported dataset
  validate    Run data quality validation checks on the dataset
  export      Export analytical results to CSV

Examples:
  atlas serve --port 8080
  atlas ingest --source huggingface
  atlas ingest --file ./data/source/games.parquet
  atlas inspect
  atlas validate
  atlas export --type matrix --output matrix.csv`)
}

func cmdServe(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	host := fs.String("host", cfg.Host, "Host address to bind")
	port := fs.Int("port", cfg.Port, "Port to listen on")
	_ = fs.Parse(args)

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		logging.Error("Failed to open database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	server := api.NewServer(db, cfg.DataDir, cfg.AllowedOrigins)
	addr := fmt.Sprintf("%s:%d", *host, *port)

	logging.Info("Steam Localization Gap Atlas API listening on http://%s", addr)
	if !db.HasData() {
		logging.Warn("No dataset loaded yet. Run 'atlas ingest' or initiate import via Web UI.")
	}

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logging.Error("HTTP server error: %v", err)
		os.Exit(1)
	}
}

func cmdIngest(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	source := fs.String("source", "huggingface", "Dataset source ('huggingface' or 'local')")
	file := fs.String("file", "", "Path to local Parquet file (if source=local)")
	_ = fs.Parse(args)

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		logging.Error("Failed to open database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	pipeline := ingestion.NewPipeline(db)
	var parquetPath string

	if *file != "" {
		parquetPath = *file
	} else if *source == "local" {
		defaultLocal := filepath.Join(cfg.DataDir, "source", "games.parquet")
		if _, err := os.Stat(defaultLocal); err != nil {
			logging.Error("No file specified and default %s does not exist", defaultLocal)
			os.Exit(1)
		}
		parquetPath = defaultLocal
	} else {
		// Hugging Face discovery & download
		sourceDir := filepath.Join(cfg.DataDir, "source")
		defaultLocal := filepath.Join(sourceDir, "games.parquet")
		if _, err := os.Stat(defaultLocal); err == nil {
			logging.Info("Found existing downloaded dataset at %s, ingesting...", defaultLocal)
			parquetPath = defaultLocal
		} else {
			logging.Info("Discovering Parquet files on Hugging Face (FronkonGames/steam-games-dataset)...")
			ds := dataset.NewDatasetSource("", sourceDir)
			files, err := ds.DiscoverParquetFiles()
			if err != nil {
				logging.Error("Discovery failed: %v", err)
				os.Exit(1)
			}
			if len(files) == 0 {
				logging.Error("No parquet files found")
				os.Exit(1)
			}

			logging.Info("Downloading %s (%d MB)...", files[0].Path, files[0].Size/1048576)
			downloadedPath, _, err := ds.DownloadFile(files[0], func(done, total int64) {
				pct := float64(done) * 100.0 / float64(total)
				fmt.Printf("\rDownloading: %.1f%% (%d / %d MB)", pct, done/1048576, total/1048576)
			})
			fmt.Println()
			if err != nil {
				logging.Error("Download failed: %v", err)
				os.Exit(1)
			}
			parquetPath = downloadedPath
		}
	}

	start := time.Now()
	stats, err := pipeline.IngestFile(parquetPath, *source, func(stage string, proc, total int64) {
		if proc%25000 == 0 || proc == total {
			fmt.Printf("\r[%s] %d / %d records (%.1f%%)", stage, proc, total, float64(proc)*100.0/float64(total))
		}
	})
	fmt.Println()
	if err != nil {
		logging.Error("Ingestion failed: %v", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== Ingestion Completed Successfully in %v ===\n", time.Since(start))
	fmt.Printf("Total Records:             %d\n", stats.TotalRecords)
	fmt.Printf("Games with Text Support:   %d\n", stats.GamesWithLanguages)
	fmt.Printf("Games with Full Audio:     %d\n", stats.GamesWithAudio)
	fmt.Printf("Games with Genres:         %d\n", stats.GamesWithGenres)
	fmt.Printf("Suspected Playtests:       %d\n", stats.SuspectedPlaytests)
	fmt.Printf("Distinct Languages:        %d\n", stats.DistinctLanguages)
	fmt.Printf("Distinct Genres:           %d\n", stats.DistinctGenres)
	fmt.Printf("Quality Issues Flagged:    %d\n", stats.TotalIssuesLogged)
}

func cmdInspect(cfg *config.Config, args []string) {
	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		logging.Error("Failed to open database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	if !db.HasData() {
		fmt.Println("No dataset currently imported in", cfg.DatabasePath)
		fmt.Println("Run 'atlas ingest' to import the dataset.")
		return
	}

	var totalGames, usableLangs, usableAudio, usableGenres int64
	db.Conn().QueryRow(`
		SELECT 
			COUNT(*),
			COUNT(DISTINCT CASE WHEN has_usable_languages THEN app_id END),
			COUNT(DISTINCT CASE WHEN audio_language_count > 0 THEN app_id END),
			COUNT(DISTINCT CASE WHEN has_usable_genres THEN app_id END)
		FROM games;
	`).Scan(&totalGames, &usableLangs, &usableAudio, &usableGenres)

	var distinctLangs, distinctGenres int64
	db.Conn().QueryRow("SELECT count(*) FROM languages").Scan(&distinctLangs)
	db.Conn().QueryRow("SELECT count(*) FROM genres").Scan(&distinctGenres)

	var lastImport, importSource string
	db.Conn().QueryRow(`
		SELECT imported_at, source_file FROM dataset_imports 
		WHERE status = 'success' ORDER BY imported_at DESC LIMIT 1
	`).Scan(&lastImport, &importSource)

	fmt.Println("=== Steam Localization Gap Atlas: Dataset Inspection ===")
	fmt.Printf("Database:               %s\n", cfg.DatabasePath)
	fmt.Printf("Last Successful Import: %s (source: %s)\n", lastImport, importSource)
	fmt.Printf("Total Games:            %d\n", totalGames)
	fmt.Printf("Eligible Languages:     %d (%.1f%% of catalog)\n", usableLangs, float64(usableLangs)*100.0/float64(totalGames))
	fmt.Printf("Eligible Full Audio:    %d (%.1f%% of catalog)\n", usableAudio, float64(usableAudio)*100.0/float64(totalGames))
	fmt.Printf("Eligible Genres:        %d (%.1f%% of catalog)\n", usableGenres, float64(usableGenres)*100.0/float64(totalGames))
	fmt.Printf("Distinct Languages:     %d\n", distinctLangs)
	fmt.Printf("Distinct Genres:        %d\n", distinctGenres)
}

func cmdValidate(cfg *config.Config, args []string) {
	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		logging.Error("Failed to open database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	if !db.HasData() {
		fmt.Println("No dataset loaded to validate. Run 'atlas ingest' first.")
		return
	}

	var total, missingNames, missingLangs, missingGenres, playtests int64
	db.Conn().QueryRow(`
		SELECT 
			COUNT(*),
			COUNT(DISTINCT CASE WHEN name IS NULL OR name = '' THEN app_id END),
			COUNT(DISTINCT CASE WHEN has_usable_languages = false THEN app_id END),
			COUNT(DISTINCT CASE WHEN has_usable_genres = false THEN app_id END),
			COUNT(DISTINCT CASE WHEN is_playtest = true THEN app_id END)
		FROM games;
	`).Scan(&total, &missingNames, &missingLangs, &missingGenres, &playtests)

	fmt.Println("=== Data Quality Validation Report ===")
	fmt.Printf("Total Records Evaluated: %d\n", total)
	fmt.Printf(" - Missing Names:         %d (%.2f%%)\n", missingNames, float64(missingNames)*100.0/float64(total))
	fmt.Printf(" - Missing Languages:     %d (%.2f%%)\n", missingLangs, float64(missingLangs)*100.0/float64(total))
	fmt.Printf(" - Missing Genres:        %d (%.2f%%)\n", missingGenres, float64(missingGenres)*100.0/float64(total))
	fmt.Printf(" - Suspected Playtests:   %d (%.2f%%)\n", playtests, float64(playtests)*100.0/float64(total))

	// Breakdown of logged issue categories
	rows, err := db.Conn().Query("SELECT issue_type, count(*) FROM data_quality_issues GROUP BY issue_type ORDER BY count(*) DESC")
	if err == nil {
		defer rows.Close()
		fmt.Println("\nLogged Quality Flag Distribution:")
		for rows.Next() {
			var issue string
			var count int64
			rows.Scan(&issue, &count)
			fmt.Printf(" - %-25s: %d\n", issue, count)
		}
	}
}

func cmdExport(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	exportType := fs.String("type", "matrix", "Export type ('matrix', 'gaps', 'games')")
	_ = fs.String("format", "csv", "Export format (default: csv)")
	output := fs.String("output", "", "Output CSV file path")
	_ = fs.Parse(args)

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		logging.Error("Failed to open database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	if !db.HasData() {
		fmt.Println("No dataset loaded. Run 'atlas ingest' first.")
		os.Exit(1)
	}

	if *output == "" {
		*output = fmt.Sprintf("export_%s_%s.csv", *exportType, time.Now().Format("20060102_150405"))
	}

	f, err := os.Create(*output)
	if err != nil {
		logging.Error("Failed to create export file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	engine := analytics.NewEngine(db)

	switch *exportType {
	case "matrix":
		resp, err := engine.GetMatrix("text", &repository.FilterParams{}, "", nil, 30, 50, "overall")
		if err != nil {
			logging.Error("Export failed: %v", err)
			os.Exit(1)
		}
		header := []string{"Language", "Overall Coverage %", "Text Games", "Audio Games"}
		for _, g := range resp.Genres {
			header = append(header, g.Name+" Coverage %", g.Name+" Numerator", g.Name+" Denominator")
		}
		_ = writer.Write(header)
		for _, row := range resp.Languages {
			rec := []string{row.Name, fmt.Sprintf("%.2f%%", row.OverallCoverage), strconv.FormatInt(row.TotalTextGames, 10), strconv.FormatInt(row.TotalAudioGames, 10)}
			for _, g := range resp.Genres {
				cell := row.Cells[g.ID]
				rec = append(rec, fmt.Sprintf("%.2f%%", cell.Percentage), strconv.FormatInt(cell.Numerator, 10), strconv.FormatInt(cell.Denominator, 10))
			}
			_ = writer.Write(rec)
		}
	case "gaps":
		gaps, err := engine.GetAudioGaps(&repository.FilterParams{}, 100)
		if err != nil {
			logging.Error("Export failed: %v", err)
			os.Exit(1)
		}
		_ = writer.Write([]string{"Language", "Text Games", "Audio Games", "Eligible Games", "Text Coverage %", "Audio Coverage %", "Gap (Percentage Points)", "Audio Conversion %"})
		for _, g := range gaps {
			_ = writer.Write([]string{
				g.LanguageName, strconv.FormatInt(g.TextGames, 10), strconv.FormatInt(g.AudioGames, 10),
				strconv.FormatInt(g.EligibleGames, 10), fmt.Sprintf("%.2f%%", g.TextCoverage),
				fmt.Sprintf("%.2f%%", g.AudioCoverage), fmt.Sprintf("%.2f pt", g.PercentagePointGap),
				fmt.Sprintf("%.2f%%", g.AudioConversion),
			})
		}
	case "games":
		games, _, err := engine.GetGames(&repository.FilterParams{}, 5000, 0)
		if err != nil {
			logging.Error("Export failed: %v", err)
			os.Exit(1)
		}
		_ = writer.Write([]string{"App ID", "Name", "Release Date", "Estimated Owners", "Price", "Free", "Playtest"})
		for _, g := range games {
			_ = writer.Write([]string{
				strconv.FormatInt(g.AppID, 10), g.Name, g.ReleaseDate, g.EstimatedOwners,
				fmt.Sprintf("%.2f", g.Price), strconv.FormatBool(g.IsFree), strconv.FormatBool(g.IsPlaytest),
			})
		}
	default:
		logging.Error("Unknown export type: %s", *exportType)
		os.Exit(1)
	}

	fmt.Printf("Successfully exported %s to %s\n", *exportType, *output)
}
