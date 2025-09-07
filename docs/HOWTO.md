Quran-Go How-To

Common tasks and tips for using Quran-Go locally and in Docker.

Prerequisites
- Go 1.22+
- make
- Internet access for seeding

Seed the Database
```
make deps
make seed   # Creates quran.db in repo root
```
Change Translation Language
- Edit `scripts/seed.go` and replace the language code in:
  `data.IngestAll(ctx, d, "id")` → e.g., `"en"`
- Re-run `make seed`

Run Locally (binaries)
```
# API
QURAN_DB_PATH=./quran.db go run ./cmd/quran-api

# Web
QURAN_DB_PATH=./quran.db go run ./cmd/quran-web

# CLI
QURAN_DB_PATH=./quran.db ./bin/quran-cli list
QURAN_DB_PATH=./quran.db ./bin/quran-cli surah -n 2
QURAN_DB_PATH=./quran.db ./bin/quran-cli search Allah

# TUI
QURAN_DB_PATH=./quran.db ./bin/quran-tui
```

API Endpoints (curl)
```
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/surah | jq '.[0]'
curl -s http://localhost:8080/surah/2 | jq
curl -s --get http://localhost:8080/search --data-urlencode q=Allah | jq
```

Docker (single container)
```
docker compose build
docker compose up -d
# API
curl -s http://localhost:8080/healthz
# Web
open http://localhost:8090/
```

Configuration
- `QURAN_DB_PATH`: SQLite database path (default varies by binary/image)
- `QURAN_BIND`/`QURAN_API_BIND`/`QURAN_WEB_BIND`: listening addresses
- `QURAN_ALLOWED_ORIGINS`: CORS (comma-separated) for API
- `QURAN_RATE_PER_MIN`: per-IP rate limit for API (default 120)

Development Workflow
```
make fmt      # format
make lint     # staticcheck
make test     # unit tests
make vuln     # govulncheck
make sec      # gosec
```

Troubleshooting
- “No surah found” in TUI/Web: Ensure `quran.db` exists and is mounted or point `QURAN_DB_PATH` correctly.
- 400 on `/surah/:n`: Number must be between 1 and 114.
- 400 on `/search`: Query max length is 100 characters.
- Rate limit 429: Increase `QURAN_RATE_PER_MIN` or test from fewer IPs.
