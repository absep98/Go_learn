# Personal Analytics Backend - Architecture Map

**Purpose:** Visual guide showing how everything connects. Read this when you feel lost.

---

## 🏗️ THE BIG PICTURE

```
USER SENDS REQUEST
        ↓
   [main.go] - Routes requests to handlers
        ↓
   [middleware chain] - RequestID → Timeout → Metrics → RateLimit → Logging → [Auth] → Handler
        ↓
   [handlers/*.go] - Processes request, validates data
        ↓
        ├── [db/db.go]        - Saves/retrieves from PostgreSQL
        ├── [cache/cache.go]  - Reads/writes Redis cache (entry listings)
        └── [worker/worker.go]- Fires background jobs (webhooks)
        ↓
   RESPONSE SENT BACK
```

---

## 📁 FILE-BY-FILE BREAKDOWN

### **cmd/server/main.go** (Entry Point)

**What it does:** Loads config, initialises Postgres + Redis, wires Handler, starts worker pool, registers routes, handles graceful shutdown.

```
Routes:
├── /health      → No auth needed
├── /ping        → No auth needed
├── /register    → No auth needed
├── /login       → No auth needed
├── /refresh     → No auth needed (refresh token proves identity via Redis)
├── /logout      → No auth needed
├── /metrics     → No auth needed
└── /entries     → 🔒 PROTECTED
                   POST   → CreateEntry
                   GET    → GetEntries
                   PATCH  → UpdateEntry
                   DELETE → DeleteEntry
```

Every route is wrapped with the full middleware chain:
```
RequestIDMiddleware → TimeoutMiddleware → MetricsMiddleware → RateLimitMiddleware → LoggingMiddleware → [AuthMiddleware → UserRateLimitMiddleware] → Handler
```
AuthMiddleware and UserRateLimitMiddleware only appear on /entries.

---

### **internal/handlers/middleware.go** (Middleware layer)

Contains multiple middlewares — each wraps the next like Russian dolls:

| Middleware | What it does | Degrades if down? |
|---|---|---|
| RequestIDMiddleware | Generates UUID, adds to context + response header | — |
| TimeoutMiddleware | Cancels request after 30s, returns 503 | — |
| MetricsMiddleware | Counts requests/errors per endpoint | — |
| RateLimitMiddleware | IP-based limit via Redis INCR | ✅ fail open (allow) |
| LoggingMiddleware | Logs method, path, duration, status | — |
| AuthMiddleware | Validates JWT, puts user_id in context | — |
| UserRateLimitMiddleware | Per-user limit via Redis INCR | ✅ fail open (allow) |

**AuthMiddleware flow:**
```
1. Get "Authorization" header
2. Strip "Bearer " prefix → token string
3. validateToken() → verify signature + expiry with JWTSecret
4. Extract user_id from claims
5. context.WithValue(r.Context(), userIDKey, userID)
6. Call next handler
If anything fails → 401 Unauthorized
```

---

### **internal/handlers/auth.go** (Register / Login / Refresh / Logout)

**What it does:** Manages user identity and session tokens.

#### Register Flow
```
POST /register {"email", "password"}
    ↓
1. Validate email (not empty, contains @)
2. Validate password (not empty, >= 6 chars)
3. bcrypt.GenerateFromPassword(password, cost=10)
4. h.Store.CreateUser(email, hash) → PostgreSQL INSERT
5. 201 {"success": true, "user_id": 42}
```

#### Login Flow
```
POST /login {"email", "password"}
    ↓
1. h.Store.GetUserByEmail(email) → fetch userID + passwordHash from DB
2. bcrypt.CompareHashAndPassword(hash, password)
3. h.generateAccessToken(userID) → JWT signed with h.JWTSecret, exp=15min
4. uuid.New() → random UUID for refresh token
5. h.Cache.Set("refresh_token:<uuid>", "<userID>", 7 days)
6. 200 {"access_token": "eyJ...", "refresh_token": "<uuid>"}
```

#### Refresh Flow
```
POST /refresh {"refresh_token": "<uuid>"}
    ↓
1. key = "refresh_token:" + uuid
2. val = h.Cache.Get(key)  ← val IS the userID string ("42")
3. userID = strconv.ParseInt(val)
4. h.Cache.Del(key)                     ← old token consumed (rotation)
5. h.generateAccessToken(userID)        ← new access token
6. h.Cache.Set("refresh_token:<newUUID>", userID, 7 days)  ← new refresh token
7. 200 {"access_token": "eyJ...", "refresh_token": "<newUUID>"}
```
Note: /refresh is NOT protected by AuthMiddleware. The refresh token itself proves identity via Redis lookup.

#### Logout Flow
```
POST /logout {"refresh_token": "<uuid>"}
    ↓
1. h.Cache.Get("refresh_token:" + uuid)   ← verify it exists
2. h.Cache.Del("refresh_token:" + uuid)   ← invalidate session
3. 200 {"success": true}
```

---

### **internal/handlers/entries.go** (Entry CRUD)

**What it does:** Creates, reads, updates, deletes journal entries. Cache-aside pattern on GET.

#### Create Entry Flow
```
POST /entries {"text", "mood", "category"}
    ↓
1. user_id from context (set by AuthMiddleware)
2. Validate text, mood (1-10), category
3. h.Store.InsertEntry(userID, text, mood, category)
4. h.Cache.Del("entries:user:<userID>:...")  ← invalidate cached list
5. worker.AddJob(webhookPayload)             ← fire-and-forget notification
6. 201 {"success": true, "id": 456}
```

#### Get Entries Flow (cache-aside)
```
GET /entries?page=1&limit=10
    ↓
1. user_id from context
2. cacheKey = "entries:user:<id>:page:<p>:limit:<l>"
3. h.Cache.Get(cacheKey)
   HIT  → write raw JSON to response, return immediately
   MISS → query PostgreSQL → marshal JSON → h.Cache.Set(key, data, 60s) → respond
```

---

### **internal/db/db.go** (Database Layer)

**What it does:** All PostgreSQL operations. Uses `modernc.org/sqlite` was replaced with `lib/pq` (PostgreSQL driver).

#### Tables
```
users table:
├── id            BIGSERIAL PRIMARY KEY
├── email         TEXT UNIQUE NOT NULL
├── password_hash TEXT NOT NULL
└── created_at    TIMESTAMPTZ DEFAULT NOW()

entries table:
├── id         BIGSERIAL PRIMARY KEY
├── user_id    BIGINT REFERENCES users(id)
├── text       TEXT NOT NULL
├── mood       INTEGER (1-10)
├── category   TEXT
└── created_at TIMESTAMPTZ DEFAULT NOW()
```

Key: always uses `$1, $2` placeholders (PostgreSQL syntax, not `?`).

---

## 🔴 REDIS — THREE USES

This is the most commonly forgotten part. Redis is used for three completely separate purposes:

### 1. Refresh Token Store (`auth.go`)
```
Key:   "refresh_token:<uuid>"
Value: userID as string ("42")
TTL:   7 days

Set at: Login, Refresh (new token issued)
Get at: Refresh, Logout (verify token exists)
Del at: Refresh (token rotation), Logout (invalidation)
```
Behaviour if Redis down: Login returns 500 — hard failure, users cannot log in.

### 2. Entry List Cache (`entries.go` + `cache/cache.go`)
```
Key:   "entries:user:<id>:page:<p>:limit:<l>"
Value: JSON string of the entries response
TTL:   60 seconds

Set at: GET /entries on cache miss (after DB query)
Get at: GET /entries (checked before DB)
Del at: POST/PATCH/DELETE /entries (cache invalidation on write)
```
Behaviour if Redis down: circuit breaker opens → Get returns miss → falls through to PostgreSQL. Degrades gracefully (slower, not broken).

### 3. Rate Limiting (`ratelimit.go`)
```
Key:   "ratelimit:<ip>"          (IP-based, RateLimitMiddleware)
       "ratelimit:user:<userID>" (per-user, UserRateLimitMiddleware)
Value: request count integer
TTL:   1 minute (set only on first request, count==1)

Commands: INCR key → returns new count. EXPIRE key 60 (first request only).
```
Behaviour if Redis down: `redis.Client.Incr()` returns error → `IsAllowed()` returns `true` → requests pass through unthrottled. Fails open intentionally — blocking all traffic because Redis is down would be worse than allowing unthrottled requests temporarily.

### Redis Degradation Summary
| Feature | Redis down behaviour |
|---|---|
| Rate limiting | Fail open — requests allowed |
| Entry caching | Fail open — falls back to Postgres |
| Refresh tokens | Hard fail — login returns 500 |

---

## ⚡ CIRCUIT BREAKER (`internal/circuitbreaker/circuitbreaker.go`)

Wraps Redis cache operations (NOT rate limiting — that uses redis.Client directly).

```
3 states:

CLOSED (normal):
  Request → call Redis → success: reset failureCount
                       → failure: failureCount++
                         if failureCount >= 5 → switch to OPEN

OPEN (Redis assumed down):
  Request → reject immediately, return error (no Redis call)
  After 30s cooldown → switch to HALF-OPEN

HALF-OPEN (testing recovery):
  ONE request allowed through
  success → switch to CLOSED (recovered)
  failure → switch back to OPEN (restart cooldown)
```

Key: circuit breaker does NOT retry. That's `retry.go`'s job. They do opposite things:
- `retry.go` — keep trying (used at startup: InitRedis)
- `circuitbreaker.go` — stop trying (used at runtime: cache operations)

---

## 🔧 WORKER POOL + WEBHOOK (`internal/worker/`, `internal/webhook/`)

```
POST /entries
    ↓
handler calls worker.AddJob(payload)   ← non-blocking, returns immediately
    ↓ (background goroutine picks it up)
worker.go → WebhookBreaker.Execute()   ← circuit breaker around webhook endpoint
    ↓
retry.Do(3 attempts, 500ms backoff)    ← retry transient failures
    ↓
webhook.HTTPSender.Send(url, payload)  ← single HTTP POST attempt
```

Why non-blocking: The HTTP handler must respond fast. Webhook delivery (which may take seconds + retries) happens in a background goroutine pool. User gets their 201 immediately.

Pipeline layers:
| Layer | Purpose |
|---|---|
| `AddJob()` | Enqueue job, never blocks the HTTP handler |
| `WebhookBreaker` | Stop hammering a down webhook endpoint |
| `retry.Do()` | Retry transient failures (network blips) |
| `HTTPSender.Send()` | One HTTP POST attempt, 5s timeout |

---

---

## 🔐 AUTHENTICATION FLOW (The Complete Journey)

### Registration
```
POST /register {"email", "password"}
    ↓ validate → bcrypt hash → INSERT INTO users → 201 {user_id}
```

### Login
```
POST /login {"email", "password"}
    ↓ SELECT user → bcrypt compare → generate JWT (15min) → store refresh token in Redis (7d)
    ↓ 200 {"access_token": "eyJ...", "refresh_token": "<uuid>"}
```

### Protected Request
```
GET /entries
Header: Authorization: Bearer eyJhbGci...

AuthMiddleware:
1. Strip "Bearer " → token string
2. jwt.ParseWithClaims(token, h.JWTSecret) → verify signature + expiry
3. Extract user_id from claims
4. context.WithValue → user_id travels to handler
5. Handler reads: r.Context().Value(userIDKey).(int64)
```

### Token Expired → Refresh
```
POST /refresh {"refresh_token": "<uuid>"}
    ↓ Redis GET "refresh_token:<uuid>" → returns userID string
    ↓ Redis DEL old token → generate new access token + new refresh UUID
    ↓ Redis SET new token → 200 {new access_token, new refresh_token}
```

### Logout
```
POST /logout {"refresh_token": "<uuid>"}
    ↓ Redis DEL "refresh_token:<uuid>" → session invalidated
```

---

## 🔑 KEY CONCEPTS

### 1. Why Hash Passwords?
```
Plain in DB:  "mysecret123"     BAD — if DB leaks, all passwords exposed
Hashed in DB: "$2a$10$xyz..."   GOOD — bcrypt is one-way + salted
                                Same password hashed twice = different result
                                bcrypt.CompareHashAndPassword re-applies same salt
```

### 2. Why JWT + Refresh Tokens?
```
Access token (JWT, 15min):
- Self-contained — server reads user_id without hitting DB
- Short-lived — if stolen, expires fast
- Cannot be revoked before expiry (stateless trade-off)

Refresh token (UUID in Redis, 7d):
- Stored server-side — can be revoked by deleting from Redis
- Not sent on every request — only used to get a new access token
- Rotation on every refresh — old token deleted, new token issued
```

### 3. Why Middleware Chaining?
```
RequestID → Timeout → Metrics → RateLimit → Logging → Auth → UserRateLimit → Handler

Each middleware calls next(w, r) to pass control inward.
Return before next() = short-circuit (what RateLimit and Auth do on reject).
Code AFTER next(w, r) runs on the way OUT (what Logging uses for duration).

Order matters:
- RequestID must be outermost → all downstream logs include request_id
- Auth before UserRateLimit → UserRateLimit reads user_id set by Auth
- RateLimit before Logging → no point logging a rate-limited request's duration
```

### 4. Why Context?
```
Problem: AuthMiddleware verifies JWT, but how does CreateEntry know the user_id?

Solution: context.WithValue
- Middleware: context.WithValue(r.Context(), userIDKey, int64(42))
- Handler:    r.Context().Value(userIDKey).(int64)
- Context is immutable and request-scoped — safe across goroutines
```

### 5. Why Dependency Injection (Handler struct)?
```
type Handler struct {
    Store     Store   // interface — real PostgreSQL or FakeStore in tests
    Cache     Cache   // interface — real Redis or FakeCache in tests
    JWTSecret string  // no os.Getenv inside handlers
}

Without DI: handlers call os.Getenv("JWT_SECRET") directly → untestable
With DI:    tests pass h := &Handler{JWTSecret: "test-secret"} → no env vars needed
```

---

## 📊 DATA FLOW EXAMPLE — Creating an Entry (Full Journey)

```
1. CLIENT:
   POST /entries
   Authorization: Bearer eyJhbGci...xyz
   {"text": "Great day!", "mood": 9, "category": "work"}

2. RequestIDMiddleware:   generate "abc-123", add to context + X-Request-ID header

3. TimeoutMiddleware:     start 30s countdown

4. MetricsMiddleware:     increment in-flight counter

5. RateLimitMiddleware:   redis.Client.INCR("ratelimit:192.168.1.1")
                          count=1 → redis.EXPIRE 60s → allow

6. LoggingMiddleware:     log "POST /entries started", start timer

7. AuthMiddleware:        parse JWT → user_id=42 → context.WithValue

8. UserRateLimitMiddleware: redis.Client.INCR("ratelimit:user:42") → allow

9. CreateEntry handler:
   ├── user_id = r.Context().Value(userIDKey).(int64)  → 42
   ├── validate text, mood, category
   ├── h.Store.InsertEntry(42, "Great day!", 9, "work")
   │   └── INSERT INTO entries ... RETURNING id  → 456
   ├── h.Cache.Del("entries:user:42:...")  ← invalidate cached list
   └── worker.AddJob({userID:42, entryID:456, ...})  ← background, non-blocking

10. LoggingMiddleware (on way out): log "POST /entries 201 in 12ms"

11. CLIENT RECEIVES:
    201 Created
    X-Request-ID: abc-123
    {"success": true, "id": 456}

12. BACKGROUND (worker goroutine):
    WebhookBreaker.Execute() → retry.Do(3x) → HTTPSender.Send(webhookURL, payload)
```

---

## 🎯 COMMON PATTERNS

### Pattern 1: Handler Structure
```go
func (h *Handler) SomeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req SomeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondJSON(w, http.StatusBadRequest, ...)
        return
    }
    // validate fields with early returns
    result, err := h.Store.DoSomething(r.Context(), req.Field)
    if err != nil {
        slog.Error("...", "error", err)
        respondJSON(w, http.StatusInternalServerError, ...)
        return
    }
    respondJSON(w, http.StatusOK, result)
}
```

### Pattern 2: Error Handling
```go
result, err := someFunction()
if err != nil {
    slog.Error("operation failed", "error", err)
    return fmt.Errorf("context: %w", err)
}
```

### Pattern 3: Cache-Aside Read
```go
cached, err := h.Cache.Get(r.Context(), cacheKey)
if err == nil {          // cache HIT
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprint(w, cached)
    return
}
// cache MISS — query DB
data, err := h.Store.GetSomething(r.Context(), ...)
// store for next time
h.Cache.Set(r.Context(), cacheKey, data, 60*time.Second)
```

### Pattern 4: Context Values
```go
// Set (in middleware):
ctx := context.WithValue(r.Context(), userIDKey, int64(42))
r = r.WithContext(ctx)

// Get (in handler):
userID := r.Context().Value(userIDKey).(int64)
```

---

## 💡 WHEN YOU FORGET SOMETHING

| Question | Where to look |
|---|---|
| How is JWT verified? | `middleware.go` → `validateToken()` |
| Where does user_id come from in handlers? | `middleware.go` → `AuthMiddleware` sets it in context |
| Where does userID come from in /refresh? | Redis value — stored at Login time as `fmt.Sprintf("%d", userID)` |
| How is password hashed? | `auth.go` → `Register` → `bcrypt.GenerateFromPassword` |
| How is refresh token stored? | `auth.go` → `Login` → `h.Cache.Set("refresh_token:"+uuid, userID, 7d)` |
| What are the 3 Redis uses? | Refresh tokens, entry caching, rate limiting — see Redis section above |
| What does circuit breaker protect? | Redis cache operations in `cache/cache.go` |
| How does webhook fire? | `entries.go` → `worker.AddJob()` → background goroutine → retry → HTTP POST |
| How to protect a route? | `main.go` → wrap with `h.AuthMiddleware(handler)` |
| How to write a unit test? | `auth_test.go` — uses FakeStore/FakeCache, no DB needed |
| How to write an integration test? | `entries_test.go` → `TestMain` checks `DB_HOST` env var |

---

## 🚀 RUNNING THE PROJECT

```powershell
# Start server (requires .env with DB/Redis config)
cd Project-Building-While-Learning\personal-analytics-backend
go run cmd/server/main.go

# Run unit tests only (no Docker needed)
go test ./internal/handlers/ -v -run TestRegister

# Run all tests including integration (Docker required)
docker run --rm -v "C:\learning\Go_learn\Project-Building-While-Learning\personal-analytics-backend:/app" -w /app golang:1.25-alpine go test ./internal/handlers/ -v

# Build check
go build ./cmd/... ./internal/...
```

---

## ✅ WHAT YOU'VE BUILT

- ✅ User registration with bcrypt password hashing
- ✅ JWT authentication (15min access tokens, 7d refresh tokens)
- ✅ Refresh token rotation + logout (Redis-backed sessions)
- ✅ Protected routes with middleware chain
- ✅ User-specific data isolation
- ✅ PostgreSQL with golang-migrate migrations
- ✅ Redis caching (cache-aside, 60s TTL, invalidation on write)
- ✅ IP + per-user rate limiting via Redis
- ✅ Circuit breaker (3-state: closed → open → half-open)
- ✅ Retry with exponential backoff
- ✅ Background worker pool + webhook delivery
- ✅ Graceful shutdown (SIGINT/SIGTERM handling)
- ✅ Structured logging (slog, JSON output, request ID tracing)
- ✅ Request timeout middleware
- ✅ Metrics endpoint
- ✅ Dependency injection (Handler struct, testable without DB)
- ✅ Unit tests (FakeStore/FakeCache) + integration tests (real Postgres)
- ✅ Multi-stage Docker build (~15MB image)
4. Week 2 Day 13: Documentation and review

**For now: Just USE what you built. Understanding comes from using, not reading.**

---

**Remember:** You don't need to memorize syntax. You need to understand:

1. What each file does
2. How data flows between them
3. When to look up syntax (always!)

**Professional developers Google syntax daily. You should too.**
