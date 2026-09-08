# Repository Guidelines

## Project Structure & Module Organization
This is a personal Go-learning monorepo — there is no single top-level module; each project below has its own `go.mod`.

- `Project-Building-While-Learning/personal-analytics-backend/` — the actively developed project: a Go REST API with Postgres, Redis, JWT auth, webhooks (HMAC-signed), a worker pool, circuit breaker, and dead-letter queue for failed jobs. Entry point `cmd/server/main.go`; logic lives under `internal/` (`handlers`, `db`, `cache`, `redis`, `webhook`, `worker`, `retry`, `circuitbreaker`, `metrics`, `logger`, `config`, `models`). `learning-demos/` and `test-endpoints/` are standalone scripts excluded from the build.
- `Learning_Basics/` — standalone tutorial `.go` files grouped by topic (`01-basics` … `12-marshalling`), mostly single-file `go run` demos. Only `09-packages/{myapp,calculator}`, `Practice/`, and `KimiExcPractices/` are real modules with their own `go.mod`.
- `xk6-learning/GoLang/` — another single-file `go run` demo module (own `go.mod`), numbered `01_basics` … `17_mux`, written while reading through an external `wss-tools` codebase (progress tracked in its `REPO_COVERAGE_MAP.md`).
- `.github/skills/draft-post/scripts/` — a Claude Code skill: one Go CLI module per platform (`post-to-bluesky`, `post-to-devto`, `post-to-hashnode`, `post-to-linkedin`, `post-to-reddit`, `post-to-x`) that publishes drafted markdown from `posts/`; `post-to-substack` is a separate TypeScript/Node CLI.

## Build, Test, and Development Commands
`cd` into the relevant module first. For `personal-analytics-backend`:
- `go build ./cmd/... ./internal/...` — compile (mirrors CI; excludes `learning-demos/`, `test-endpoints/`)
- `go vet ./cmd/... ./internal/...`
- `go test ./cmd/... ./internal/...` — run all tests; add `-run TestName` to target one test, e.g. `go test ./internal/handlers/ -run TestRegister`
- `golangci-lint run` — static analysis (`.golangci.yml` excludes `learning-demos/`)
- `docker compose up` — full stack: app + Postgres + Redis + nginx/certbot
- Migrations live in `internal/db/migrations` (applied via `golang-migrate`)

## Coding Style & Naming Conventions
- Standard `gofmt` formatting; `golangci-lint` enforced via `.golangci.yml`.
- An ignored error must carry a `//nolint:errcheck` comment explaining why (e.g. `// cache is best-effort; circuit breaker handles Redis failures`) — bare ignored errors are flagged.
- Handlers depend on `Store`/`Cache` interfaces, not concrete DB types, so tests substitute `FakeStore`/`FakeCache` (`internal/handlers/fake_store_test.go`) instead of a real database.
- Prefer `make([]T, 0)` over `var s []T` for slices that get JSON-encoded, so empty results serialize as `[]` rather than `null`.

## Testing Guidelines
- Standard library `testing` package only (no testify); tests live in `*_test.go` next to the code they cover.
- Most handler tests run against `FakeStore`/`FakeCache`; files named `*_db_test.go` exercise the real database layer.
- CI (`.github/workflows/ci.yml`) runs build, vet, and test on every push/PR to `main`.

## Commit & Pull Request Guidelines
Commit subjects use short conventional-style prefixes seen in history: `feat:`, `fix:`, `refactor:`, `docs:`. Keep the subject line terse; put rationale or details in the body.
