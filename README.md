# Quran-Go Monorepo

A beautifully crafted set of Go apps for learning Qur'an, powered by data from `semarketir/quranjson`.

## Apps
- CLI, TUI
- REST API (Gin) + OpenAPI
- gRPC
- Web (HTMX + Pico.css)
- Indexer/Seeder into SQLite
- .codex agents for code & security review

## Quickstart
```bash
make deps
make seed
cp .env.example .env
# fill env then encrypt:
dotenvx encrypt -f .env -o .env.encrypted
make api
```

---

![Go](https://img.shields.io/badge/Go-1.22-blue)
![CI](https://github.com/foozio/quran-go/actions/workflows/ci.yml/badge.svg)
![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)

## GitHub Setup (push this repo)
```bash
./scripts/init_github.sh your-github-username quran-go
# OR do it manually:
git init
git add .
git commit -m "chore: bootstrap quran-go monorepo"
git branch -M main
git remote add origin git@github.com:your-github-username/quran-go.git
git push -u origin main
```

### Publish to GitHub (HTTPS alternative)
```bash
git init
git add .
git commit -m "chore: bootstrap quran-go monorepo"
git branch -M main
git remote add origin https://github.com/your-github-username/quran-go.git
git push -u origin main
```
