---
title: I Deployed My Go Backend to a Real VPS. Here's Exactly What Happened.
subtitle: Part 15 of Learning Go in Public
tags: go, docker, devops, buildinpublic
series: Learning Go in Public
---

In [Part 14](link-to-part-14), I finished HMAC webhook signing. The backend was complete — JWT auth, PostgreSQL, Redis caching, rate limiting, circuit breaker, worker pool, webhook delivery, migrations, Docker. All running locally.

But "runs on my machine" isn't a portfolio project. It's a homework assignment.

Time to ship it.

## The Stack Being Deployed

- **Go backend** — ~15MB Docker image (multi-stage build, CGO_ENABLED=0)
- **PostgreSQL 16** — with golang-migrate running schema migrations on startup
- **Redis** — for caching, rate limiting, and refresh token storage
- **Oracle Cloud Free Tier** — 1GB RAM, 45GB disk, already provisioned

Everything wired together with `docker-compose.yml`. One command to start the entire stack.

## Step 1: The VM Was Fine, Actually

I was worried about the free tier specs. Turned out the "1GB" in the tier name refers to RAM, not disk. The actual disk is 45GB — plenty.

```bash
free -h    # 954MB RAM, 552MB available
df -h      # 45GB disk, 41GB free
```

The real constraint: **no swap**. Go's compiler is memory-hungry. Without swap, building the Docker image on the VM would exhaust RAM and kill the process. More on this in a moment.

## Step 2: Install Docker

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker ubuntu
newgrp docker
```

The official install script handles everything — Docker Engine, containerd, Docker Compose plugin. One command, done.

## Step 3: Clone the Repo

The repo is private. Created a fine-grained GitHub personal access token with **Contents: Read-only** permission. Used it to clone:

```bash
git clone https://TOKEN@github.com/absep98/Go_learn.git
```

Security note: never paste tokens in chat, email, or anywhere visible. Type them directly into the terminal. Tokens in chat history are compromised tokens.

## Step 4: Create the .env File

The `.env` from my local machine needed two changes for Docker:

```bash
# LOCAL (wrong for Docker):
DB_HOST=localhost
REDIS_HOST=localhost

# DOCKER (correct):
DB_HOST=postgres    # ← service name in docker-compose.yml
REDIS_HOST=redis    # ← service name in docker-compose.yml
```

This is the Docker networking model: each container has its own `localhost`. The service names (`postgres`, `redis`) are DNS hostnames that Docker Compose registers automatically on the internal network. Containers talk to each other by service name, not by `localhost`.

## Step 5: The OOM Problem

First build attempt:

```bash
docker compose up --build -d
```

The build ran for 30 minutes on the compile step, then the VM became unresponsive. SSH wouldn't connect. Classic OOM kill — the kernel killed processes when RAM was exhausted during `go build` inside Docker.

The fix: add swap space before building.

```bash
sudo dd if=/dev/zero of=/swapfile bs=1M count=1024
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab  # persist across reboots
```

With 1GB swap added, the build completed in ~5 minutes. The Go compiler used swap when it needed more than physical RAM allowed.

**Why does building inside Docker use more RAM than running the binary?**

The multi-stage Dockerfile has two stages:
- Stage 1 (builder): `golang:1.25-alpine` (~300MB) + Go compiler + all source code. The compiler is memory-hungry — it loads your entire dependency graph to type-check and optimize.
- Stage 2 (runner): `alpine:latest` (~5MB) + just the compiled binary (~10MB)

The builder stage is thrown away after compilation. But during the build, both stages are in memory simultaneously. On a 1GB machine with no swap, this is tight.

## Step 6: The Firewall

Containers started. Tried hitting the health endpoint from my local machine:

```
Invoke-RestMethod : Unable to connect to the remote server
```

The app was running — `docker compose ps` showed all three containers up. The issue was Oracle Cloud's network security — a firewall layer outside the VM that blocks all ports by default except SSH (22).

Navigation path:
1. Compute → Instances → your instance → Networking tab
2. Click the Subnet link
3. Security tab → Default Security List
4. Add Ingress Rule: TCP, 0.0.0.0/0, port 8080

After adding the rule, immediate success:

```powershell
Invoke-RestMethod -Uri http://140.245.202.129:8080/health

status  database  redis
------  --------  -----
healthy connected connected
```

## Step 7: Full End-to-End Test

```powershell
# Register
$body = @{email="test@example.com"; password="test123"} | ConvertTo-Json
Invoke-RestMethod -Uri http://140.245.202.129:8080/register -Method Post -Body $body -ContentType "application/json"
# → success: True, user_id: 1

# Login
$response = Invoke-RestMethod -Uri http://140.245.202.129:8080/login -Method Post -Body $body -ContentType "application/json"
$token = $response.access_token

# Create entry
$headers = @{Authorization = "Bearer $token"}
$body = @{text="First live entry!"; mood=9; category="milestone"} | ConvertTo-Json
Invoke-RestMethod -Uri http://140.245.202.129:8080/entries -Method Post -Headers $headers -Body $body -ContentType "application/json"
# → success: True, id: 1
```

Everything worked. The full stack — registration, JWT auth, PostgreSQL insert, Redis caching — running on a real public IP.

## What the Multi-Stage Build Actually Does

This is worth understanding. The Dockerfile has two `FROM` lines:

```dockerfile
# Stage 1: Heavy construction site (~300MB)
FROM golang:1.25-alpine AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2: Clean delivery box (~15MB)
FROM alpine:latest
COPY --from=builder /app/server /app/server
CMD ["/app/server"]
```

Stage 1 builds the binary. Stage 2 takes only the binary — not the Go compiler, not the source code, not the build tools. The final image is ~15MB instead of ~300MB.

`CGO_ENABLED=0` makes the binary fully static — it doesn't link against any C libraries from the OS. The binary runs on any Linux regardless of which C library it has. Copy it to a machine with no Go installed, it runs.

This is different from Java: a `.class` file needs the JVM to run. A Go binary is machine code — the CPU executes it directly with no runtime required.

## What I'd Do Differently in Production

**No HTTPS**: The API is HTTP only. Production needs TLS — either via nginx as a reverse proxy terminating SSL, or Let's Encrypt directly. Anyone sniffing traffic can see the JWT tokens in plain text.

**No domain name**: `140.245.202.129:8080` works but isn't professional. A domain + HTTPS is the next step.

**Secrets in .env on disk**: The `.env` file sitting in the repo directory isn't ideal. Production would use Docker secrets, environment injection from a secrets manager, or at minimum ensure the file has restricted permissions.

**Manual deployment**: Every code change requires SSH + `git pull` + `docker compose up --build`. Production would have a CI/CD pipeline (GitHub Actions) that automatically builds and deploys on push.

**No health monitoring**: If the server crashes at 3am, nobody knows. Production needs uptime monitoring (UptimeRobot is free) and alerting.

These are real production concerns — but for a portfolio project demonstrating backend fundamentals, a live public IP with a working API is the right stopping point.

## What's Live

`http://140.245.202.129:8080`

Endpoints:
- `GET /health` — database + Redis status
- `POST /register` — create account
- `POST /login` — get JWT + refresh token
- `POST /refresh` — rotate tokens
- `POST /logout` — invalidate session
- `GET/POST/PATCH/DELETE /entries` — journal entries (protected)
- `GET /metrics` — request counts, latency per endpoint

Full source: [github.com/absep98/Go_learn](https://github.com/absep98/Go_learn)

---

*This is Part 15 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i) | [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g) | [Part 8](https://dev.to/abhishek_sharma_a9792aee8/i-built-my-backend-on-sqlite-then-i-deleted-it-15a) | [Part 9](https://dev.to/abhishek_sharma_a9792aee8/my-database-disappeared-every-time-i-restarted-docker-migrations-fixed-that-n93) | [Part 10](https://dev.to/abhishek_sharma_a9792aee8/every-request-looked-the-same-in-my-logs-then-i-added-one-line-19m7) | [Part 11](https://dev.to/abhishek_sharma_a9792aee8/i-had-no-idea-which-endpoint-was-slowest-so-i-built-my-own-metrics-5fjp) | [Part 12](https://dev.to/abhishek_sharma_a9792aee8/my-backend-crashed-every-time-redis-went-down-three-patterns-fixed-that-43j2) | [Part 13](https://dev.to/abhishek_sharma_a9792aee8/i-needed-to-send-an-http-request-without-slowing-down-my-api-goroutines-fixed-that) | [Part 14](link-to-part-14)*
