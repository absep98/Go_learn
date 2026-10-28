# Engineering notes: correctness & concurrency hardening (Oct 2026)

This page records one focused pass over the service. It covers what was broken, how each defect
was proven, how it was fixed, and what is still weak. Every number below is something I ran and
observed. The commands are included so anyone can reproduce them.

## Where it started

- **CI had been red since 2026-09-08.** GitHub Actions run `34210588801` failed at `go vet`:
  `internal/db/db_test.go:76:19: not enough arguments in call to InitDB`. The DB tests were never
  updated after the SQLite → Postgres migration. Because vet failed, the *Test* and *Lint* steps
  had never run against the Postgres-era code.
- **The DB tests couldn't have passed anyway.** Migrations were loaded from
  `file://internal/db/migrations`, which is relative to the working directory. `go test` runs from
  each package directory, so with a real database both `TestMain`s panicked with `failed to open source`.
- **Lint would have failed next.** The workflow used `golangci-lint-action@v6` with
  `version: latest`. "latest" is now v2, which refuses the v1 config (`unsupported version of the configuration`).
- **No test exercised the concurrent paths.** The existing suite was race-clean only because nothing
  ran the racy code. That set the method below: write a test that makes the bug observable
  (usually under `-race`), watch it fail, fix it, and watch it pass repeatedly.

## Defects found and fixed

| # | Defect | How it was proven (before the fix) | Fix |
|---|---|---|---|
| 1 | **Timeout middleware data race.** The handler ran in a goroutine; on timeout the middleware wrote a 504 to the *same* `ResponseWriter` the handler was still writing to. | `timeout_test.go` under `-race`: 4 `DATA RACE` reports. The client received **status 200 with body `late bodyRequest timed out`**. | `http.TimeoutHandler`: buffers handler output behind a mutex; late writes get `ErrHandlerTimeout`. |
| 2 | **One handler panic crashed the whole server.** The panic happened off the serving goroutine, so `net/http`'s per-connection `recover()` never saw it. | `TestTimeoutMiddleware_HandlerPanicSurfacesOnServingGoroutine`: `panic: boom` killed the test binary. | Same fix. `TimeoutHandler` re-panics on the serving goroutine, where `net/http` recovers it. |
| 3 | **Refresh-token replay.** `/refresh` did `GET` then `DEL`. Each is atomic, the pair isn't: N concurrent replays of one token all read it before any deleted it. | `TestRefresh_ConcurrentReplayOfOneTokenSucceedsOnce`, with a barrier forcing the worst-case interleaving: **10 of 10 replays accepted.** | One atomic `GETDEL` (Redis ≥ 6.2). Logout uses it too. Real-Redis test: 20 rounds × 50 goroutines, exactly one winner each round. |
| 4 | **Shutdown could return with jobs still queued.** Workers called `wg.Add(1)` *after* receiving a job, so `Wait()` could see 0 while jobs were buffered. | `-race`: `DATA RACE` plus the runtime's own panic, `sync: WaitGroup is reused before previous Wait has returned`. | `wg.Add(numWorkers)` before starting workers; each worker calls `defer wg.Done()` and exits only when the closed queue is empty. |
| 5 | **`AddJob` after shutdown panicked (`send on closed channel`).** A handler that outlived its request timeout can still enqueue after `close(JobQueue)`. | `TestAddJob_AfterDrain…`: panic. `TestAddJob_ConcurrentWithDrain…`: **49 of 50 concurrent `AddJob` calls panicked.** | `RWMutex` + `queueClosed` flag. `AddJob` checks-and-sends under the read lock and drops the job (logged) once closed. |
| 6 | **`/failed-jobs` was public, and dead-letter payloads contain users' journal text.** | E2E: unauthenticated `GET /failed-jobs` → **200**, payload contained my test entry `PRIVATE-JOURNAL-b40465` (base64-encoded, which is encoding, not protection). | `AdminOnly` middleware: `X-Admin-Token` checked with `subtle.ConstantTimeCompare`; endpoint returns `404` unless `ADMIN_TOKEN` is set (fail closed). |
| 7 | **SIGTERM skipped graceful shutdown.** Only `os.Interrupt` was handled; `docker stop`/Kubernetes send SIGTERM, which fell through to Go's default handler (immediate exit). | E2E: `docker stop` → process gone, **0** shutdown log lines. | `signal.Notify(quit, os.Interrupt, syscall.SIGTERM)`. |
| 8 | **No server timeouts** (Slowloris-style connection exhaustion). | E2E: a connection that never finishes its headers was **still open after 15s**. | `ReadHeaderTimeout` 5s, `ReadTimeout` 15s, `WriteTimeout` = request timeout + 5s (so the 503 can still be written), `IdleTimeout` 60s. |
| 9 | **Database password written to logs** on every start (`slog.Info("Database connected", "path", dsn)`). | `TestInitDB_DoesNotLogPassword` found `password=…` in captured logs. | Log without the DSN. |
| 10 | **DB tests order-dependent.** `TestMain` truncated tables once per process, so `-count=3` accumulated rows and two exact-count tests failed. | `go test -count=3 ./internal/db/`: `expected 2 entries with mood >= 7 and got 4`. | `resetTables(t)` at the start of tests that assert exact counts. |
| 11 | **Circuit breaker serialized every call.** `Execute` held its mutex while running the operation: for webhooks, `retry.Do` (500ms + 1s backoff sleeps) plus the HTTP POST. All 3 workers share one breaker, so they **sent one webhook at a time**. | `TestClosed_OperationsRunConcurrently`: 3 callers could not be inside `Execute` together (failed after 1s). `TestHalfOpen_…`: a second caller **blocked behind the half-open trial** instead of being rejected. | Lock only for the before/after state checks; the operation runs unlocked. Half-open admits exactly one trial and rejects the rest. A generation counter means a call that started before a state change can't flip the breaker afterwards (the approach `sony/gobreaker` uses). |
| 12 | **Test packages corrupted each other's data.** `go test ./...` runs packages as parallel processes; `db` and `handlers` shared one database, and each truncates tables. | Seen once as an intermittent `TestUpdateEntry` failure. Then reproduced on demand: looping the db tests (200 iterations) while starting the handlers tests 25 times → **6 failed cases**. | `internal/testdb`: each package gets its own database (`<DB_NAME>_<pkg>`, created on first use). Same stress run: **0 failures**. |
| 13 | **Ignored errors** (golangci-lint v2.5.0: 11 `errcheck` + 5 `staticcheck`). The important one: a failed dead-letter insert was silently dropped, so the job vanished with no trace. | Lint output, 16 issues. | Dead-letter insert failures are logged with the full job context; cache errors are logged; style issues fixed. **0 issues.** |

Also changed:

- **Migrations:** embedded with `go:embed` + golang-migrate's `iofs` source, and the Dockerfile no longer copies SQL files.
- **CI:** gained a Redis service, `-race`, and a pinned linter (`golangci-lint-action@v8`, `v2.5.0`).
- **Dev overlay and scripts:** `docker-compose.dev.yml` and two scripts make the results reproducible.
- **Secrets:** an SSH command containing the VPS address was removed from a comment in `main.go`.

## Before / after: end-to-end, against the Docker image

Run with `docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build` and a small
Python probe (raw socket for the slow-header check):

| Check | Before | After |
|---|---|---|
| `GET /failed-jobs`, no credentials | `200`, payload includes another user's entry text | `401` (`200` with `X-Admin-Token`) |
| Client sends partial headers, then stalls | still open after 15s | closed by server at 5.0s |
| `docker stop` (SIGTERM) | exits, 0 graceful-shutdown log lines | `Shutdown signal received` → `Shutting down server` → `Server stopped gracefully` |
| Container with no migration files on disk | n/a (Dockerfile copied them) | migrates and serves requests (embedded) |

## Verification (2026-10-29, `golang:1.25` / go1.25.14 linux/amd64, Postgres 16, Redis 7)

```bash
bash scripts/test-integration.sh             # go vet + go test -race -count=1, all in Docker
bash scripts/test-integration.sh -count=3    # stability
```

| Check | Result |
|---|---|
| `go build ./cmd/... ./internal/...` | OK |
| `go vet ./cmd/... ./internal/...` | clean |
| `go test -race -count=1` | **78 passed, 0 failed, 0 `DATA RACE`** (54 before this pass) |
| `go test -race -count=3` | all packages `ok` |
| `golangci-lint run` (v2.5.0) | `0 issues` |
| `gofmt -l` on changed files | clean |

The CI workflow change itself is **not yet verified on GitHub**: nothing has been pushed. The
commands above reproduce each CI step locally in the same toolchain family.

## Decisions and trade-offs

**Timeouts: `http.TimeoutHandler` vs. fixing the hand-rolled middleware.**
A correct hand-rolled version needs exactly what `TimeoutHandler` already has: a buffered writer, a
mutex, a timed-out flag, and panic propagation. I chose the battle-tested stdlib version.
- *Cost: status code.* It returns `503`, not `504`. I think that's actually more accurate: `504` is
  what a gateway (nginx) returns when its upstream is slow; `503` says the app itself shed the request.
- *Cost: buffering.* Responses are buffered, and the wrapper doesn't implement `http.Flusher`/`Hijacker`.
  There are no streaming or WebSocket routes, so that's acceptable today.
- *Limit:* the handler goroutine still runs to completion after a timeout. Only context-aware DB and
  Redis calls stop early.

**Refresh tokens: `GETDEL` vs. Lua script vs. `WATCH`/`MULTI`.**
Consuming a single key is exactly what `GETDEL` does, in one round trip, with no script to version.
Lua would be needed for multi-key logic (e.g. token families for reuse detection). `WATCH` adds
retry loops for no benefit here. It requires Redis ≥ 6.2, so the CI and dev images use Redis 7.

**Worker shutdown: `RWMutex` + flag vs. never closing the channel / `context` cancellation.**
- *Never closing:* workers would need a separate stop signal, and buffered jobs could be abandoned.
- *Closing:* `range` drains the buffer for free.

The read lock lets many handlers enqueue concurrently. The write lock is taken once, at shutdown.
`AddJob` remains non-blocking (load shedding), so holding a read lock never waits on a full queue.

**Admin endpoint: static token vs. admin users/RBAC.**
One operator, one secret, constant-time compare, off by default. A team would need per-person
identities and an audit log; that is a different feature, so it is not covered here.

**Embedded migrations vs. copying SQL into the image.**
One artifact, identical behaviour in tests, local runs and the container, and no cwd assumptions.
The trade-off is that changing a migration requires a rebuild, which is what you want anyway.

**Pinned lint version vs. `latest`.**
"latest" made the build depend on someone else's release schedule. Pinning makes upgrades
deliberate (bump, run, fix).

## Known limitations

Measured or confirmed, not fixed in this pass, ordered by impact:

1. **Behind nginx, every client shares one rate-limit bucket.** nginx sets `X-Real-IP`, but the
   limiter keys on `r.RemoteAddr`, which is nginx's address. In the compose deployment the whole
   service would allow ~100 requests/minute in total.
   *Fix:* trust `X-Real-IP` only when the peer is the known proxy.
2. **Login timing reveals whether an email is registered.** Measured: wrong password for an existing
   email, median **145.6 ms**; unknown email, median **19.5 ms** (15 requests each). bcrypt only runs
   when the user exists.
   *Fix:* compare against a dummy hash for unknown emails, and rate-limit `/login` per account.
3. **Cache invalidation covers only page 1 / limit 10.** Other cached pages stay stale for up to their 60s TTL.
   *Fix:* a per-user version counter in the cache key (`INCR` on write), so all pages invalidate at once.
4. **The rate limiter's `INCR` and `EXPIRE` aren't atomic.** If the process dies between them, the key
   never expires and the client stays blocked. The limiter also *fails open* when Redis is down.
   *Fix:* a Lua script (or `SET NX EX` + `INCR`).
5. **No request-body size limit.** Large JSON bodies are decoded in full.
   *Fix:* `http.MaxBytesReader`.
6. **The webhook queue is in-memory.** Jobs accepted but not yet sent are lost if the process crashes.
   The dead-letter queue only covers jobs that were attempted.
   *Fix:* a transactional outbox table polled by the workers.
7. **Schema:** no foreign key from `entries.user_id`, no index on `(user_id, created_at)`. Pagination
   queries will degrade with data size.
8. **No connection-pool limits.** `SetMaxOpenConns`/`SetMaxIdleConns` appear in the design notes
   (`System design Questions/Day16-10000-Requests.md`) but not in `db.go`, so `database/sql` will open
   unlimited connections under load. *Fix:* set them from config and expose pool stats.
9. **`failed_jobs.payload` is returned as base64** (the model uses `[]byte`), which makes it awkward to read.
10. **Observability:** `/metrics` is custom JSON, not Prometheus format, and unauthenticated. There's no tracing.
11. **Global package state** (`db.DB`, `redis.Client`, worker globals) makes tests share state and
    blocks running two pools in one process. Coverage is low in `cmd/server` (1.6%) and `internal/redis` (12.5%).
    The handlers tests point a Redis client at `localhost:6379` and rely on the breaker when nothing
    listens there, which is why that package takes ~10s.
12. **The `Deploy` workflow (SSH to a VPS) has been failing** since 2026-09-08. It was not investigated
    in this pass. `token.txt`, a committed JWT, is still in the repository and its history. If the
    `JWT_SECRET` that signed it is still in use anywhere, rotate it.

## Next steps (in order)

1. Push, and confirm the CI run is green on GitHub (then disable or fix `Deploy` so it stops showing red).
2. Fix limitations 1 and 2. Both are small, security-relevant, and good interview stories.
3. Versioned cache keys (3), with a test that creates an entry and reads page 2 at `limit=20`.
4. Migration `000004`: foreign key + `(user_id, created_at)` index. Measure `EXPLAIN ANALYZE` before and after on ~100k rows.
5. Outbox table for webhooks (6), with a crash-in-the-middle test.
