#!/usr/bin/env bash
# Full test suite — unit + Postgres + Redis integration — with the race detector,
# run entirely in Docker. Same toolchain family as CI; no local Go, Postgres or
# Redis needed (and immune to OS policies that block freshly built test binaries).
#
#   bash scripts/test-integration.sh            # vet + go test -race -count=1
#   bash scripts/test-integration.sh -count=3   # extra args go to `go test`
set -euo pipefail
cd "$(dirname "$0")/.."
SRC="$(pwd -W 2>/dev/null || pwd)"   # Windows path under Git Bash, POSIX path elsewhere

NET=pab-it-net
cleanup() {
  docker rm -f pab-it-pg pab-it-redis >/dev/null 2>&1 || true
  docker network rm "$NET" >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup

docker network create "$NET" >/dev/null
docker run -d --name pab-it-pg --network "$NET" \
  -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=analytics_test \
  postgres:16-alpine >/dev/null
docker run -d --name pab-it-redis --network "$NET" redis:7-alpine >/dev/null

# -h 127.0.0.1: during first-boot init Postgres runs a temporary socket-only
# server; waiting on TCP avoids racing that restart.
until docker exec pab-it-pg pg_isready -h 127.0.0.1 -U user -d analytics_test >/dev/null 2>&1; do sleep 1; done

MSYS_NO_PATHCONV=1 docker run --rm --network "$NET" -v "$SRC:/src" -w /src \
  -e DB_HOST=pab-it-pg -e DB_PORT=5432 -e DB_USER=user -e DB_PASSWORD=password -e DB_NAME=analytics_test \
  -e REDIS_ADDR=pab-it-redis:6379 \
  golang:1.25 sh -c 'go vet ./cmd/... ./internal/... && go test -race -count=1 "$@" ./cmd/... ./internal/...' -- "$@"
