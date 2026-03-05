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

---

## Current Focus: Path B - Interview Prep

Why before more coding:
You have MORE than enough features. The gap is explaining them clearly.
Talking about code is a separate skill - needs practice.

Format: You explain each system like I am an interviewer.
I challenge with follow-up questions. You answer.

Topics to cover (in order):
- [x] Circuit breaker - what, why, how
- [x] Retry with exponential backoff - what, why, how
- [x] Rate limiting - what, why, Redis role
- [x] Worker pool - goroutines, channels, why buffered
- [x] JWT authentication - flow, what is inside token
- [x] Request timeout - context, goroutine lifecycle
- [x] Metrics system - thread safety, ResponseWriter wrapper
- [x] Caching strategy - Redis vs in-memory, TTL, eviction

---

## Path A - After Interview Prep

### Webhook/Notification System
Make worker pool do real work:
- Send HTTP POST webhook when entry is created
- Combine retry + worker pool + circuit breaker
- Learn: outbound HTTP calls, error handling on external requests

---

## Interview Answer Framework (use for every topic)

1. What - what is it in one sentence
2. Why - what problem does it solve (with concrete example)
3. How - how you implemented it (specific code decisions)
4. Trade-offs - what you would do differently in production
