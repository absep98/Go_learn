#!/usr/bin/env bash
# Five-minute walkthrough of the API against a local stack.
#
#   docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build -d app
#   bash scripts/demo.sh
#
# Needs only bash + curl. ADMIN_TOKEN must match docker-compose.dev.yml.
set -euo pipefail
BASE=${BASE:-http://localhost:8080}
ADMIN_TOKEN=${ADMIN_TOKEN:-dev-admin-token}
EMAIL="demo-$(date +%s)@example.com"

step() { printf '\n\033[1m== %s\033[0m\n' "$*"; }
json_field() { sed -n "s/.*\"$1\":\"\([^\"]*\)\".*/\1/p"; }
status() { curl -s -o /dev/null -w '%{http_code}' "$@"; }

step "1. Health: Postgres + Redis reachable"
until curl -sf "$BASE/health" >/dev/null 2>&1; do sleep 1; done
curl -s "$BASE/health"

step "2. Register + login → 15-min access JWT and a single-use refresh token"
curl -s -X POST "$BASE/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"correct-horse\"}"
LOGIN=$(curl -s -X POST "$BASE/login" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"correct-horse\"}")
ACCESS=$(echo "$LOGIN" | json_field access_token)
REFRESH=$(echo "$LOGIN" | json_field refresh_token)
echo "access token: ${ACCESS:0:24}…   refresh token: $REFRESH"

step "3. Protected write without / with a token"
echo "no token   → HTTP $(status -X POST "$BASE/entries" -H 'Content-Type: application/json' -d '{}')"
curl -s -X POST "$BASE/entries" -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d '{"text":"Shipped the race-condition fixes","mood":8,"category":"work"}'

step "4. Paginated read (cache-aside: second call is served from Redis)"
curl -s "$BASE/entries?page=1&limit=10" -H "Authorization: Bearer $ACCESS"

step "5. Refresh-token rotation: first use succeeds, replay is rejected"
echo "first use  → HTTP $(status -X POST "$BASE/refresh" -H 'Content-Type: application/json' -d "{\"refresh_token\":\"$REFRESH\"}")"
echo "replay     → HTTP $(status -X POST "$BASE/refresh" -H 'Content-Type: application/json' -d "{\"refresh_token\":\"$REFRESH\"}")"

step "6. Webhook failure path: retries → circuit breaker → dead-letter queue"
echo "(the dev overlay points WEBHOOK_URL at a closed port; waiting for 3 attempts with backoff)"
sleep 4
echo "/failed-jobs without admin token → HTTP $(status "$BASE/failed-jobs")"
echo "/failed-jobs with admin token    → HTTP $(status -H "X-Admin-Token: $ADMIN_TOKEN" "$BASE/failed-jobs")"
curl -s -H "X-Admin-Token: $ADMIN_TOKEN" "$BASE/failed-jobs" | cut -c1-200; echo " …"

step "7. Request metrics"
curl -s "$BASE/metrics" | cut -c1-300; echo " …"

step "8. Graceful shutdown on SIGTERM (what docker stop sends)"
echo "run: docker compose -f docker-compose.yml -f docker-compose.dev.yml stop app"
echo "then: docker compose -f docker-compose.yml -f docker-compose.dev.yml logs app | grep -E 'Shutdown signal|stopped gracefully'"
