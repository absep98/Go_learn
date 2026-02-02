# Learning Journey

**My transformation from SDET to Backend Developer - documented week by week.**

---

# Week 1: Foundation (Jan 2-7, 2026)

## What I Built ✅

- HTTP server with `net/http` standard library
- SQLite database connection and schema design
- CRUD operations (POST /entries, GET /entries)
- Input validation patterns
- SQL injection prevention (parameterized queries)
- JSON struct tags for API mapping
- Error handling with proper HTTP status codes

## Major Challenges 🤔

### 1. Confidence Crisis (Day 4)

**The problem:** Felt like "I won't be able to code this on my own." Worried about relying too much on guidance.

**The breakthrough:** When asked to add a new field, I correctly identified all 3 layers that needed changes (struct, handler, database). This proved I understood the architecture—learning is pattern recognition, not memorization.

### 2. JSON Struct Tags Confusion

**The problem:** Thought tags were optional or just "lowercase versions" of field names.

**The fix:** Ran practice examples showing exact matching rules. Now understand tags map JSON keys to Go fields precisely. Case-sensitive. Tag value must match API exactly.

### 3. SQL Injection (Had It Backwards!)

**The mistake:** Thought parameterized queries (?) were the vulnerable approach.

**The fix:** Ran attack demo that actually destroyed a test database using string concatenation. Saw the exploit work in real-time. Now viscerally understand why parameterized queries matter.

## Key Moments 💡

**Day 4 - Architecture Understanding:**
> "We need to change our model file, entries file which handle the logic of db interaction, and then db.go too as need to add table entries or update the table with new column"

This statement proved I understood the 3-layer flow without realizing it.

**Day 6 - Security Demo:**
Watching SQL injection delete an entire table made security concepts concrete. Never forgetting this lesson.

## Skills Acquired

**Technical:**

- REST API design with proper HTTP methods
- Database CRUD operations
- Input validation (sequential checks with early returns)
- Security awareness (parameterized queries prevent injection)
- Error handling with multiple return values
- JSON serialization/deserialization

**Conceptual:**

- 3-layer architecture (Handler → Database → Storage)
- Separation of concerns for testability
- Request/response lifecycle
- Database driver registration (blank imports)

**Confidence:**

- Can read and understand Go backend code
- Can modify existing features
- Can explain architectural decisions
- Can debug issues systematically

## Time Investment

- **Total:** ~9 hours (6 days × 1.5hr)
- **Outcome:** Working backend with 2 endpoints, clean architecture, security awareness
- **ROI:** High - went from zero backend code to interview-ready project in 6 days

## Confidence Progression

**Before Week 1:** 3/10

- Felt like "just SDET"
- Scared of backend work
- Thought I needed to memorize everything

**After Week 1:** 7/10

- Can build and explain backend systems
- Understand security fundamentals
- Comfortable modifying and debugging code
- Know what I don't know (auth, concurrency)

---

# Week 2: Authentication System (Jan 8-12, 2026)

## What I Built ✅

**Days 8-10: Core Authentication**

- User registration with email validation
- bcrypt password hashing (never store plain text!)
- JWT token generation on login
- Authentication middleware
- Protected API routes
- User-specific data isolation

**Days 11-12: Polish & Testing**

- Better error messages (specific token errors)
- Request logging middleware
- Comprehensive test suite (18 tests, all passing)
- Complete documentation (README, API-ENDPOINTS, TESTING)

## Major Breakthroughs 💡

### 1. Understanding JWT Flow

**What clicked:** JWT tokens are like "tickets" that prove you logged in. The server doesn't need to remember you logged in—the signed token proves it.

**The flow:**

1. User logs in → Server creates token with user_id inside
2. User sends token with every request
3. Server verifies signature → extracts user_id → uses it

**Why it matters:** Stateless authentication scales better (no session storage needed)

### 2. Middleware Pattern

**What clicked:** Middleware is like a security checkpoint. Request comes in, middleware checks it, then passes it to the actual handler.

```go
LoggingMiddleware(AuthMiddleware(ActualHandler))
```

This composition is powerful—can add features without changing handler code.

### 3. Context for Passing Data

**The problem:** How does middleware pass user_id to handler?

**The solution:** Go's context! Middleware puts data in context, handler retrieves it.

```go
// Middleware
ctx := context.WithValue(r.Context(), "user_id", userID)
next(w, r.WithContext(ctx))

// Handler
userID := r.Context().Value("user_id").(int64)
```

**Why it matters:** Clean way to pass request-scoped data without global variables.

### 4. Type Conversions in Go

**The challenge:** JWT claims return numbers as `float64` (JSON limitation), but database expects `int`.

**The chain:**

```
JWT (float64) → Middleware (int64) → Handler → Database (int)
```

**What I learned:** Go's strict typing catches bugs. The type mismatch error actually prevented a subtle bug.

## Major Challenges 🤔

### 1. Cognitive Overload (Day 9 Evening)

**The feeling:** "Things are piling up, I'm not remembering syntax, everything feels complex."

**What helped:**

- Agent created ARCHITECTURE-MAP.md and CHEAT-SHEET.md
- Stopped learning new things, just read/reviewed existing code
- Tested API manually to see it working
- Rested for a day

**Lesson:** Learning needs consolidation time. Sometimes best action is STOP and review, not push forward.

### 2. Type System Confusion

**The error:** `cannot use userID (int64) as int`

**What I learned:**

- Go doesn't auto-convert even between int types
- JWT stores numbers as float64 (JSON)
- Need explicit conversion: `int(userID)`

**Why it matters:** Type safety prevents bugs. This strictness is a feature, not a bug.

### 3. Understanding Middleware Wrapping

**Initial confusion:** How does `AuthMiddleware(handler)` actually work?

**What clicked:** It returns a NEW function that:

1. Does auth check
2. Calls original handler if auth passes

It's function composition—wrapping one function with another.

## Skills Gained ✅

**Technical Skills:**

- JWT token generation and verification
- bcrypt password hashing
- Middleware pattern implementation
- Context usage for request-scoped data
- Type conversions in Go
- Comprehensive testing strategies
- Better error message design

**Conceptual Understanding:**

- Stateless authentication vs sessions
- Middleware composition pattern
- Request lifecycle with middleware layers
- Security: Why hash passwords, why sign tokens
- User data isolation importance
- Test-driven development benefits

**Debugging Skills:**

- Reading Go compiler errors more effectively
- Understanding type mismatch errors
- Using logs to trace request flow
- Testing edge cases systematically

## What I'd Do Differently Next Time

### 1. Start Testing Earlier

I built comprehensive tests on Day 12, but having tests from Day 8 would have:

- Caught bugs earlier
- Given confidence during development
- Served as documentation

**Next time:** Write tests as I build features, not after.

### 2. Draw Architecture First

ARCHITECTURE-MAP.md was incredibly helpful, but created AFTER building.

**Next time:** Draw flow diagrams BEFORE coding to clarify thinking.

### 3. Take More Breaks

Day 9 overload could have been avoided with:

- More frequent consolidation breaks
- Reviewing previous day's code before adding new features
- Not pushing through when feeling overwhelmed

**Lesson:** Rest is part of learning, not a break from it.

## Confidence Progression

**Before Week 2:**

- ✅ Could build CRUD APIs
- ✅ Understood database operations
- ❌ Authentication felt intimidating
- ❌ Middleware was abstract concept
- ❌ Didn't understand JWT

**After Week 2:**

- ✅ Can implement JWT authentication from scratch
- ✅ Can explain middleware pattern clearly
- ✅ Understand security basics (hashing, signing)
- ✅ Can debug type system errors
- ✅ Can write comprehensive tests

**Most importantly:** Can explain HOW it works, not just make it work.

## Interview-Ready Stories

### Story 1: "How I Implemented Authentication"

"I built a JWT-based authentication system in Go. Users register with email/password, I hash the password with bcrypt, never storing plain text. On login, I verify the password and generate a JWT token containing their user_id, signed with a secret key.

For protected routes, I use middleware that verifies the token signature, extracts the user_id, and passes it via Go's context to the handler. This ensures users can only access their own data."

### Story 2: "Debugging Type Mismatch"

"I encountered an error where JWT claims returned user_id as float64 (JSON limitation), but my database function expected int. Go's strict typing caught this. I had to explicitly convert: float64 → int64 (in middleware for safety) → int (for database call).

This taught me Go's type system prevents subtle bugs. The compiler error was actually protecting me."

### Story 3: "Cognitive Overload Recovery"

"During Week 2 Day 9, I felt overwhelmed by syntax complexity. Instead of pushing through, I stopped, created reference documentation (ARCHITECTURE-MAP and CHEAT-SHEET), and spent a day just reviewing existing code.

This consolidation made everything click. I realized I didn't need to memorize—I needed to understand the patterns. When I returned to coding, everything felt 50% less complex."

---

# Cumulative Progress

## Technical Capabilities

**Week 1:**

- ✅ HTTP server setup
- ✅ Database integration
- ✅ CRUD operations
- ✅ Input validation
- ✅ Security fundamentals

**Week 2:**

- ✅ User authentication
- ✅ JWT implementation
- ✅ Middleware patterns
- ✅ Protected routes
- ✅ Comprehensive testing

**Total:** Production-ready backend with authentication, security, and testing.

## Confidence Trajectory

```
Week 0: 3/10 (Just SDET, scared of backend)
Week 1: 7/10 (Can build CRUD, understand security)
Week 2: 9/10 (Can implement auth, explain systems, ready for interviews)
```

## Key Lessons Learned

1. **Authentication isn't magic** - It's just: hash passwords, sign tokens, verify tokens
2. **Middleware is powerful** - Separation of concerns makes code cleaner
3. **Type safety is a feature** - Go's strictness prevents bugs
4. **Testing gives confidence** - Comprehensive tests = sleep better at night
5. **Rest is learning** - Brain needs time to consolidate
6. **Documentation helps** - Architecture maps and cheat sheets reduce cognitive load
7. **Pattern recognition > Memorization** - Understand the flow, look up syntax
8. **Debugging is a skill** - Compiler errors are helpful, not scary

## Interview Readiness

**Can I explain my project?** ✅ YES

> "I built a Go REST API with JWT authentication, CRUD operations, and user-specific data isolation. It uses a 3-layer architecture separating HTTP handlers, database logic, and storage. The API validates input, uses parameterized queries to prevent SQL injection, and returns structured JSON responses. All protected routes use middleware for authentication."

**Technical questions I can answer:**

- ✅ How does JWT work?
- ✅ Why hash passwords with bcrypt?
- ✅ What is SQL injection and how to prevent it?
- ✅ How does middleware work in Go?
- ✅ Why use JSON struct tags?
- ✅ How to handle errors in Go?
- ✅ What is context and when to use it?

**System design questions I can answer:**

- ✅ Why separate layers?
- ✅ Why stateless authentication?
- ✅ How to scale authentication?
- ✅ What breaks first at scale?
- ✅ Why this database choice?

## Time Investment Summary

- **Week 1:** ~9 hours (1.5hr/day × 6 days)
- **Week 2:** ~9 hours (1.5hr/day × 6 days)
- **Total:** ~18 hours
- **Outcome:** Production-ready backend, interview-ready skills, confidence to apply for backend roles

**ROI:** Extremely high - transformed from "just SDET" to "backend developer" in 2 weeks.

---

# Week 3: Scaling & Production Features (Jan 13-21, 2026)

## What I Built ✅

**Days 13-15: CRUD Completion & Pagination**

- UPDATE operation (PATCH /entries?id=X)
- DELETE operation (DELETE /entries?id=X)
- Pagination with page/limit parameters
- User-specific data isolation for all operations

**Days 16-17: Caching Layer**

- In-memory cache with TTL (time-to-live)
- Mutex for thread safety
- Cache invalidation on create/update/delete
- COUNT query optimization

**Day 18: Rate Limiting**

- Fixed window rate limiting algorithm
- 100 requests per minute per IP
- Return 429 Too Many Requests when exceeded
- System design notes on distributed rate limiting

**Days 19-20: Redis Integration**

- Docker Desktop installation
- Redis container setup
- Converted cache from in-memory to Redis
- Converted rate limiter from in-memory to Redis
- Learned Redis commands (SET, GET, INCR, EXPIRE, etc.)
- Fixed IP extraction from RemoteAddr (port removal)

## Key Concepts Learned 🧠

| Concept | What I Learned |
|---------|---------------|
| **Pagination** | LIMIT/OFFSET vs cursor-based, deep paging problems |
| **Caching** | TTL, cache invalidation, thundering herd |
| **Rate Limiting** | Fixed window vs sliding window vs token bucket |
| **Redis** | Centralized storage, atomic operations (INCR), auto-TTL |
| **Docker** | Containers, images, port mapping, exec commands |
| **Distributed Systems** | Why in-memory doesn't scale, shared storage solutions |

## Challenges & Breakthroughs 💡

**Challenge 1: Redis keys not appearing**

- Made requests but `KEYS *` showed empty
- Cause: Server running old code (stale process)
- Fix: Kill old process, rebuild, restart

**Challenge 2: IP:Port in rate limit key**

- `RemoteAddr` includes port (e.g., `[::1]:54321`)
- Each connection got different key = broken rate limiting
- Fix: `net.SplitHostPort()` to extract just IP

**Breakthrough: In-Memory → Redis Pattern**

```go
// BEFORE (in-memory)
c.mu.Lock()
c.data[key] = value
c.mu.Unlock()

// AFTER (Redis)
redis.Client.Set(ctx, key, value, ttl)
// No mutex needed - Redis handles concurrency!
```

## System Design Questions I Can Answer

1. "What breaks at 10,000 concurrent requests?" → Deep paging, cache stampede, DB concurrency
2. "How would you implement rate limiting for distributed systems?" → Redis INCR with TTL
3. "Why not use in-memory caching in production?" → Data loss on restart, not shared across servers

**Day 21: Graceful Shutdown**

- `os/signal` package to catch Ctrl+C (SIGINT)
- Channels to communicate shutdown signal
- `context.WithTimeout` for 5-second shutdown deadline
- `server.Shutdown(ctx)` to finish in-flight requests
- Proper cleanup of Redis and DB connections

**Day 22: Enhanced Health Checks**

- Health endpoint pings Redis AND database
- Returns structured JSON with component status
- Returns 200 when healthy, 503 when any service down
- Production-ready monitoring endpoint

**Day 23: Background Worker Pool**

- Job struct with Type and Payload
- Buffered channel as job queue (capacity 100)
- Goroutines as workers processing jobs concurrently
- Non-blocking job submission with `select`
- Triggered on entry creation for async processing

## Key Concepts Learned 🧠

| Concept | What I Learned |
|---------|---------------|
| **Graceful Shutdown** | Stop accepting new requests, finish current ones, cleanup resources |
| **os/signal** | Notify channel when OS sends interrupt signal |
| **Goroutines** | Lightweight threads - `go func()` runs in background |
| **Channels** | Pipes for goroutine communication - `chan Job` |
| **Buffered Channels** | `make(chan Job, 100)` - holds 100 items before blocking |
| **range over channel** | `for job := range JobQueue` blocks until job arrives |
| **select statement** | Non-blocking channel operations with default case |
| **context.WithTimeout** | Automatic cancellation after deadline |

## Breakthroughs 💡

**Breakthrough 1: Understanding Goroutines + Channels Together**

```go
// The worker pool pattern:
JobQueue := make(chan Job, 100)  // Create the "pipe"

go worker(id)  // Start worker in background

for job := range JobQueue {  // Worker blocks here, waiting for jobs
    processJob(job)           // When job arrives, process it
}

JobQueue <- Job{...}  // Add job to queue - wakes up a worker!
```

**Analogy that clicked:** Pizza shop with multiple chefs
- `JobQueue` = order counter where tickets pile up
- `worker` goroutines = chefs waiting for orders
- `range JobQueue` = chef grabs next ticket when free
- `go worker(id)` = hire a chef to work in background

**Breakthrough 2: Graceful Shutdown Flow**

```
Ctrl+C pressed
    ↓
signal.Notify sends to 'quit' channel
    ↓
<-quit unblocks main()
    ↓
server.Shutdown(ctx) called
    ↓
Server stops accepting NEW requests
    ↓
Waits for current requests (up to 5 seconds)
    ↓
defer statements run (close Redis, DB)
    ↓
Program exits cleanly
```

**Breakthrough 3: Health Check Pattern**

```go
// Don't re-initialize! Ping existing connection
err := redis.Client.Ping(ctx).Err()
if err != nil {
    redisStatus = "disconnected"
    healthy = false
}
```

## Confidence Crisis & Recovery (Day 21) 🧘

**The fear:** "I got that fear again that I'll not be able to build this on my own"

**The test:** Asked to answer 3 conceptual questions without help

**Results:**
1. ✅ Knew `context.WithTimeout` creates deadline-based cancellation
2. ✅ Understood `server.Shutdown` stops new requests, finishes current
3. ✅ Knew goroutines don't share memory automatically (channels needed)

**Lesson:** I understand concepts even when syntax feels shaky. Pattern recognition > memorization.

## Interview-Ready Stories (Week 3)

### Story 1: "How I Handle Graceful Shutdown"

"When my Go server receives Ctrl+C, I don't just kill it. I use `os/signal.Notify` to catch the signal, then call `server.Shutdown` with a 5-second timeout context. This stops accepting new requests while letting in-flight requests complete. Finally, deferred statements close Redis and database connections cleanly. No dropped requests, no resource leaks."

### Story 2: "Why I Migrated to Redis"

"Initially I used in-memory maps with mutexes for caching and rate limiting. But I realized: if I deploy two servers, each has its own cache—users hit different rate limits per server! Redis solves this: it's a centralized store all servers share. Plus, Redis INCR is atomic, so I don't need mutexes anymore. It was a key lesson in thinking beyond single-server architecture."

### Story 3: "Implementing Background Workers"

"When a user creates an entry, I don't want them waiting for slow tasks like sending notifications. So I built a worker pool: a buffered channel holds jobs, and 3 goroutines constantly `range` over that channel. The handler just adds a job to the queue and responds immediately. Workers process jobs in background. It's the producer-consumer pattern in Go, using channels for thread-safe communication."

---

# Week 4 Plan (Feb 2+)

## Focus: Polish & Deployment Readiness

**Day 24-25: Structured Logging**
- Replace `log.Printf` with structured JSON logging
- Add request IDs for tracing
- Log levels (INFO, WARN, ERROR)
- Learn `slog` package (Go 1.21+)

**Day 26-27: Metrics & Monitoring**
- Add `/metrics` endpoint
- Track request counts, response times
- Prometheus-style metrics format
- Understand observability basics

**Day 28-29: Configuration & Environment**
- Move all config to environment variables
- Add configuration validation on startup
- Different configs for dev/prod
- 12-Factor App principles

**Day 30: Documentation & Review**
- Update API documentation
- Create deployment guide
- Final architecture review
- Interview prep consolidation

---

# Cumulative Progress

## Technical Capabilities

**Week 1:** HTTP server, SQLite, CRUD, validation, security fundamentals  
**Week 2:** JWT auth, middleware, protected routes, testing  
**Week 3:** Pagination, caching, rate limiting, Redis, graceful shutdown, worker pools

**Total:** Production-ready backend with auth, caching, rate limiting, background processing, and proper shutdown handling.

## Confidence Trajectory

```
Week 0: 3/10 (Just SDET, scared of backend)
Week 1: 7/10 (Can build CRUD, understand security)
Week 2: 9/10 (Can implement auth, explain systems)
Week 3: 9/10 (Can scale systems, use Redis, goroutines)
```

---

# Quote That Changed My Mindset

> "You don't need to memorize syntax. You need to understand: 1) What each file does, 2) How data flows between them, 3) When to look up syntax (always!)"

This shifted my thinking from "I should know everything" to "I should understand the system."

---

# Final Achievement Statement

**In 3 weeks, I built a production-ready backend** with JWT authentication, CRUD operations, pagination, Redis caching, rate limiting, graceful shutdown, and background worker pools.

**I can confidently say in interviews:**

> "I designed and built a Go backend service with JWT auth, Redis caching, rate limiting, graceful shutdown, and async worker pools. It handles concurrent requests with goroutines, uses channels for job queues, and properly cleans up resources on shutdown. I understand both the code AND the system design trade-offs."

---

**Status:** Week 3 Complete ✅
**Confidence Level:** 9/10
**Ready for:** Backend Developer Interviews
**Next:** Week 4 (Polish & Deployment Readiness)
