# Docker & Docker Compose — Deep Learning Notes

## Core Concept: What is Docker?

Docker packages your app and everything it needs (code, runtime, dependencies) into a
**container** — an isolated, self-contained unit that runs the same everywhere.

Without Docker: "It works on my machine" problem.
With Docker: The container IS your machine config. Ships with the app.

---

## Dockerfile — Two Stage Build

Our Dockerfile has two `FROM` statements — this is called a **multi-stage build**.

### Why two stages?

You need a large environment (Go compiler, build tools) to BUILD the app.
You need almost nothing to RUN the compiled app.
Separating them keeps the final image tiny and secure.

```
Stage 1: Builder (~300MB)    Stage 2: Runner (~15MB)
──────────────────────────   ──────────────────────
golang:1.25-alpine           alpine:latest
Go compiler                  (nothing installed)
Source .go files             ca-certificates
go mod cache                 server binary ← copied from Stage 1
         │
         └─── go build → server binary ──→ COPY --from=builder
```

### Stage 1 — The Builder

```dockerfile
FROM golang:1.25-alpine AS builder
```
Base image = Alpine Linux + Go 1.25 compiler pre-installed. Named `builder` so Stage 2 can reference it.

```dockerfile
WORKDIR /app
```
Creates `/app` directory inside the container. All subsequent commands run from here.

```dockerfile
COPY go.mod go.sum ./
RUN go mod download
```
**Why copy only these two files first?**
Docker caches each step. If `go.mod` didn't change, it skips `go mod download` entirely on the
next rebuild (saves ~40 seconds every time you change your code).

```
go.mod unchanged → steps CACHED → go mod download SKIPPED
go.mod changed   → steps rerun  → packages re-downloaded
```

```dockerfile
COPY . .
```
Now copy all `.go` source files. This happens AFTER dependency download so cache isn't busted
by every code change.

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
```
Compiles everything into a single binary called `server`.

- `CGO_ENABLED=0` — disables C library dependencies. Without this, the binary looks for `glibc`
  at runtime. Alpine uses `musl` instead → crash. With this flag, the binary is **static** —
  completely self-contained, needs nothing from the OS.
- `GOOS=linux` — cross-compile for Linux even if building on Windows/Mac.
- `-o server` — name the output file `server`.

### Stage 2 — The Runner

```dockerfile
FROM alpine:latest
```
Brand new container, completely separate from Stage 1. Only ~5MB. Stage 1 is thrown away.

`alpine:latest` = "latest version of Alpine Linux" — the tag after `:` is the version.
Same pattern: `postgres:16-alpine` = Postgres version 16 on Alpine.

```dockerfile
RUN apk add --no-cache ca-certificates
```
Alpine is so minimal it doesn't have SSL/TLS certificates. Without these, any HTTPS connection
(to Postgres, Redis, external APIs) would fail.

```dockerfile
COPY --from=builder /app/server /app/server
```
The key line. Reaches back into `builder` and grabs ONLY the compiled binary.
No source code, no Go compiler, no build tools included.

```dockerfile
CMD ["/app/server"]
```
The command that runs when the container starts. Because binary is static, it carries
everything it needs — no Go runtime required.

---

## docker-compose.yml — The Full Stack

Docker Compose defines all services and starts them with one command.

```
docker compose up --build   → build images + start all containers
docker compose down         → stop and remove all containers
docker compose logs         → view logs without starting
docker compose ps           → see what's running
```

### Internal Network

Docker Compose automatically creates a private network for all services.
Services talk to each other using their **service name as hostname** — not `localhost`.

```
Your machine       → localhost:8080  → app container
app container      → postgres:5432   → postgres container (service name as hostname)
app container      → redis:6379      → redis container
Outside Docker     → localhost:5432  → postgres (only if ports: - "5432:5432" is set)
```

### `ports` — only for external access

```yaml
ports:
  - "8080:8080"   # HOST:CONTAINER
```

- **Left side** (HOST) = port on YOUR machine
- **Right side** (CONTAINER) = port inside Docker

`ports` only exposes a service to your machine. Container-to-container communication
works WITHOUT `ports` — they're already on the same internal network.

```
Remove ports from postgres → your pgAdmin/psql tools can't connect ❌
                           → your Go app still connects fine ✅ (internal network)
```

The only port that MUST exist is `8080:8080` on the app — so curl/browser can reach the API.

### `depends_on` — startup ordering

```yaml
depends_on:
  redis:
    condition: service_started    # wait for container to start
  postgres:
    condition: service_healthy    # wait for healthcheck to pass
```

`service_started` = container process started (doesn't mean app inside is ready).
`service_healthy` = the healthcheck command returned success (app inside IS ready).

### `healthcheck` — readiness probe

```yaml
# On the postgres service:
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U user -d dbname"]
  interval: 5s     # check every 5 seconds
  timeout: 5s      # fail if no response in 5 seconds
  retries: 5       # mark unhealthy after 5 consecutive failures
```

`pg_isready` is a Postgres tool that returns success ONLY when Postgres is accepting connections.
Docker runs this on a timer. Only when it passes does `service_healthy` become true.

**Why we still needed `restart: on-failure`:**
Postgres has a quirk — during initialization it starts once, then briefly restarts to become
the real production server. The healthcheck can pass during the init phase, app starts,
Postgres shuts down briefly → "connection refused" → app crashes.

`restart: on-failure` = if the container exits with a non-zero code, Docker restarts it automatically.
Combined with healthcheck, Postgres is fully up on the retry.

### Environment Variables

```yaml
environment:
  - DB_HOST=postgres      # service name = hostname inside Docker
  - DB_PORT=5432
  - DB_USER=user
  - DB_PASSWORD=password
  - DB_NAME=dbname
```

These become `os.Getenv()` values inside the container. Same as your `.env` file,
but injected by Docker instead of loaded by `godotenv`.

**⚠️ Production security:**
Hardcoding secrets in `docker-compose.yml` is dangerous — if the file is committed to Git,
anyone with repo access can see passwords and JWT secrets and can:
- Connect directly to your database
- Forge JWT tokens and log in as any user

**The fix for local dev:**
```yaml
# docker-compose.yml
environment:
  - JWT_SECRET=${JWT_SECRET}    # reads from .env
  - DB_PASSWORD=${DB_PASSWORD}

# .env (added to .gitignore — never committed)
JWT_SECRET=supersecretkey
DB_PASSWORD=password
```

**Production:** Use Docker Secrets, AWS Secrets Manager, or HashiCorp Vault.

### `restart: on-failure`

```yaml
app:
  restart: on-failure
```

If the app container crashes (non-zero exit), Docker restarts it automatically.
Only restarts the `app` container — `postgres` and `redis` in their own containers are
completely unaffected. Each service is isolated.

---

## Image naming: `imagename:tag`

| Image | Meaning |
|---|---|
| `golang:1.25-alpine` | Go 1.25 on Alpine Linux |
| `postgres:16-alpine` | Postgres 16 on Alpine Linux |
| `redis:alpine` | Latest Redis on Alpine Linux |
| `alpine:latest` | Latest Alpine Linux (bare) |

Images without `build:` are pulled from Docker Hub automatically on first run.
Images with `build: .` are built from the local Dockerfile.

---

## Full startup sequence

```
docker compose up --build

1. Build Dockerfile → compile Go → create app image
2. Pull redis:alpine (if not cached)
3. Pull postgres:16-alpine (if not cached)
4. Start redis container        (service_started)
5. Start postgres container     (waiting for healthcheck)
6. pg_isready runs every 5s → passes → postgres marked Healthy
7. Start app container          (depends_on condition met)
8. App crashes (postgres brief restart window)
9. Docker restarts app          (restart: on-failure)
10. App connects to postgres:5432 ✅
11. App connects to redis:6379 ✅
12. Server listening on :8080 ✅
```

---

## Data Persistence — Volumes

By default, Postgres stores data **inside the container filesystem**.
When the container is removed, all data is gone.

```
docker compose down  → container removed → ALL postgres data deleted ❌
docker compose up    → fresh empty database
```

### Fix: Docker Volumes

A volume maps a folder inside the container to a folder on the **host machine's real disk**.
Data lives on your actual disk, not inside the container.

```yaml
postgres:
  image: postgres:16-alpine
  volumes:
    - postgres_data:/var/lib/postgresql/data  # container path → named volume

volumes:
  postgres_data:   # Docker manages this folder on host disk
```

Now:
```
docker compose down  → container removed, but data on host disk survives ✅
docker compose up    → new container mounts same volume → data still there ✅
```

---

## Test Database vs Production Database

Tests always run against a **separate throwaway database** — never production.

```
Local dev     → Docker Postgres on localhost:5432  (fake data, safe to wipe)
CI/CD         → GitHub Actions spins up fresh Postgres per run
Staging       → copy of production, used for QA
Production    → real users, real data — tests NEVER run here
```

### Why TRUNCATE before tests?

Tests need predictable state. If `TestValidRegistration` creates `test@test.com` in Postgres,
running tests again finds that email already exists → 409 → test fails.

`TRUNCATE entries, users RESTART IDENTITY CASCADE` gives a clean slate every run.
- `TRUNCATE` = delete all rows (fast)
- `RESTART IDENTITY` = reset auto-increment IDs back to 1
- `CASCADE` = also truncate tables referencing users (entries has user_id foreign key)

This is the standard pattern for integration tests against a real database.
Analogy: a hospital practices surgery on mannequins that reset between sessions, not real patients.

---

## Production Database — What Real Companies Use

Nobody runs a production database in Docker on a single server.

| Layer | What it does |
|---|---|
| **Volumes** | Survive container restarts (local dev) |
| **Managed DB service** | AWS RDS, Google Cloud SQL — Postgres on dedicated hardware |
| **Automated backups** | Daily snapshots, point-in-time recovery |
| **Replication** | Primary + replica(s) — if primary dies, replica takes over |
| **Multi-AZ** | Data replicated across availability zones (different data centers) |

### If primary DB server fails:
- Without replication → data loss, downtime
- With Multi-AZ replication → automatic failover in seconds, no data loss

For our project: volumes are enough for local dev. For production: use a managed database
service (AWS RDS etc.), never raw Docker for the DB.
