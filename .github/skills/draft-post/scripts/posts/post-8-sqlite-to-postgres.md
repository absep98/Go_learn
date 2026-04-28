---
title: I Built My Backend on SQLite. Then I Deleted It.
subtitle: Part 8 of Learning Go in Public
tags: go, postgres, buildinpublic, backend
series: Learning Go in Public
---

In [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g), I taught my server how to die gracefully. It felt production-ready. Auth, rate limiting, caching, clean shutdown — the API was doing real work.

I said Part 8 would be "Docker + Postgres." I was wrong. Postgres alone deserved its own post, because swapping the database under a running backend taught me more in a weekend than the three weeks before it.

Then I tried to run my tests in CI.

And everything broke.

## The SQLite Honeymoon

When I started this project, SQLite was the obvious choice:

```go
import _ "modernc.org/sqlite"

DB, err = sql.Open("sqlite", "./data.db")
```

One file. No server. No Docker. No connection strings. Pure Go driver — not even CGO. It felt like cheating. I built auth, entries, pagination, rate limiting, and caching all on top of a single `.db` file sitting next to my binary.

For 7 posts, SQLite carried this backend beautifully.

Then I wrote this in my GitHub Actions workflow:

```yaml
- name: Run tests
  run: go test ./...
```

And the CI logs lit up red. Concurrent test runs stepping on each other. File locks. Tests that passed locally failed in CI. Tests that passed in CI failed on the next run.

SQLite isn't broken. But it's a single-writer database pretending to be a multi-user one. The moment my test suite started running handlers in parallel, the cracks showed.

## Day 23: The Choice

I had two options:

1. **Make SQLite work** — serialize tests, mock the DB layer, fight the concurrency model
2. **Switch to Postgres** — the thing I'd use in production anyway

I picked #2. Not because SQLite is bad, but because I was lying to myself. Every tutorial that says "SQLite is fine for production" forgets to mention the asterisk: *unless you need concurrent writes, CI pipelines, connection pooling, or anything resembling a real deployment.*

If I was going to deploy this, it was going to hit Postgres eventually. Better to feel that pain on Day 23 than on launch day.

## The Migration That Looked Simple

On paper, swapping databases in Go looks trivial. `database/sql` is an interface. Change the driver, change the connection string, done.

In reality, the diff was 11 files and 708 lines.

**Step 1: Swap the driver.**

```go
// Before
_ "modernc.org/sqlite"
DB, _ = sql.Open("sqlite", "./data.db")

// After
_ "github.com/lib/pq"
DB, _ = sql.Open("postgres", dsn)
```

**Step 2: Build a DSN instead of a file path.**

SQLite takes a filename. Postgres takes a connection string with five moving parts:

```go
dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
```

My config struct grew from `DBPath string` to five separate env vars. Every one of them needs a default, a validation, and a place to live in `.env`, `docker-compose.yml`, and CI.

**Step 3: Rewrite every query.**

This is where I got humbled. SQLite uses `?` placeholders. Postgres uses `$1`, `$2`, `$3`:

```go
// SQLite
query := `INSERT INTO entries (user_id, text, mood, category) VALUES (?, ?, ?, ?)`

// Postgres
query := `INSERT INTO entries (user_id, text, mood, category) VALUES ($1, $2, $3, $4)`
```

Every `WHERE user_id = ?` had to become `WHERE user_id = $1`. Miss one, and you get a runtime error that only surfaces when that specific handler runs.

**Step 4: Rewrite every schema.**

SQLite is loose with types. Postgres is strict.

```sql
-- SQLite
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Postgres
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW()
);
```

`AUTOINCREMENT` → `SERIAL`. `DATETIME` → `TIMESTAMP`. `CURRENT_TIMESTAMP` → `NOW()`. Small changes. Each one a landmine if you forget.

## The One That Stung

`LastInsertId()`.

In SQLite:

```go
result, err := DB.ExecContext(ctx, query, userID, text, mood, category)
id, err := result.LastInsertId()  // works
```

In Postgres: **that method returns an error.** The `lib/pq` driver doesn't support it. Because Postgres doesn't know what the "last" insert was — there's no global counter like SQLite has.

The Postgres way is to ask the database to hand back the ID as part of the insert:

```go
query := `INSERT INTO entries (user_id, text, mood, category)
          VALUES ($1, $2, $3, $4) RETURNING id`

var id int64
err := DB.QueryRowContext(ctx, query, userID, text, mood, category).Scan(&id)
```

`RETURNING id` is Postgres-specific. It's also beautiful — one round trip instead of two, and you get the ID guaranteed to be the one you just inserted, even under concurrent load.

This is the kind of thing a tutorial never teaches you. You have to migrate a database to discover it.

## The CI Rewrite

The last piece was GitHub Actions. My CI was running tests against SQLite files. I needed a real Postgres instance for every test run.

```yaml
services:
  postgres:
    image: postgres:16
    env:
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
      POSTGRES_DB: testdb
    ports:
      - 5432:5432
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
```

GitHub Actions spins up a Postgres container as a sidecar to my test job. The health check waits for it to be ready. My tests run against a real database. If they pass in CI, they'll pass in production — because the database is the same.

That felt like a milestone.

## What This Actually Taught Me

I thought this was a migration. It was a lesson in what "production-ready" actually means.

**SQLite is incredible — for prototypes.** Fast, simple, zero setup. For learning Go, it removed every obstacle between me and my first working REST API.

**Postgres is what production looks like.** Connection pooling, concurrent writes, real constraints, explicit types, dedicated container in CI. Every "pain point" I hit was a concept I needed to learn anyway — just sooner than I wanted to.

The biggest mindset shift: **my tests got honest.** A test suite running against Postgres in CI is testing my code. A test suite running against SQLite was testing a fiction — a simpler world where concurrency didn't exist.

If you're building something you actually want to deploy, migrate early. The longer you wait, the more queries you write, the more schemas you create, the more rewriting you'll do.

I lost a weekend to this migration.

I gained a backend I can actually ship.

---

Up next: with Postgres in place, I finally tackled the thing I'd been avoiding for weeks — **proper database migrations with `golang-migrate`**. Because writing `CREATE TABLE IF NOT EXISTS` in Go code was a sin I was ready to confess.

*This is Part 8 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i) | [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g)*
