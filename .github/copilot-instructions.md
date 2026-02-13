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

# Dependencies (backend uses separate module)
go mod download
```

No Go unit tests - testing done via PowerShell scripts in [test-endpoints/](../Project-Building-While-Learning/personal-analytics-backend/test-endpoints/).

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
