# How to run reviewers

## 1) Prepare repo context
- Ensure `quran.db` exists: `make seed`
- Generate OpenAPI: already provided as `openapi.yaml`

## 2) Run code reviewer
Prompt:
> Use the **Go Code Reviewer** agent on this repository. Start with repo-wide architecture feedback, then deep dive into `internal/db`, `internal/data`, and `cmd/quran-api`. Include concrete diffs.

## 3) Run security reviewer
Prompt:
> Use the **Security Reviewer** agent on this repository. Confirm dotenvx is used (no plaintext `.env`), check CORS, rate limit, input validation, SQLite file permissions, Docker distroless, and CI secrets. Provide prioritized remediations.
