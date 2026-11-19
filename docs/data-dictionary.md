# Data Dictionary & Schema Specification

This document details the upstream source schema, normalized analytical database tables, relational modeling, and data quality flag definitions for the **Steam Localization Gap Atlas**.

---

## 1. Upstream Source Dataset

**Dataset Source:** [FronkonGames/steam-games-dataset](https://huggingface.co/datasets/FronkonGames/steam-games-dataset)  
**Format:** Apache Parquet (`train-00000-of-00001.parquet`)  
**Record Count:** 124,146 actual rows (Note: Upstream README reports 148,909, documented as an upstream discrepancy).

| Field                  | Source Type      | Description                                    | Handling                                                    |
| :--------------------- | :--------------- | :--------------------------------------------- | :---------------------------------------------------------- |
| `appID`                | VARCHAR / BIGINT | Unique Steam application identifier            | Parsed to BIGINT primary key; checked for duplicates        |
| `name`                 | VARCHAR          | Title of the game or software                  | Sanitized; audited for missing values and playtest cues     |
| `release_date`         | VARCHAR          | Date of release on Steam (various formats)     | Extracted to standardized `YYYY-MM-DD` and `release_year`   |
| `estimated_owners`     | VARCHAR          | Estimated owner bracket (e.g. `0 - 20000`)     | Standardized with commas; lower, upper, and midpoint parsed |
| `price`                | DOUBLE           | USD retail price at snapshot time              | Coalesced to 0.0; used for free vs paid categorization      |
| `supported_languages`  | VARCHAR[]        | Listed text/interface localization languages   | Parsed, HTML decoded, deduplicated into `game_languages`    |
| `full_audio_languages` | VARCHAR[]        | Listed full voice audio localization languages | Parsed, normalized into `game_languages`                    |
| `genres`               | VARCHAR[]        | Steam genre tags                               | Parsed, deduplicated into `game_genres`                     |
| `windows`              | BOOLEAN          | Windows OS compatibility                       | Standardized boolean flag                                   |
| `mac`                  | BOOLEAN          | macOS compatibility                            | Standardized boolean flag                                   |
| `linux`                | BOOLEAN          | Linux / SteamOS compatibility                  | Standardized boolean flag                                   |
| `metacritic_score`     | BIGINT           | Metascore from Metacritic                      | Coalesced to 0 if missing                                   |
| `header_image`         | VARCHAR          | Steam CDN capsule/header banner URL            | Preserved; fallback displayed if unavailable                |
| `developers`           | VARCHAR[]        | Development studio(s)                          | Preserved for catalog details                               |
| `publishers`           | VARCHAR[]        | Publishing entity/entities                     | Preserved for catalog details                               |

---

## 2. Normalized DuckDB Schema

### 2.1 Table: `games`

Primary normalized catalog entity.

```sql
CREATE TABLE games (
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
);
```

### 2.2 Table: `languages`

Canonical language registry with catalog-wide aggregated frequencies.

```sql
CREATE TABLE languages (
    id VARCHAR PRIMARY KEY,         -- e.g. "simplified_chinese", "french"
    display_name VARCHAR,           -- e.g. "Simplified Chinese", "French"
    total_text_games BIGINT,        -- games listing language for text
    total_audio_games BIGINT        -- games listing language for audio
);
```

### 2.3 Table: `genres`

Canonical genre registry with catalog-wide aggregated frequencies.

```sql
CREATE TABLE genres (
    id VARCHAR PRIMARY KEY,         -- e.g. "action", "rpg"
    display_name VARCHAR,           -- e.g. "Action", "RPG"
    total_games BIGINT,
    total_eligible_games BIGINT     -- games with usable language metadata
);
```

### 2.4 Relational Bridge: `game_languages`

Captures the multidimensional localization relationship per game.

```sql
CREATE TABLE game_languages (
    game_id BIGINT,
    language_id VARCHAR,
    has_text BOOLEAN,
    has_audio BOOLEAN,
    raw_label VARCHAR,
    PRIMARY KEY (game_id, language_id)
);
```

### 2.5 Relational Bridge: `game_genres`

Captures multi-genre associations.

```sql
CREATE TABLE game_genres (
    game_id BIGINT,
    genre_id VARCHAR,
    display_name VARCHAR,
    PRIMARY KEY (game_id, genre_id)
);
```

### 2.6 Table: `data_quality_issues`

Auditable record of all detected anomalies and parsing flags.

```sql
CREATE TABLE data_quality_issues (
    id VARCHAR PRIMARY KEY,
    issue_type VARCHAR,             -- e.g. "missing_languages", "suspected_playtest"
    severity VARCHAR,               -- "info", "warning", "error"
    app_id BIGINT,
    game_name VARCHAR,
    description VARCHAR
);
```

---

## 3. Data Quality Flag Classifications

| Flag                 | Category       | Criteria                                                       | Impact                                         |
| :------------------- | :------------- | :------------------------------------------------------------- | :--------------------------------------------- |
| `missing_languages`  | Incompleteness | Source `supported_languages` is null or empty array            | Excluded from localization denominators        |
| `missing_genres`     | Incompleteness | Source `genres` is null or empty array                         | Excluded from genre-specific aggregations      |
| `missing_name`       | Incompleteness | Title string is null or whitespace                             | Flagged in quality audit; retained in database |
| `suspected_playtest` | Classification | Title contains `playtest`, `closed beta`, etc.                 | Excluded by default via global filters         |
| `audio_not_in_text`  | Inconsistency  | Language appears in `full_audio` but not `supported_languages` | Preserved in both relations; logged in audit   |
| `duplicate_appid`    | Integrity      | Identical App ID encountered multiple times                    | First occurrence kept; duplicates audited      |
