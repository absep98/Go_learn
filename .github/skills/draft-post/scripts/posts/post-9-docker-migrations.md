---
title: My Database Disappeared Every Time I Restarted Docker. Migrations Fixed That.
subtitle: Part 9 of Learning Go in Public
tags: go, docker, buildinpublic, backend
series: Learning Go in Public
---

In [Part 8](https://dev.to/abhishek_sharma_a9792aee8/i-built-my-backend-on-sqlite-then-i-deleted-it-15a), I ripped out SQLite and switched to Postgres. The database was real now — but it still only ran on my laptop.

Three terminal tabs. One for `go run`, one for Redis, one for Postgres. If I closed any of them, the whole thing fell over.

It was time to Dockerize.

## The First Dockerfile: 800MB of Regret

My first attempt was a single-stage Dockerfile. Copy everything, build, run:

```dockerfile
FROM golang:1.25-alpine
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server
CMD ["/app/server"]
```

It worked. The image was 800MB. That's bigger than most operating systems. I was shipping the entire Go compiler, my source code, and every dependency just to run a 15MB binary.

## Multi-Stage Builds: The Aha Moment

The fix is a pattern called multi-stage builds. Two `FROM` statements in one Dockerfile:

```dockerfile
# STAGE 1: Build (heavy — has Go compiler)
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# STAGE 2: Run (tiny — just Alpine + binary)
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 8080
CMD ["/app/server"]
```

Stage 1 is a construction site — Go compiler, 300MB of build tools, your source code. It compiles your app into a single binary.

Stage 2 is the delivery box — a bare Alpine Linux image (~5MB) with just your binary copied in. The entire build stage gets thrown away.

**Result: ~15MB.** From 800MB to 15MB. Same binary. Same functionality. One line did the heavy lifting:

```dockerfile
COPY --from=builder /app/server /app/server
```

That `--from=builder` is reaching back into the dead first stage and grabbing the one file that matters.

### The CGO Trap

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
```

`CGO_ENABLED=0` tells Go: don't use any C libraries. Without it, the binary links against `glibc` — but Alpine uses `musl`. The binary compiles fine in the builder, then crashes in the runner with a cryptic error. I learned this the hard way.

This produces a **static binary** — completely self-contained. It doesn't need Go, doesn't need C, doesn't need anything except a Linux kernel to run.

## Docker Compose: One Command to Rule Them

My backend needs three services: Go app + Postgres + Redis. Without docker-compose, starting the stack looks like:

```bash
docker network create mynet
docker run -d --name postgres --network mynet -e POSTGRES_USER=user ...
docker run -d --name redis --network mynet redis:alpine
docker build -t myapp . && docker run --network mynet -e DB_HOST=postgres ...
```

With docker-compose:

```bash
docker compose up
```

That's it. One command. Here's the full stack definition:

```yaml
services:
  app:
    build: .
    restart: on-failure
    ports:
      - "8080:8080"
    environment:
      - JWT_SECRET=supersecretkey
      - REDIS_HOST=redis
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=user
      - DB_PASSWORD=password
      - DB_NAME=analytics
    depends_on:
      redis:
        condition: service_started
      postgres:
        condition: service_healthy

  redis:
    image: redis:alpine
    ports:
      - "6379:6379"

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: analytics
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d analytics"]
      interval: 5s
      timeout: 5s
      retries: 5
```

Three things I didn't know before writing this:

**Docker Compose networking is automatic.** The service name *is* the hostname. My Go app connects to `redis` and `postgres` — not `localhost`. Docker creates a virtual network and resolves service names to container IPs. No manual network setup.

**`depends_on` isn't enough.** Just because Postgres starts doesn't mean it's *ready*. `depends_on` with `condition: service_healthy` waits for the healthcheck to pass. Without it, my Go app would boot, try to connect, fail, and crash — 50% of the time.

**`restart: on-failure` is a safety net.** If the app crashes during the startup race (Redis isn't ready, Postgres is still initializing), Docker restarts it automatically. Crude but effective.

## The Problem Nobody Mentions

I ran `docker compose up`. Everything worked. Tables created. API responded. I inserted some entries, tested auth — beautiful.

Then I ran `docker compose down` and `docker compose up` again.

**Tables were gone.** Data gone. Schema gone. Every restart was a blank slate.

My `CREATE TABLE IF NOT EXISTS` code ran on every boot, so the tables came back — but this was hardcoded SQL strings inside Go functions. The moment I needed to add a column, rename a field, or create a new table, I'd have to manually edit Go code and hope the `IF NOT EXISTS` didn't silently ignore my changes.

This is the problem database migrations solve.

## Migrations: Schema as Code

Instead of SQL strings buried in Go code, I created a `migrations/` folder:

```
internal/db/migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_entries.up.sql
├── 000002_create_entries.down.sql
```

Each migration is a pair: `up` (apply) and `down` (rollback).

```sql
-- 000001_create_users.up.sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 000001_create_users.down.sql
DROP TABLE IF EXISTS users;
```

Simple SQL files. Numbered. Versioned. Reviewable in a PR.

## golang-migrate: The Runner

I used `golang-migrate` to execute these files on startup:

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
    m, err := migrate.New(
        "file://internal/db/migrations",
        dsn,
    )
    if err != nil {
        return fmt.Errorf("creating migrator: %w", err)
    }
    defer m.Close()

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("running migrations: %w", err)
    }

    slog.Info("Database migrations applied")
    return nil
}
```

`m.Up()` runs every unapplied migration in order. If all migrations are already applied, it returns `migrate.ErrNoChange` — which we ignore. Idempotent. Safe to run on every boot.

Behind the scenes, `golang-migrate` creates a `schema_migrations` table in Postgres that tracks which version has been applied:

```
version | dirty
--------+------
      2 | false
```

Version 2 means both `000001` and `000002` have run. `dirty: false` means no migration failed halfway. If I add `000003_add_tags_column.up.sql` tomorrow and restart — only migration 3 runs. The first two are skipped.

## The DSN Gotcha

One thing that tripped me up: `database/sql` and `golang-migrate` expect *different* connection string formats.

```go
// database/sql wants key=value format
dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

// golang-migrate wants URL format
migrateDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
    cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

err = db.InitDB(dsn, migrateDSN)
```

Same credentials, different format. I spent 30 minutes debugging "invalid connection string" before I realized the migrate driver expects a URL, not key-value pairs.

## The Dockerfile Update

Migrations are SQL files on disk. The runner stage needs them:

```dockerfile
COPY --from=builder /app/server /app/server
COPY --from=builder /app/internal/db/migrations /app/internal/db/migrations
```

Without this line, the container boots, tries to read `file://internal/db/migrations`, finds nothing, and silently starts with no tables.

## Before and After

**Before:**
```
Terminal 1: docker run postgres ...
Terminal 2: docker run redis ...
Terminal 3: go run cmd/server/main.go
Schema: hardcoded in Go functions
Adding a column: edit Go code, pray
```

**After:**
```
docker compose up
Schema: versioned SQL files
Adding a column: write a new .sql file, restart
```

## What I Learned

**Multi-stage builds are non-negotiable.** Every Go project should ship a ~15MB image, not an 800MB one. The pattern is always the same: build in `golang:alpine`, run in `alpine:latest`.

**Docker Compose replaces three terminal tabs with one command.** Service names become hostnames. Health checks prevent race conditions. `restart: on-failure` handles the rest.

**`CREATE TABLE IF NOT EXISTS` is not a migration strategy.** It works until you need to change something. Real migrations are numbered SQL files, tracked by a version table, applied in order. The tooling exists. There's no excuse to skip it.

**Schema should live next to code, not inside it.** When your table definitions are `.sql` files in a `migrations/` folder, they show up in diffs, get reviewed in PRs, and can be rolled back. When they're string literals inside Go functions, they're invisible.

---

Up next: the thing I kept running into while building all of this — **I couldn't debug anything without proper logging.** So I replaced `log.Println` with structured JSON logging, added request IDs for tracing, and built a metrics endpoint. The week my backend learned to explain itself.

*This is Part 9 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i) | [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g) | [Part 8](https://dev.to/abhishek_sharma_a9792aee8/i-built-my-backend-on-sqlite-then-i-deleted-it-15a)*
