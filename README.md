# Steam Localization Gap Atlas

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.27%2B-00ADD8.svg?logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18-61DAFB.svg?logo=react)](https://react.dev)
[![DuckDB](https://img.shields.io/badge/DuckDB-Embedded%20Analytical%20Engine-FFF000.svg?logo=duckdb)](https://duckdb.org)

**Steam Localization Gap Atlas** is an open-source descriptive research workstation for analyzing the empirical distribution of language support listed across the Steam games catalog.

It systematically investigates two critical fields across **124,146 catalog records**:

1. `supported_languages`: Languages listed for text, interface, and subtitle localization.
2. `full_audio_languages`: Languages for which games list full voice audio support.

---

### Central Research Question

> **How does language support differ across languages, genres, and estimated ownership groups—and where are the biggest differences between listed text/interface support and full audio support?**

---

## Important Descriptive Research Notice

> [!IMPORTANT]
> This is a **descriptive research application**, not a commercial sales forecasting or market opportunity prediction tool.
>
> - **Observed Metadata:** Percentages represent observed publisher listings on Steam store pages, not verified regional player demand.
> - **No Sales Inference:** The dataset does not contain geographic revenue, regional sales, player demographics, or purchasing power by language.
> - **Cost Differences:** A large text-to-audio gap reflects the substantial production cost disparity between written text translation and voice acting/dubbing; it does not indicate negligence or commercial viability.

---

## Key Features

1. **Interactive Localization Coverage Matrix (Flagship Feature)**
   - Granular cross-tabulation of languages (rows) against game genres (columns).
   - Dynamic metric switching between:
     - _Listed text/interface coverage_
     - _Listed full-audio coverage_
     - _Audio conversion conditional on text support_ ($\text{Audio} / \text{Text} \times 100$)
   - Strict genre-specific denominators excluding games lacking language metadata.
   - Interactive hover and focus inspection showing numerators, denominators, and low-sample alerts.
   - Sticky language rows, sticky genre headers, sorting, search, and CSV export.
2. **Overview Dashboard**
   - High-level headline metrics: Total catalog titles (124,146), eligible text titles (115,759), full audio titles (50,518), and central tendency statistics.
   - Most supported languages ranking with text vs audio toggle.
   - Language support breadth histogram (single-language vs multi-language titles).
   - Top genres by language diversity (median languages per game).
   - Largest text-to-audio gap rankings.
3. **Language Explorer**
   - Granular profile of any individual language: overall representation, audio conversion rate, and percentage-point gap.
   - Highest vs lowest representation genres.
   - Coverage across approximate estimated ownership brackets.
   - Release year adoption trends (2005–2025).
   - Searchable game catalog listing titles that support the language.
4. **Genre Analysis & Genre-to-Genre Comparisons**
   - In-depth localization profiles for individual genres (e.g. RPG, Strategy, Indie).
   - Side-by-side comparative tool evaluating two genres with exact percentage-point differences ($\text{pt}$).
5. **Audio Localization Gap Analysis**
   - Dedicated investigation into the structural divide between text translation and full voice dubbing.
   - Identifies that 56.4% of text-localized games list zero full-audio languages.
   - Ranked tables of languages experiencing the steepest drops between text and audio.
6. **Estimated Ownership Analysis**
   - Exploration of localization patterns across source-provided categorical ownership brackets (`0 - 20,000`, `20,000 - 50,000`, etc.).
   - Explicit methodology disclosures explaining why ownership brackets must not be interpreted as causal proof of commercial demand.
7. **Game Catalog Explorer**
   - Filterable catalog with search by title or Steam App ID.
   - Game Detail Drawer displaying Steam CDN capsule image, release date, developers, publishers, price, platforms, and a clear breakdown of text vs audio localization.
8. **Data Quality & Completeness Audit**
   - Transparent reporting on raw dataset irregularities: missing names, missing genres, missing language arrays, and 8,166 suspected playtest titles.
   - Drill-down issue inspection and representative raw records viewer.
9. **Global Filter System**
   - Consistent global filtering across release years, platforms (Windows, macOS, Linux), pricing (Free vs Paid), and genre subsets.
   - Methodology safeguard toggle to include or exclude suspected playtest titles.
10. **Full CSV Export Engine**
    - One-click UTF-8 CSV exports for the visible matrix, language rankings, genre comparisons, gap rankings, and filtered game catalogs.

---

## Architecture & Technology Stack

```
steam-localization-atlas/
├── apps/
│   └── web/                   # React 18 + TypeScript + Vite + Tailwind CSS Frontend
├── cmd/
│   └── atlas/                 # Go CLI & HTTP API Binary Entry Point
├── internal/
│   ├── analytics/             # Authoritative SQL Analytics Engine (DuckDB)
│   ├── api/                   # REST API Handlers, Middleware, CSV Exporters
│   ├── config/                # Environment Variable Configuration
│   ├── database/              # DuckDB Connection Management & Schemas
│   ├── dataset/               # Hugging Face Discovery & Downloader
│   ├── ingestion/             # Parquet Pipeline, Staging, & Atomic Table Swap
│   ├── logging/               # Structured Logger
│   ├── normalization/         # Parsers for Languages, Genres, Ownership, Dates
│   └── repository/            # Parameterized SQL Filter Builders
├── docs/
│   ├── architecture.md        # Technical System Design & Flow
│   ├── data-dictionary.md     # Upstream Fields & Database Schema
│   └── methodology.md         # Mathematical Formulas & Research Ethics
├── Makefile                   # Developer Commands
└── go.mod                     # Go Dependencies
```

- **Backend:** Go 1.27+ with embedded [DuckDB](https://duckdb.org) (`github.com/marcboeker/go-duckdb`). Columnar vectorized queries execute in 20–80ms across 124,146 games.
- **Frontend:** React 18, TypeScript, Tailwind CSS, Lucide icons, Vite 6.
- **Single-Binary Delivery:** The Go server automatically serves the compiled SPA frontend from `apps/web/dist` on `http://localhost:8080`.

---

## Dataset Attribution & License

- **Source Dataset:** [FronkonGames/steam-games-dataset](https://huggingface.co/datasets/FronkonGames/steam-games-dataset) hosted on Hugging Face.
- **Dataset License:** MIT License, according to the upstream dataset repository.
- **Record Count Discrepancy Note:** The upstream README lists 148,909 games, whereas the Hugging Face dataset viewer and `train-00000-of-00001.parquet` shard contain **124,146 records**. This discrepancy originates upstream and is accurately reported by this application.

---

## Getting Started

### Prerequisites

- **Go:** 1.22 or higher
- **Node.js:** 20 or higher
- **pnpm:** 9 or 10

### Quick Start (3 Steps)

```bash
# 1. Clone repository
git clone https://github.com/laymanmax/steam-localization-atlas.git
cd steam-localization-atlas

# 2. Build frontend and backend
make build

# 3. Start the application
./bin/atlas serve --port 8080
```

Open **[http://localhost:8080](http://localhost:8080)** in your browser.

---

## Dataset Ingestion & CLI Commands

The application provides a comprehensive CLI for downloading, inspecting, validating, and serving the dataset.

```bash
# Discover, download, and ingest the official Hugging Face dataset (~183 MB Parquet)
./bin/atlas ingest --source huggingface

# Ingest an existing local Parquet file
./bin/atlas ingest --file ./data/source/games.parquet

# Inspect current database statistics and row counts
./bin/atlas inspect

# Run data quality validation checks
./bin/atlas validate

# Export matrix or gap rankings to CSV
./bin/atlas export --type matrix --output matrix.csv
./bin/atlas export --type gaps --output gaps.csv

# Start HTTP server
./bin/atlas serve --host 0.0.0.0 --port 8080
```

### Makefile Convenience Targets

```bash
make setup     # Install dependencies
make ingest    # Download and ingest dataset
make dev       # Start backend server
make build     # Compile Go binary and frontend SPA
make test      # Run all backend and frontend automated tests
make validate  # Run dataset quality checks
make clean     # Clean build artifacts
```

---

## API Overview

The backend exposes a structured REST API under `/api/v1`:

| Endpoint                      | Method | Description                                            |
| :---------------------------- | :----- | :----------------------------------------------------- |
| `/api/v1/health`              | GET    | Healthcheck and database availability status           |
| `/api/v1/dataset/status`      | GET    | Record counts, file sizes, and import state            |
| `/api/v1/dataset/metadata`    | GET    | Upstream source details and metadata discrepancy notes |
| `/api/v1/dataset/import`      | POST   | Triggers background dataset ingestion                  |
| `/api/v1/overview`            | GET    | Headline statistics, language distributions, rankings  |
| `/api/v1/filters`             | GET    | Available filter options (genres, ownership, years)    |
| `/api/v1/languages`           | GET    | Paginated list of languages with coverage metrics      |
| `/api/v1/languages/:id`       | GET    | Granular profile, genre breakdown, release trends      |
| `/api/v1/languages/:id/games` | GET    | Catalog games supporting the selected language         |
| `/api/v1/genres`              | GET    | Catalog genres with eligible game counts               |
| `/api/v1/genres/:id`          | GET    | Genre profile and language rankings                    |
| `/api/v1/genres/:id/compare`  | GET    | Side-by-side comparison between Genre A and Genre B    |
| `/api/v1/localization/matrix` | GET    | Full matrix cross-tabulation (text, audio, conversion) |
| `/api/v1/localization/gaps`   | GET    | Languages ranked by absolute text-to-audio gap         |
| `/api/v1/ownership/coverage`  | GET    | Coverage across approximate ownership brackets         |
| `/api/v1/games`               | GET    | Paginated game catalog with search and filters         |
| `/api/v1/games/:appId`        | GET    | Single game details, languages breakdown, store link   |
| `/api/v1/quality/summary`     | GET    | Data quality diagnostics and missingness metrics       |
| `/api/v1/quality/issues`      | GET    | Paginated logged data quality issues                   |
| `/api/v1/exports/matrix`      | GET    | Streams CSV export of the active matrix                |
| `/api/v1/exports/games`       | GET    | Streams CSV export of filtered games                   |

---

## Testing

Run the automated test suite covering language normalization, genre deduplication, ownership bounds parsing, analytics denominators, API handlers, and frontend logic:

```bash
make test
```

- **Language Normalization Tests:** Tests HTML entities, asterisks, brackets, delimiters, Simplified vs Traditional Chinese distinction, and regional Spanish/Portuguese variants.
- **Analytics Tests:** Verifies exact matrix numerators and denominators, multi-genre isolation, audio conversion conditional on text, and percentage-point deltas.
- **Ingestion Tests:** Verifies schema inspection, duplicate AppID handling, and atomic database replacement where failed imports leave existing data intact.
- **API Tests:** Validates HTTP responses, error structures, and CSV streaming.
- **Frontend Tests:** Validates filter serialization, matrix calculations, and math helpers.

---

## License

This project is licensed under the [MIT License](./LICENSE).  
The upstream dataset is published under the [MIT License](https://huggingface.co/datasets/FronkonGames/steam-games-dataset).
