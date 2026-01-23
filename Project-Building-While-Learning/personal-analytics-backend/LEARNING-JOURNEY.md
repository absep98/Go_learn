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

---

# Week 4 Plan (Jan 22+)

## Day 21: Graceful Shutdown

- Handle Ctrl+C signal properly
- Close Redis and DB connections cleanly
- Learn `os/signal` package
- Production-ready server lifecycle

## Day 22+: Background Workers / Async Processing

- Worker pool pattern
- Goroutines for async tasks
- Job queues with Redis
- Non-blocking operations

## Week 4 Review (Consolidation)

- Document all Week 3-4 learnings
- Update interview stories
- Create system design summaries
- Architecture diagram updates

---

# What's Next

## Week 3 Preview (Scaling & Polish)

**Planned Features:**

- UPDATE and DELETE operations
- Pagination for GET /entries
- Caching layer
- Rate limiting
- Background job processing

**Learning Focus:**

- System design thinking (scaling considerations)
- Performance optimization
- Async processing with goroutines
- Production deployment considerations

---

# Quote That Changed My Mindset

> "You don't need to memorize syntax. You need to understand: 1) What each file does, 2) How data flows between them, 3) When to look up syntax (always!)"

This shifted my thinking from "I should know everything" to "I should understand the system."

---

# Final Achievement Statement

**In 2 weeks, I built a production-ready authentication system** with registration, login, JWT tokens, middleware protection, user-specific data isolation, and comprehensive testing.

**I can confidently say in interviews:**

> "I designed and built a backend service in Go with JWT auth, middleware protection, user-specific data isolation, and security best practices. It includes comprehensive testing with 18 test scenarios covering registration, login, authentication, and data operations."

That sentence alone changes my backend developer profile.

---

**Status:** Week 2 Complete ✅
**Confidence Level:** 9/10
**Ready for:** Backend Developer Interviews
**Next:** Week 3 (Scaling & Advanced Features)
