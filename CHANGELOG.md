# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## [0.1.0] - 2025-09-07
### Added
- Initial public release of Quran-Go.
- REST API (`/surah`, `/surah/:n`, `/search`, `/healthz`).
- Web UI (HTMX + Pico.css) with basic search, bookmarks, and notes.
- CLI (`list`, `surah -n N`, `search <q>`).
- TUI (Bubble Tea) with Surah list and detail views.
- SQLite schema with FTS5 search and triggers.
- Seeder pulling data from `semarketir/quranjson` (Indonesian translation by default).
- Basic middleware for CORS and rate limiting.

### Changed
- N/A

### Fixed
- N/A

### Security
- N/A

