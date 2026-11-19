# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-11-19

### Added

- Automated Hugging Face Parquet dataset discovery, streaming download, and local Parquet import.
- Robust data normalization pipeline for multi-lingual strings, genres, release dates, and ownership brackets.
- Dialect and regional variant preservation for Chinese, Spanish, and Portuguese.
- Embedded DuckDB analytical storage engine with zero-downtime atomic table swap.
- Interactive Localization Coverage Matrix cross-tabulating 119 languages against 33 genres.
- Dynamic metric switcher (`text`, `audio`, `conversion`) with strict genre-specific denominators.
- Overview Dashboard with headline indicators, distribution histograms, and gap rankings.
- Language Explorer with genre breakdowns, ownership comparisons, and release year trends.
- Genre Analysis featuring side-by-side genre comparison tool with percentage-point differences.
- Audio Localization Gap Analysis investigating catalog drop-offs between text and audio support.
- Estimated Ownership Analysis across source-provided brackets with statistical disclosures.
- Filterable Game Catalog Explorer with modal details view and Steam store links.
- Data Quality & Completeness auditing view with issue drill-downs and raw record inspection.
- Streaming UTF-8 CSV exports across all analytical views and filtered game tables.
- Comprehensive CLI (`serve`, `ingest`, `inspect`, `validate`, `export`) and Makefile convenience targets.
- Full automated test suite for language normalization, analytics calculations, ingestion, API handlers, and frontend logic.
