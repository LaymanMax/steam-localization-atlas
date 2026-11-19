# Research Methodology & Analytical Framework

## 1. Product Vision & Scope

The **Steam Localization Gap Atlas** is an open-source descriptive research tool designed to analyze the empirical distribution of language support listed across the Steam games catalog.

It systematically investigates two primary metadata fields:

- `supported_languages`: Languages listed as supported by a game, representing listed text, interface, and subtitle localization.
- `full_audio_languages`: Languages for which the game explicitly lists full voice audio support.

### The Central Research Question

> **How does language support differ across languages, genres, and estimated ownership groups—and where are the biggest differences between listed text/interface support and full audio support?**

---

## 2. Descriptive Research vs. Commercial Inference

> [!IMPORTANT]
> This application is strictly a descriptive research tool. It communicates **observed representation** across catalog metadata. It is **not** a sales forecasting, player demographic, or commercial market opportunity prediction tool.

- **No Regional Sales Claims:** The dataset does not contain geographic sales breakdowns, regional revenue, player demographics, or purchasing power by language.
- **No Untapped Opportunity Inferences:** Low representation of a language does not imply an untapped commercial market. Publishers localization choices are influenced by development budgets, localization costs (which vary heavily between text translation and full voice dubbing), genre conventions, and regional distribution channels.
- **Observed Metadata:** Percentages represent the proportion of games listing a language on their Steam store page.

---

## 3. Mathematical Definitions & Formulas

### 3.1 Listed Text/Interface Coverage

For any given language $L$ and genre $G$:

$$\text{Coverage}_{\text{text}}(L, G) = \frac{\text{Distinct eligible games in } G \text{ listing } L \text{ for text}}{\text{Distinct eligible games in } G \text{ with usable language metadata}} \times 100$$

### 3.2 Listed Full-Audio Coverage

$$\text{Coverage}_{\text{audio}}(L, G) = \frac{\text{Distinct eligible games in } G \text{ listing } L \text{ for full audio}}{\text{Distinct eligible games in } G \text{ with usable language metadata}} \times 100$$

### 3.3 Audio Conversion Rate Conditional on Text Support

Investigates how often a language listed for text/interface localization also receives full voice dubbing within genre $G$:

$$\text{Conversion}_{\text{audio}}(L, G) = \frac{\text{Distinct games in } G \text{ listing } L \text{ in full audio and text}}{\text{Distinct games in } G \text{ listing } L \text{ in text}} \times 100$$

### 3.4 Absolute Percentage-Point Gap

When comparing text coverage to audio coverage, we calculate the absolute difference in percentage points ($\text{pt}$), never relative percentage change:

$$\text{Gap}(L) = \text{Coverage}_{\text{text}}(L) - \text{Coverage}_{\text{audio}}(L)$$

_Example:_ If French has 25.8% text coverage and 5.0% audio coverage, the gap is **20.8 percentage points**, not a 20.8% decrease.

---

## 4. Denominators & Eligibility Criteria

1. **Genre-Specific Denominators:** Every cell in the localization matrix uses the total eligible game count of that specific genre as its denominator.
2. **Missing Metadata Exclusions:** Records missing language metadata or containing empty arrays are excluded from coverage denominators. They are counted separately in the Data Quality audit. Missing metadata is never treated as a confirmed negative case.
3. **Multi-Genre Counting:** A game belonging to both _Action_ and _RPG_ contributes once to the _Action_ denominator and once to the _RPG_ denominator. Multi-genre titles never count twice within the same genre. Consequently, the sum of genre denominators exceeds total unique games.

---

## 5. Language Normalization

The source dataset contains varying representations of languages (e.g. `Spanish - Spain`, `Spanish - Latin America`, `Simplified Chinese`, `Traditional Chinese`, `French`, HTML-encoded entities, asterisks, brackets, and newlines).

- **Variant Preservation:** Meaningful dialectal and regional distinctions are strictly preserved:
  - `Simplified Chinese` (`simplified_chinese`) is never merged into `Traditional Chinese` (`traditional_chinese`).
  - `Spanish - Spain` (`spanish_spain`) is distinct from `Spanish - Latin America` (`spanish_latam`).
  - `Portuguese - Brazil` (`portuguese_brazil`) is distinct from `Portuguese - Portugal` (`portuguese_portugal`).
- **Auditable Mapping:** Raw strings from the source are preserved alongside normalized identifiers.
- **Deduplication:** Repeated language entries within a single game are deduplicated.

---

## 6. Heuristics & Safeguards

- **Suspected Playtests:** Titles matching explicit cues (e.g. `playtest`, `closed beta`, `open beta`, `server test`) are identified. The global filter defaults to excluding them to maintain data integrity.
- **Small Sample Threshold:** Any cell or metric where the denominator is below the user-selected threshold (default $n < 30$) is visually flagged with a low-sample warning to prevent over-interpreting sparse figures.
- **Estimated Ownership Brackets:** Categorical brackets provided by the upstream dataset (`0 - 20,000`, `20,000 - 50,000`, etc.) are preserved verbatim. Midpoints are used solely for sorting.
