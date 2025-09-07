# Quran-Go

Beautiful, hackable Qur'an tooling in Go: REST API, Web UI, CLI, and TUI — backed by SQLite + FTS5 search. Seeded from the excellent `semarketir/quranjson` dataset.

<img alt="Full Snack Developers" src="docs/assets/full-snack-developers.png" height="72" />

![Go](https://img.shields.io/badge/Go-1.22-blue)
![CI](https://github.com/foozio/quran-go/actions/workflows/ci.yml/badge.svg)
![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)

## Features
- REST API (Gin) with simple CORS and rate limiting
- Web app (HTMX + Pico.css) with search, bookmarks, and notes
- Terminal apps: interactive TUI (Bubble Tea) and simple CLI
- Full‑text search over Arabic text and translation (SQLite FTS5)
- One‑command seeding from upstream JSON

## Apps
- `cmd/quran-api`: JSON API
- `cmd/quran-web`: Minimal server‑rendered HTMX UI
- `cmd/quran-cli`: Quick shell utility
- `cmd/quran-tui`: Interactive terminal UI

## Quick Start
Prerequisites: Go 1.22+, `make`, internet access for seeding.

```bash
make deps          # tidy modules
make seed          # downloads data + fills quran.db (ID translation)

# Optional: manage env securely
cp .env.example .env
dotenvx encrypt -f .env -o .env.encrypted

# Run something
make api           # starts REST API on :8080
# or
make web           # starts web UI on :8090
# or
make cli           # try: list | surah -n 2 | search Allah
# or
make tui           # open the terminal UI
```

## Documentation
- Deployment: see `docs/DEPLOYMENT.md`
- How-To and usage tips: see `docs/HOWTO.md`

## Configuration
These environment variables are recognized by the apps:
- `QURAN_DB_PATH`: path to SQLite DB (default: `quran.db`)
- `QURAN_BIND`: API bind address (default: `:8080`)
- `QURAN_ALLOWED_ORIGINS`: CORS origins (API)
- `QURAN_RATE_PER_MIN`: requests per minute (API)

Seeding uses Indonesian translation (`id`). You can change the language by editing `scripts/seed.go` to pass a different code to `data.IngestAll` (e.g., `"en"`).

## API Overview
- `GET /healthz` → `{ "ok": true }`
- `GET /surah` → list of surah metadata
- `GET /surah/:n` → ayah for a surah
- `GET /search?q=<query>` → FTS hits (Arabic/translation)

An OpenAPI sketch lives at `openapi.yaml`.

## Development
Common tasks via `Makefile`:
- `make deps` `make fmt` `make test` `make lint` `make vuln` `make sec`
- `make api` `make web` `make cli` `make tui` `make seed`

We use `pre-commit` (see `.pre-commit-config.yaml`). Install hooks with `make precommit`.

## Architecture
- SQLite schema in `internal/db/migrate.sql` (ayah table + FTS5 mirror)
- Data ingestion in `internal/data` (pulls from `semarketir/quranjson`)
- App code under `cmd/*` with shared helpers in `internal/*`

## gRPC (experimental)
The proto lives at `cmd/quran-grpc/proto/quran.proto`. Code‑gen wiring is TBD.

## Docker
Run with Docker Compose:
```bash
docker compose up -d --build
```

## Contributing
Welcomed! Please read `CONTRIBUTING.md` and abide by `CODE_OF_CONDUCT.md`.

## Changelog
See `CHANGELOG.md`.

## Credits
- Data: https://github.com/semarketir/quranjson

## License
MIT — see `LICENSE`.
