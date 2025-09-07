# Contributing

Thanks for your interest in contributing! This guide outlines how to propose changes and get them merged.

## Getting Started
- Install Go 1.22+, `make`, and optionally `pre-commit` and `dotenvx`.
- Clone the repo and bootstrap:
  - `make deps`
  - `make seed` (creates `quran.db`)
  - `make api` or `make web` to run locally

## Development Workflow
- Create a branch from `main` using a descriptive name, e.g. `feat/tui-search`, `fix/api-cors`.
- Keep changes focused and small. Update docs when behavior changes.
- Run checks before pushing:
  - `make fmt` — gofmt
  - `make lint` — staticcheck
  - `make test` — unit tests
  - `make vuln`/`make sec` — basic security scans
- Optional: install hooks with `make precommit`.

## Commit Messages
- Follow conventional hints when possible: `feat:`, `fix:`, `docs:`, `refactor:`, `chore:`, `test:`.
- Use the imperative mood (“add …”, “fix …”).

## Pull Requests
- Open a PR against `main` with:
  - A clear description of what and why
  - Screenshots or logs where helpful
  - Notes on breaking changes or migrations
- Keep PRs small; split large ones.
- Draft PRs are welcome for early feedback.

## Code Style
- Prefer clarity over cleverness.
- Keep changes minimal and localized; match existing patterns and structure.
- Avoid introducing new dependencies unless necessary.

## Testing
- Add or update tests where it makes sense. Start from the smallest unit you touched.
- If no test harness exists for an area, document manual test steps in the PR.

## Security
- Do not publicly disclose vulnerabilities. See `SECURITY.md` for reporting instructions.

## Documentation
- Update `README.md`, `openapi.yaml`, and comments when behavior changes.
- Add usage examples for new CLI/TUI commands or API endpoints.

## License
By contributing, you agree that your contributions will be licensed under the project’s MIT license.
