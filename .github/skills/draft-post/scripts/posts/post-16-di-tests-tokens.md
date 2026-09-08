---
title: My Tests Required a Running Database. Interfaces Fixed That.
subtitle: Part 16 of Learning Go in Public
tags: go, testing, backend, buildinpublic
series: Learning Go in Public
---

In [Part 15](https://dev.to/abhishek_sharma_a9792aee8/i-deployed-my-go-backend-to-a-real-vps-heres-exactly-what-happened-2dpj), I deployed the backend to a real VPS. But there was a problem I'd been ignoring: the test suite needed a running PostgreSQL database and a Redis instance just to run. No database? Tests skip. No Redis? Panic.

That's not unit testing. That's integration testing pretending to be unit testing.

## The Problem: Global Variables Everywhere

Every handler was a package-level function that reached directly into global state:

```go
// Before — handlers called globals directly
func CreateEntry(w http.ResponseWriter, r *http.Request) {
    id, err := db.InsertEntry(r.Context(), userID, req.Text, req.Mood, req.Category)
    // ...
    cache.Delete(ctx, cacheKey)
}
```

`db.InsertEntry` called the global `db.DB` connection. `cache.Delete` called the global `redis.Client`. The handler was welded to its dependencies — no way to swap them out.

This meant:
- Tests needed a running Postgres → slow, flaky, CI requires a database service
- Tests needed Redis (or at least a non-nil client) → without it, `cache.Delete` panicked
- Can't test a handler without testing the entire stack underneath it

## The Fix: Interfaces + Dependency Injection

### Step 1: Define What Handlers Need

Instead of importing `db` and calling `db.InsertEntry`, define an interface — "here's what I need, I don't care how you implement it":

```go
// internal/handlers/handler.go
type Store interface {
    InsertEntry(ctx context.Context, userID int, text string, mood int, category string) (int64, error)
    GetEntryByID(ctx context.Context, entryID int, userID int64) (*models.Entry, error)
    GetAllEntriesPaginated(ctx context.Context, userID int64, limit int, offset int) ([]models.Entry, error)
    GetEntryCount(ctx context.Context, userID int64) (int, error)
    UpdateEntry(ctx context.Context, entryID int, userID int64, text string, mood int, category string) (int64, error)
    DeleteEntry(ctx context.Context, entryID int, userID int64) (int64, error)
    CreateUser(ctx context.Context, email string, passwordHash string) (int64, error)
    GetUserByEmail(ctx context.Context, email string) (int64, string, error)
}

type Cache interface {
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Get(ctx context.Context, key string) (string, error)
    Del(ctx context.Context, key string) error
}

type Handler struct {
    Store     Store
    Cache     Cache
    JWTSecret string
}
```

The `Handler` struct holds its dependencies. It doesn't know (or care) if `Store` is backed by Postgres, SQLite, or a fake in-memory map.

### Step 2: Convert Functions to Methods

Every handler changed from a package-level function to a method on `*Handler`:

```go
// Before — package-level function, calls global db directly
func CreateEntry(w http.ResponseWriter, r *http.Request) {
    id, err := db.InsertEntry(...)
}

// After — method on Handler, calls the injected Store
func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
    id, err := h.Store.InsertEntry(...)
}
```

One character changed in the function signature (`h *Handler`). Every `db.Something()` became `h.Store.Something()`. Every `cache.Something()` became `h.Cache.Something()`.

The handler's logic didn't change at all. It still validates the request, calls the store, invalidates cache, returns JSON. It just doesn't know which store it's calling.

### Step 3: Wire Real Dependencies in main.go

```go
// main.go — production wiring
h := &handlers.Handler{
    Store:     &db.PostgresStore{},
    Cache:     redis.NewRedisCache(),
    JWTSecret: cfg.JWTSecret,
}

// Routes use h.Method instead of handlers.Function
http.HandleFunc("/register", middleware(h.Register))
http.HandleFunc("/login", middleware(h.Login))
```

`db.PostgresStore` already had methods like `InsertEntry` — they just needed to be on a struct instead of being standalone functions. Same with `redis.RedisCache` implementing `Set/Get/Del`.

### Step 4: Build FakeStore for Tests

This is where it pays off. A `FakeStore` that satisfies the `Store` interface with zero infrastructure:

```go
// fake_store_test.go — only exists in test files
type FakeStore struct {
    InsertEntryFn func() (int64, error)
    users         map[string]struct {
        id           int64
        passwordHash string
    }
}

func (f *FakeStore) InsertEntry(ctx context.Context, userID int, text string, mood int, category string) (int64, error) {
    return f.InsertEntryFn()  // returns whatever the test tells it to
}

func (f *FakeStore) CreateUser(ctx context.Context, email string, passwordHash string) (int64, error) {
    if _, exists := f.users[email]; exists {
        return 0, &pq.Error{Code: "23505"}  // simulate duplicate email
    }
    id := int64(len(f.users) + 1)
    f.users[email] = struct{ id int64; passwordHash string }{id, passwordHash}
    return id, nil
}

type FakeCache struct{}
func (f *FakeCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error { return nil }
func (f *FakeCache) Get(ctx context.Context, key string) (string, error) { return "", fmt.Errorf("cache miss") }
func (f *FakeCache) Del(ctx context.Context, key string) error { return nil }
```

`FakeCache` always returns "cache miss." `FakeStore` returns whatever the test configures via `InsertEntryFn`. No database. No Redis. No network calls. Tests run in milliseconds.

### Step 5: Tests That Don't Need Infrastructure

```go
func TestCreateEntry(t *testing.T) {
    h := &Handler{
        Store: &FakeStore{InsertEntryFn: func() (int64, error) {
            return 1, nil  // simulate successful insert
        }},
        Cache: &FakeCache{},
    }

    body := strings.NewReader(`{"text":"feeling good","mood":8,"category":"mood"}`)
    req := httptest.NewRequest(http.MethodPost, "/entries", body)
    ctx := context.WithValue(req.Context(), userIDKey, int64(1))
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    h.CreateEntry(rec, req)

    if rec.Code != http.StatusCreated {
        t.Errorf("expected 201, got %d", rec.Code)
    }
}
```

Three lines to set up the handler. No `TestMain` initializing a database. No env vars. No cleanup. `go test` runs instantly with zero external dependencies.

## Token Refresh: Fixing the 24-Hour Session Problem

While refactoring handlers, I also fixed the authentication model. The old system issued a single JWT valid for 24 hours. If it leaked, an attacker had access for an entire day.

The new model:
- **Access token**: JWT, 15 minutes, stateless (server doesn't store it)
- **Refresh token**: UUID, 7 days, stored in Redis (server can revoke it)

### Login Now Returns Two Tokens

```go
// Access token — short-lived, for API calls
claims := jwt.MapClaims{
    "user_id": userID,
    "exp":     time.Now().Add(15 * time.Minute).Unix(),
}
tokenString, _ := token.SignedString([]byte(secret))

// Refresh token — long-lived, stored in Redis
refreshUUID := uuid.New().String()
key := "refresh_token:" + refreshUUID
redis.Client.Set(ctx, key, fmt.Sprintf("%d", userID), 7*24*time.Hour)

// Return both
json.NewEncoder(w).Encode(LoginResponse{
    AccessToken:  tokenString,
    RefreshToken: refreshUUID,
})
```

### POST /refresh — Rotate Tokens

When the access token expires, the client sends the refresh token:

```go
func Refresh(w http.ResponseWriter, r *http.Request) {
    key := "refresh_token:" + req.RefreshToken
    val, err := redis.Client.Get(ctx, key).Result()
    if err != nil {
        // Invalid or expired → force re-login
        errorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token")
        return
    }
    userID, _ := strconv.ParseInt(val, 10, 64)

    // Delete the old refresh token (one-time use)
    redis.Client.Del(ctx, key)

    // Issue new access token + new refresh token
    // ... (same JWT creation as login)
}
```

The old refresh token is deleted immediately — each refresh token is single-use. If an attacker steals a refresh token and the real user uses it first, the attacker's stolen token is already invalidated.

### POST /logout — Server-Side Revocation

```go
redis.Client.Del(ctx, "refresh_token:" + req.RefreshToken)
```

One line. Delete the refresh token from Redis. The access token will expire in ≤15 minutes naturally. No token blacklist needed.

## Per-User Rate Limiting

The existing rate limiter tracked by IP address. Problem: behind a shared network (office, university), hundreds of users share one IP. One heavy user triggers the limit for everyone.

The fix is simple once you have authenticated users with user IDs in context:

```go
func UserRateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, ok := r.Context().Value(userIDKey).(int64)
        if !ok {
            errorResponse(w, http.StatusUnauthorized, "User not authenticated")
            return
        }

        key := fmt.Sprintf("user:%d", userID)

        if !IsAllowed(key, UserRateLimitRequests, UserRateLimitWindow) {
            http.Error(w, "Rate limit exceeded. Try again later.", http.StatusTooManyRequests)
            return
        }

        next(w, r)
    }
}
```

The Redis key changed from the IP address to `user:123`. Same `IsAllowed` function, same fixed-window algorithm, different key. The global rate limiter still exists on unauthenticated routes (`/login`, `/register`) — it runs before `AuthMiddleware`. The per-user limiter runs after, on protected routes.

Two layers:
1. IP-based → blocks brute-force login attempts
2. User-based → blocks individual users who abuse the API

## What Changed Under the Hood

The middleware chain in `main.go` tells the full story:

```go
// Unauthenticated routes — IP rate limiting
http.HandleFunc("/login", RateLimitMiddleware(LoggingMiddleware(h.Login)))

// Authenticated routes — IP + user rate limiting
http.HandleFunc("/entries", RateLimitMiddleware(AuthMiddleware(
    UserRateLimitMiddleware(h.CreateEntry),
)))
```

Request hits IP rate limit first → then auth middleware extracts user_id → then per-user rate limit checks against that user's quota → then the handler runs with injected dependencies.

## What I Learned

**Interfaces in Go aren't for abstraction — they're for testability.** I didn't define `Store` because I might switch databases. I defined it because my tests shouldn't need a database to verify that a handler returns 400 on invalid input.

**The _test.go suffix is powerful.** `FakeStore` lives in `fake_store_test.go`. Go only compiles `_test.go` files during `go test`. The fakes don't ship in the production binary — zero runtime cost.

**Token refresh isn't optional.** A 24-hour JWT means a leaked token is valid for a full day. A 15-minute access token with a revocable refresh token means you can kill any session immediately via `/logout`.

**Rate limiting changes meaning after authentication.** Before auth, you're limiting IPs (blunt). After auth, you're limiting users (precise). Both are necessary — they protect against different attack vectors.

---

*This is Part 16 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i) | [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g) | [Part 8](https://dev.to/abhishek_sharma_a9792aee8/i-built-my-backend-on-sqlite-then-i-deleted-it-15a) | [Part 9](https://dev.to/abhishek_sharma_a9792aee8/my-database-disappeared-every-time-i-restarted-docker-migrations-fixed-that-n93) | [Part 10](https://dev.to/abhishek_sharma_a9792aee8/every-request-looked-the-same-in-my-logs-then-i-added-one-line-19m7) | [Part 11](https://dev.to/abhishek_sharma_a9792aee8/i-had-no-idea-which-endpoint-was-slowest-so-i-built-my-own-metrics-5fjp) | [Part 12](https://dev.to/abhishek_sharma_a9792aee8/my-backend-crashed-every-time-redis-went-down-three-patterns-fixed-that-43j2) | [Part 13](https://dev.to/abhishek_sharma_a9792aee8/i-needed-to-send-an-http-request-without-slowing-down-my-api-goroutines-fixed-that-20o5) | [Part 14](https://dev.to/abhishek_sharma_a9792aee8/my-webhooks-were-sending-data-anyone-could-fake-hmac-signing-fixed-that-gpm) | [Part 15](https://dev.to/abhishek_sharma_a9792aee8/i-deployed-my-go-backend-to-a-real-vps-heres-exactly-what-happened-2dpj)*
