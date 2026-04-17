# Next Path - What's Done & What's Next

## Completed Features

### Tier 1: Core Backend
- JWT Authentication (register/login)
- CRUD entries (POST, GET, PATCH, DELETE)
- SQLite database with parameterized queries
- Input validation + error handling
- Pagination

### Tier 2: Scaling and Performance
- Redis caching (GET entries)
- Rate limiting (fixed window via Redis)
- Worker pool (goroutines + channels)
- Graceful shutdown (signal handling)

### Tier 3: Observability
- Structured logging (slog, JSON output)
- Request ID tracing
- Metrics endpoint (request count, latency, errors per endpoint)

### Tier 4: Production Readiness
- Centralized config (12-Factor App, environment variables)
- Fail-fast validation (JWT_SECRET required)
- Cache cleanup goroutine (ticker-based eviction, memory leak prevention)
- Request timeout middleware (context.WithTimeout, 504 on slow handlers)
- Retry with exponential backoff (generic higher-order function)
- Circuit breaker (3-state machine: closed -> open -> half-open)
- Per-user rate limiting (UserRateLimitMiddleware, Redis key per user ID)

### Tier 5: Containerization
- Multi-stage Docker build (~15MB final image, CGO_ENABLED=0)
- docker-compose with 3 services: app, postgres:16-alpine, redis:alpine
- Healthchecks + depends_on (postgres healthy before app starts)

---

## Roadmap (in order)

### Next: Webhook / Notification System ← CURRENT
Make worker pool do real work:
- Send HTTP POST webhook when entry is created
- Combine retry + worker pool + circuit breaker on real outbound HTTP
- Learn: outbound HTTP calls, fire-and-forget, reliability on external requests

### After That: Token Refresh & Revocation
Real security gap in current JWT implementation:
- Short-lived access tokens (15min) + refresh tokens stored in Redis
- Revocation by deleting the refresh token from Redis
- Teaches: auth flows, Redis as token store, security trade-offs

### Then: Database Migrations
Replace CREATE TABLE IF NOT EXISTS with proper migration tool:
- golang-migrate library
- Version-controlled schema changes (up/down migrations)
- Teaches: production schema management, rollback strategies

### Skip for now: Prometheus + Kubernetes
- Prometheus: good in real job, but 80% YAML config not Go learning
- Kubernetes: too much overhead for portfolio project, Docker already proves containerization

---

## Interview Answer Framework (use for every topic)

1. What - what is it in one sentence
2. Why - what problem does it solve (with concrete example)
3. How - how you implemented it (specific code decisions)
4. Trade-offs - what you would do differently in production
