Quran-Go Deployment Guide

This guide covers running Quran-Go in production with Docker. You can deploy the API and Web together in a single container (recommended for simplicity) or separately.

Artifacts
- Combined image: `Dockerfile.all` (API + Web)
- API-only image: `Dockerfile`
- Web-only image: `Dockerfile.web`
- Compose file: `docker-compose.yml`

Quick Start (single container)
1) Seed the database on the host (first time only):
```
make deps
make seed   # creates ./quran.db
```
2) Build and run the combined service:
```
docker compose build
docker compose up -d
```
3) Verify health:
- API: http://localhost:8080/healthz
- Web: http://localhost:8090/

Images
- Build combined image directly:
```
docker build -t quran-all:local -f Dockerfile.all .
```
- Build individual images (optional):
```
# API
docker build -t quran-api:local -f Dockerfile .
# Web
docker build -t quran-web:local -f Dockerfile.web .
```

Environment Variables
- `QURAN_DB_PATH` (default: `/data/quran.db`)
- `QURAN_API_BIND` (combined; default `:8080`)
- `QURAN_WEB_BIND` (combined; default `:8090`)
- `QURAN_BIND` (API-only or Web-only images)
- `QURAN_ALLOWED_ORIGINS` (API CORS; comma-separated; default `*`)
- `QURAN_RATE_PER_MIN` (API rate limit per IP; default `120`)

Volumes and Data
- The images declare `VOLUME /data` and expect a SQLite file at `/data/quran.db`.
- With Compose, a named volume `quran_data` is attached to `/data`.
- Seed the DB on the host (or inside a one-off container) and mount it read-only in production if desired.

Healthchecks
- Images include HEALTHCHECKs using the app binaries (`-selfcheck`).
- The combined image checks the API endpoint.

Reverse Proxy (optional)
- Terminate TLS and route with Nginx/Caddy/Traefik.
- Sample Nginx locations:
```
location /api/ { proxy_pass http://127.0.0.1:8080/; }
location / { proxy_pass http://127.0.0.1:8090; }
```

Production Considerations
- Enable TLS at the proxy; the app itself serves HTTP only.
- Keep SQLite file on persistent storage; back up regularly.
- Set strict CORS (`QURAN_ALLOWED_ORIGINS`) for public deployments.
- Tune rate-limiting via `QURAN_RATE_PER_MIN`.
- Resource limits: configure CPU/memory limits in Compose/K8s.
- K8s: create two Services in one Pod (single container) or two Deployments (API/Web split).

Seeding in a Container (optional)
You can populate the DB using a one-off tooling container:
```
# Build a lightweight Go builder container and run seed (example)
docker run --rm -v "$(pwd)":/src -w /src golang:1.22 \
  bash -lc 'go mod download && go run ./scripts/seed.go'
# Then start the app normally
```

Troubleshooting
- Empty lists in the Web UI or TUI usually mean the DB was not seeded or not mounted correctly.
- API 500 errors typically reflect DB path/mount issues; check `QURAN_DB_PATH`.
- Healthcheck failing: ensure port bindings are correct and the service started.
