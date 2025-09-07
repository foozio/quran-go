# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## [0.2.0] - 2025-09-07
### Added
- SvelteKit web UI with Tailwind styling and live search.
- Dockerized web UI (`Dockerfile.svelte`) and compose service; API + Web run together.
- Top‑nav API status indicator (live ping + latency).
- Uthmanic Hafs Arabic font support (`web/static/fonts/uthmanic_hafs.*`), applied to Arabic text.
- Runtime `GET /stats` endpoint reporting total counts and per‑surah mismatches.
- Verification tool `cmd/quran-verify` and `make verify`; `make deploy` now verifies seeded DB.
- Deployment and how‑to docs (`docs/DEPLOYMENT.md`, `docs/HOWTO.md`).

### Changed
- Default web font for Arabic to KFGQPC Uthmanic Script HAFS (fallback to Amiri Quran).
- Docker Compose: split API and SvelteKit Web services; API image exposes only :8080.
- Per‑IP rate limiting via `golang.org/x/time/rate` and safer context‑aware DB calls.

### Fixed
- Seeding and ingestion for Surah 94–114; ensured complete ayah coverage.
- API `/surah/:n` struct scanning (no longer uses map scan that caused 500s).
- API `/search` JSON field names (now `surah`, `number`, `snip`).
- TUI list rendering and DB mapping; shows Surah on first paint.
- Compose/seed workflow to avoid stale DB copies (`make seed.data`).

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
