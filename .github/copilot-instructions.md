# AI Agent Instructions

## Code Style

Follow [functions.go](../01-basics/functions.go) and [structs.go](../03-data-structures/structs.go) for Go formatting patterns. Use short variable names in loops (`i`, `n`), descriptive names in larger scope (`userID`, `passwordHash`). Always use `err` for errors, `ok` for comma-ok idioms.

## Architecture

This workspace has two contexts:
- **Learning examples** (01-basics, 02-control-flow, etc.): Single-file demos with extensive inline comments explaining concepts
- **Production backend** ([personal-analytics-backend](../Project-Building-While-Learning/personal-analytics-backend/)): Clean architecture with `cmd/server/` (entry), `internal/` (handlers→db→models layers). See [ARCHITECTURE-MAP.md](../Project-Building-While-Learning/personal-analytics-backend/ARCHITECTURE-MAP.md)

Backend uses middleware chaining: `RequestIDMiddleware(RateLimitMiddleware(LoggingMiddleware(AuthMiddleware(handler))))` - see [main.go](../Project-Building-While-Learning/personal-analytics-backend/cmd/server/main.go) for ASCII flow diagram.

## Build and Test

```bash
# Learning examples - single file execution
go run 01-basics/functions.go

# Backend server
cd Project-Building-While-Learning/personal-analytics-backend
go run cmd/server/main.go

# Run Go unit tests (handlers package)
cd Project-Building-While-Learning/personal-analytics-backend
go test ./internal/handlers/ -v

# Dependencies (backend uses separate module)
go mod download
```

Go unit tests exist in `internal/handlers/entries_test.go` — uses real Postgres via env vars, stub Redis client, and httptest. Integration testing also done via PowerShell scripts in [test-endpoints/](../Project-Building-While-Learning/personal-analytics-backend/test-endpoints/).

## Project Conventions

**Error handling** (critical pattern found everywhere):
```go
result, err := someFunction()
if err != nil {
    slog.Error("operation failed", "error", err)  // structured logging
    return fmt.Errorf("context: %w", err)         // error wrapping
}
```
Always check errors immediately, return early on failure.

**HTTP handlers** - sequential validation with early returns ([entries.go](../Project-Building-While-Learning/personal-analytics-backend/internal/handlers/entries.go)):
1. Validate HTTP method
2. Parse JSON request
3. Validate required fields (return 400 for each failure)
4. Process and respond with JSON

**JSON struct tags** - consistent pattern in [models.go](../Project-Building-While-Learning/personal-analytics-backend/internal/models/models.go):
- Use `json:"field_name"` for all exported fields
- Use `json:"-"` for sensitive data like `PasswordHash`
- Use `json:"field,omitempty"` for optional fields

**Context usage** - request-scoped data passed via context.Context:
```go
userID := r.Context().Value("user_id").(int64)  // set by AuthMiddleware
requestID := r.Context().Value("request_id").(string)  // set by RequestIDMiddleware
```

## Integration Points

**Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO). Global `DB *sql.DB` in [db.go](../Project-Building-While-Learning/personal-analytics-backend/internal/db/db.go). Always use parameterized queries (`?` placeholders) to prevent SQL injection.

**Redis**: Rate limiting and caching. Separate client in [redis.go](../Project-Building-While-Learning/personal-analytics-backend/internal/redis/redis.go).

**HTTP Server**: Standard library only, no frameworks. Route registration in [main.go](../Project-Building-While-Learning/personal-analytics-backend/cmd/server/main.go).

## Security

**Authentication**: JWT with bcrypt password hashing (cost=10). Token format: `Authorization: Bearer <token>`. Middleware extracts user_id and adds to context - see [auth.go](../Project-Building-While-Learning/personal-analytics-backend/internal/handlers/auth.go) and [middleware.go](../Project-Building-While-Learning/personal-analytics-backend/internal/handlers/middleware.go).

**Logging**: Use `log/slog` with JSON output. Pattern: `slog.Info("msg", "key", val)`. Never log passwords/tokens - use `json:"-"` struct tags.

**Input validation**: Sequential checks with early returns. Examples: email contains `@`, password ≥6 chars, mood range 1-10.

## Agent Behavior (Karpathy Principles)

**Think Before Coding** — Before implementing anything non-trivial:
- State assumptions explicitly. If uncertain, ask rather than guess.
- If multiple interpretations exist, present them — don't pick silently.
- If a simpler approach exists, push back and say so.
- If something is unclear, name what's confusing and ask.

**Goal-Driven Execution** — For multi-step tasks, define success criteria first:
- Transform "add X" into "write tests for X, then make them pass"
- State a brief plan with verification steps before starting
- Example: `1. Add DB query → verify: returns correct rows. 2. Add handler → verify: 200 with correct JSON`
- Strong success criteria enable independent looping. Weak criteria ("make it work") waste iterations.

## Self-Improvement Protocol

This agent learns from **user corrections and repeated mistakes** across sessions using `/memories/repo/` files.

The goal is NOT to document codebase facts (that's what `copilot-instructions.md` is for).
The goal IS to stop repeating the same mistakes the user has already corrected.

### Pollution Prevention (CRITICAL)

**Memory files cost tokens on every interaction.** Over-capturing degrades Copilot performance.

**Hard limits:**
- Max 15 files in `/memories/repo/` total
- Max 10 lines of JSON content per file
- If at limit, must DELETE an outdated/lower-value memory before creating a new one

**Before creating ANY memory, ask 3 questions:**
1. Is this already covered in `copilot-instructions.md`? → DON'T capture
2. Is this a one-time fix or a recurring pattern? → Only capture recurring patterns
3. Will this still be true in 3 months? → If unsure, don't capture

**Default to NOT capturing.** Memory is for high-signal corrections only, not for safety-net documentation.

### How Repo Memory Works
- Each learning is stored as a **separate JSON file** via `create` (append-only, no edits)
- File naming: `/memories/repo/<category>-<short-description>.json`
- JSON format: `{ "subject", "fact", "citations", "reason", "category" }`

### Categories

| Category | When to Capture | Example filename |
|----------|----------------|------------------|
| `mistake` | User explicitly corrects agent behavior 2+ times | `mistake-dont-use-global-db-in-tests.json` |
| `pattern` | User explicitly says "always do X" or rejects an alternative | `pattern-prefer-fakes-over-mocks.json` |
| `debugging` | Non-obvious root cause that took >30min to find | `debugging-docker-mod-cache-stale.json` |

Note: `testing` removed as a separate category — fold into `pattern` if it's a preference.

### When to CAPTURE (high bar — most things should NOT be captured)
- **User explicitly says "remember this" or "don't do this again"** → capture
- **User has corrected the same mistake 2+ times in the same session** → capture
- **A debugging root cause is genuinely non-obvious AND likely to recur** → capture

### When NOT to CAPTURE
- A one-off correction (typo, wrong file path) — just fix it and move on
- Anything documented in `copilot-instructions.md`, `ARCHITECTURE-MAP.md`, or other repo docs
- Personal preferences about commit messages, formatting, or style (those go in instructions)
- Facts about external libraries (those are in their docs)

### When to CONSULT existing memories
- **Before writing test code** → quick scan of memory file names
- **When debugging failures** → check for matching `debugging-*` files
- **When user says "we fixed this before"** → search all memory files
- Skip consultation for trivial tasks — don't waste tokens

### Pruning Protocol (run when at 15-file limit)
Before creating a new memory file, evaluate existing ones and DELETE one that is:
1. Contradicted by newer knowledge or the latest code
2. Now covered in `copilot-instructions.md` (it graduated)
3. Hasn't been useful in 3+ months (no citations referenced)
4. Lower-value than the new memory you're about to create

### Rules
- Only store **corrections and preferences** that came from user interaction
- Keep JSON content under 10 lines per file
- Never store secrets, credentials, or user-specific data
- Include `citations` (file paths) so future lookups have context
- If in doubt, DON'T create a memory file
