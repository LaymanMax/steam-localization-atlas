# Contributing to Steam Localization Gap Atlas

We welcome contributions of all kinds, including bug reports, documentation improvements, dataset parsing refinements, and feature enhancements.

---

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](./CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

---

## Development Workflow

### 1. Prerequisites

- **Go:** 1.22+
- **Node.js:** 20+
- **pnpm:** 9+

### 2. Fork & Clone

```bash
git clone https://github.com/laymanmax/steam-localization-atlas.git
cd steam-localization-atlas
```

### 3. Setup Dependencies

```bash
make setup
```

### 4. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 5. Running the Application Locally

```bash
# Ingest the dataset (or use an existing local Parquet file)
make ingest

# Start the Go server (serves API and frontend)
make dev
```

Visit `http://localhost:8080` in your browser.

---

## Automated Testing & Validation

Before submitting a pull request, ensure all backend and frontend tests pass:

```bash
# Run Go unit tests and frontend test suite
make test

# Run dataset validation checks
make validate

# Verify production builds
make build
```

---

## Contribution Guidelines

1. **Analytical Integrity:**
   - Calculations, denominators, and aggregates must remain in the Go backend (`internal/analytics`).
   - Maintain the distinction between _observed metadata representation_ and unsupported claims of commercial market size or player demand.
2. **Language Normalization:**
   - Preserve regional variants (Simplified Chinese vs. Traditional Chinese, Brazilian Portuguese vs. European Portuguese, European Spanish vs. Latin American Spanish).
   - Test new language mappings in `internal/normalization/normalization_test.go`.
3. **Commit Messages:**
   - Write clear, imperative commit messages (e.g., `feat: add release period filter to gap ranking`, `fix: handle edge case in delimiter parsing`).
4. **Pull Requests:**
   - Reference any relevant issues.
   - Summarize the changes made, the motivation, and test coverage provided.
