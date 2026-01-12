
# Week 1 Review (Jan 2-7, 2026)

## What Went Well ✅

- **Completed all 6 days on schedule** - No skipped days, consistent 1.5hr sessions
- **Built working CRUD API** - POST and GET endpoints functional with validation
- **Learned security fundamentals** - SQL injection prevention through hands-on demo
- **Understood JSON struct tags deeply** - Can explain matching rules and omitempty
- **Fixed bugs independently** - Found defer placement issue, missing return statements
- **Overcame self-doubt** - Day 4 crisis turned into architectural understanding breakthrough

## Challenges I Faced 🤔

### 1. Confidence Crisis (Day 4)
Felt like "I won't be able to code this on my own" after reaching Day 4. Worried about relying too much on guidance.

**Resolution:** Realized learning is pattern recognition, not memorization. When asked to add the category field, I correctly identified all 3 layers that needed changes—proving I understood the architecture.

### 2. JSON Struct Tags Confusion
Initially thought tags were optional or just "lowercase versions" of field names.

**Resolution:** Ran practice examples showing exact matching rules. Now understand tags map JSON keys to Go fields precisely.

### 3. SQL Injection Backwards
Thought parameterized queries (?) were the vulnerable approach.

**Resolution:** Ran attack demo that actually destroyed a test database using string concatenation. Saw the exploit work in real-time.

## Key Breakthroughs 💡

### Day 4: Architecture Understanding
When asked how to add a new field, I said:
> "We need to change our model file, entries file which handle the logic of db interaction, and then db.go too as need to add table entries or update the table with new column"

This proved I understood the 3-layer flow without realizing it.

### Day 6: SQL Injection Demo
Watching the attack delete an entire table made security concepts concrete. Now I viscerally understand why parameterized queries matter.

### Day 6: JSON Tags Practice
Running 5 different examples with various JSON formats clarified exact vs case-sensitive matching. The `omitempty` usage finally clicked.

## Skills Acquired

**Technical:**
- REST API design with proper HTTP methods and status codes
- Database CRUD operations (Create, Read)
- Input validation patterns (sequential checks with early returns)
- Security awareness (parameterized queries prevent injection)
- Error handling with multiple return values
- JSON serialization/deserialization with struct tags

**Conceptual:**
- 3-layer architecture (Handler → Database → Storage)
- Separation of concerns for testability
- Request/response lifecycle understanding
- Database driver registration (blank imports)

**Confidence:**
- Can read and understand Go backend code
- Can modify existing features (add fields, change validation)
- Can explain architectural decisions
- Can debug issues systematically

## Week 2 Goals

### Primary Objectives
1. **JWT Authentication** - Token generation, validation, middleware
2. **Goroutines & Concurrency** - Async processing, worker pools
3. **User Isolation** - Each user sees only their data
4. **More Endpoints** - UPDATE and DELETE operations

### Secondary Goals
- Write basic tests for handlers
- Add database indexes for performance
- Implement better error handling (centralized responses)
- Add rate limiting (if time permits)

## Confidence Level

**Before Week 1:** 3/10
- Felt like "just SDET"
- Scared of backend work
- Thought I needed to memorize everything

**After Week 1:** 7/10
- Can build and explain backend systems
- Understand security fundamentals
- Comfortable modifying and debugging code
- Know what I don't know (auth, concurrency)

**Gap to 10/10:** Need more experience with auth, concurrency, and real production concerns (caching, scaling).

## Interview Readiness

### Can I explain my project? ✅ YES
> "I built a Go REST API with CRUD operations. It uses a 3-layer architecture separating HTTP handlers, database logic, and storage. The API validates input, uses parameterized queries to prevent SQL injection, and returns structured JSON responses."

### Can I explain SQL injection prevention? ✅ YES
> "I use parameterized queries with placeholders instead of concatenating user input into SQL strings. The database driver handles escaping automatically, preventing attackers from injecting malicious SQL commands."

### Can I explain JSON struct tags? ✅ YES
> "Struct tags provide metadata for JSON marshaling. The `json` tag maps between Go field names and JSON keys, handling different naming conventions like snake_case in APIs vs PascalCase in Go."

### Can I explain why separation matters? ✅ YES
> "Separating handlers and database logic improves testability (can mock database), reusability (same DB functions work for API, CLI, jobs), and maintainability (switching databases only changes one layer)."

### What can't I explain yet? 🟨 NEXT WEEK
- JWT token validation flow
- Goroutine synchronization patterns
- Caching strategies
- Load balancing concepts

## Time Investment

- **Total hours:** ~9 hours (6 days × 1.5hr)
- **Outcome:** Working backend with 2 endpoints, clean architecture, security awareness
- **ROI:** High - went from zero backend code to interview-ready project in 6 days

## Mistakes Made & Fixed

1. **Half-implemented Tags feature** - Added to struct but not database schema
   - **Fix:** Removed completely, will add properly in Week 2

2. **Missing return after validation** - Error sent but code continued
   - **Fix:** Added `return` after all error responses

3. **Models layer unused** - Created folder but used handlers directly
   - **Fix:** Will refactor in Week 2 when adding more complexity

## What I'd Do Differently

**If starting over:**
1. Test each feature immediately after building (I waited until Day 5)
2. Ask "why" questions earlier (understanding vs just copying)
3. Document learnings daily (not just at end of week)

**But overall:** The structure worked. Daily tasks kept momentum, assessment revealed real gaps, and the learning-by-building approach stuck better than tutorials would have.

## Proof of Learning

**Repo:** [personal-analytics-backend](.)
**Demos:**
- [SQL Injection Demo](learning-demos/01-sql-injection-demo.go)
- [JSON Struct Tags Practice](learning-demos/02-json-struct-tags-practice.go)

**Test Results:** Server runs, POST creates entries, GET retrieves them, validation rejects bad input.

## Next Action (Day 7 - Jan 8)

**Week 2 starts tomorrow.** Focus: Authentication.

Day 7 task will be: Research JWT, understand token structure, plan auth middleware.

---

**Status:** Week 1 COMPLETE ✅
**Readiness:** Moving to Week 2 with solid foundation
**Momentum:** High - excited for auth and concurrency challenges

