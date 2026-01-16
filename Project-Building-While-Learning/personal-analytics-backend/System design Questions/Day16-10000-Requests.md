# Day 16: What Breaks at 10,000 Concurrent Requests?

**Question:** What happens if 10,000 users hit `GET /entries?page=1` simultaneously?

---

## 🔴 Problem 1: Deep Paging (OFFSET/LIMIT)

### What It Is

When someone requests page 1000 with `LIMIT 10 OFFSET 10000`, the database doesn't jump to row 10,001. It **scans and discards** the first 10,000 rows to find the starting point.

### Library Analogy 📚
>
> You ask: "Give me the 100th book on the shelf."
>
> Librarian counts: 1, 2, 3... 99, 100. Here you go!
>
> Now 10,000 people ask for the **millionth book**.
> The librarian must count to one million **for each person** - even though they only want 1 book.

### What Breaks

- CPU and I/O spike dramatically
- Each query gets **linearly slower** as offset increases
- Page 1 = fast, Page 1000 = slow, Page 10000 = database locks up

### The Fix: Keyset Pagination (Cursor-based)

```sql
-- ❌ Slow (OFFSET)
SELECT * FROM entries ORDER BY id LIMIT 10 OFFSET 10000

-- ✅ Fast (Cursor)
SELECT * FROM entries WHERE id > 500 ORDER BY id LIMIT 10
```

**Why it's faster:** "Start from ID 500" is instant vs "Skip 500 rows"

---

## 🔴 Problem 2: Thundering Herd (Cache Stampede)

### What It Is

You cache `COUNT(*)` to avoid recalculating on every request. But when the cache **expires**, all waiting requests hit the database at once.

### Library Analogy 📚
>
> The library has a "Total Books: 50,000" sign on the wall.
>
> Every hour, the sign is erased to be updated.
>
> At that moment, 10,000 people see the blank sign.
> Instead of waiting, they ALL jump over the counter and start counting books themselves.
>
> Chaos. Nobody gets their answer.

### What Breaks

- Cache expires → 10,000 requests all try to recalculate
- Database gets hammered with identical expensive queries
- System becomes unresponsive

### The Fix: `singleflight` Package

```go
import "golang.org/x/sync/singleflight"

var group singleflight.Group

// Only ONE request actually counts, others wait for the result
result, _, _ := group.Do("count-entries", func() (interface{}, error) {
    return db.CountEntries(userID)
})
```

**What it does:** "Wait! One person is already counting. I'll tell everyone the answer when they're done."

---

## 🔴 Problem 3: Database Concurrency (SQLite vs Postgres)

### Traffic Analogy 🚗

| Database | Road Type | Concurrent Writes |
|----------|-----------|-------------------|
| SQLite | Single-lane road | 1 at a time |
| PostgreSQL | 16-lane highway | Many at once |
| MySQL | 16-lane highway | Many at once |

### What Breaks with SQLite

- Only ONE write can happen at a time
- 10,000 write requests = massive queue
- "Database is locked" errors

### The Fix

- **Development:** SQLite is fine
- **Production with traffic:** Switch to PostgreSQL or MySQL

---

## 🔴 Problem 4: Connection Exhaustion

### Phone Line Analogy 📞
>
> Every time you call the library, you must:
>
> 1. Install a new phone line
> 2. Ask one question
> 3. Rip out the phone line
>
> 10,000 calls/second? You'll run out of wire!

### What Breaks

- **File Descriptors:** Linux defaults to 1,024 open files → "too many open files" error
- **Ephemeral Ports:** Server runs out of ports to connect to DB
- Each new connection has overhead (TCP handshake, auth, etc.)

### The Fix: Connection Pooling

```go
// Go's database/sql already does this!
db, _ := sql.Open("sqlite", "data.db")
db.SetMaxOpenConns(25)      // Max 25 connections
db.SetMaxIdleConns(5)       // Keep 5 ready
db.SetConnMaxLifetime(5*time.Minute)
```

**What it does:** Keep 25 "phone lines" always connected. Requests share them.

---

## 🔵 Deep Dive: Connection Pooling Explained

Think of the database as a **high-security office building** and the "Connection" as a **security badge** required to enter.

### ❌ Without Pooling (Expensive & Slow)

Every time a user makes a request:

```
1. 🤝 Handshake     → Your Go server calls the database
2. 🔐 Verification  → DB checks username/password, allocates memory
3. 📝 Work          → Run your query
4. 🗑️ Cleanup       → Close connection, DB tears down resources
```

**The Problem:**
If 10,000 people arrive at once, the database must:

- Check 10,000 IDs
- Print 10,000 badges
- Set up 10,000 desks

**This "setup" often takes longer than the actual query!** (20ms-100ms just to connect)

### ✅ With Pooling (Fast & Reusable)

Your Go server acts like a **manager who already checked out 50 badges in advance**.

```
1. 🏁 Initialization → App starts, creates pool of 50 open connections
2. 🎫 Borrowing      → Request comes in, borrows an existing badge
3. ⚡ Using          → Query runs INSTANTLY (no handshake needed!)
4. ↩️ Returning      → Request returns badge to pool (NOT destroyed)
```

### Security Badge Analogy 🎫

| Scenario | Without Pool | With Pool |
|----------|--------------|-----------|
| New Request | "Print a new badge, verify ID, set up desk" | "Grab badge #23 from the drawer" |
| Login/Auth | Happens **every single request** | Happens **once when app starts** |
| After Query | Badge destroyed, desk torn down | Badge returned to drawer |
| 10,000 Requests | DB crashes (too many logins) | Requests wait briefly (stable) |

### Why This Solves 10,000 Requests

| Benefit | Explanation |
|---------|-------------|
| **Speed** | Skip 20-100ms connection setup per request |
| **Stability** | 10,000 requests share 50 connections instead of creating 10,000 |
| **Resource Control** | `MaxOpenConnections` prevents killing the database |

### Go Connection Pool Settings

```go
import (
    "database/sql"
    "time"
)

func InitDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite", "data.db")
    if err != nil {
        return nil, err
    }

    // Pool Configuration
    db.SetMaxOpenConns(25)              // Max connections open at once
    db.SetMaxIdleConns(5)               // Keep 5 connections ready (warm)
    db.SetConnMaxLifetime(5 * time.Minute)  // Recycle connections every 5 min
    db.SetConnMaxIdleTime(1 * time.Minute)  // Close idle connections after 1 min

    return db, nil
}
```

### Settings Explained

| Setting | What It Does | When To Tune |
|---------|--------------|--------------|
| `MaxOpenConns` | Max simultaneous connections | Increase for high traffic |
| `MaxIdleConns` | Pre-warmed connections ready to use | Higher = faster response, more memory |
| `ConnMaxLifetime` | Force reconnect after X time | Helps with DB failover |
| `ConnMaxIdleTime` | Close unused connections | Saves resources when quiet |

### 🔑 Key Insight

> **In Go, `sql.Open()` doesn't open a connection - it opens a POOL.**
>
> You're not "connecting to a database" - you're "creating a connection manager."

---
### 🔬 Going Deeper: What Exactly Gets Skipped?

It's not just about skipping authentication - it's about skipping the **entire handshake process**.

#### The 4-Step Setup Cost (Without Pooling)

Every single request must do ALL of these before running a query:

```
Step 1: 🌐 Network Handshake
        → Computer says "Hello" to database (opens a socket)
        → Takes: 1-5ms

Step 2: 🔒 TLS/SSL Handshake  
        → Agree on encryption so nobody can spy
        → Takes: 10-30ms

Step 3: 🔐 Authentication
        → Database checks username/password
        → Takes: 5-20ms

Step 4: 🧵 Resource Allocation
        → Database creates a dedicated worker (thread/process) for you
        → Takes: 5-10ms

        ─────────────────────────
        TOTAL SETUP: 20-65ms
        ACTUAL QUERY: 5-10ms
```

**With Pooling:** All 4 steps happen **ONCE** when the app starts. New users skip straight to the query.

---

### ⚡ The Two Benefits of Pooling

#### Benefit 1: Speed (Skip the Handshake)

| Scenario | Handshake | Query | **Total** |
|----------|-----------|-------|-----------|
| No Pool | 50ms | 10ms | **60ms** |
| With Pool | 0ms | 10ms | **10ms** |

**Result:** Your app feels **6x faster**

---

#### Benefit 2: Stability (Traffic Control)

| Scenario | What Happens | Result |
|----------|--------------|--------|
| No Pool | 10,000 users → 10,000 workers on DB | DB runs out of RAM → **CRASH** |
| With Pool | 10,000 users → share 100 connections | Users take turns → **STABLE** |

**Result:** Your app **doesn't break** under pressure

---

### 🌉 Bridge Analogy

```
❌ No Pooling:
   → Build a new bridge
   → Cross it once  
   → Blow up the bridge
   → Repeat 10,000 times
   
✅ With Pooling:
   → Build 50 permanent bridges
   → Everyone takes turns crossing them
   → Bridges stay forever
```

---

### 💡 Extra Detail: Multi-User Pools

If you have different types of users (Admin vs Customer), the pooler is smart enough to keep **separate bridges** for each so they don't accidentally use the wrong credentials.

```go
// In practice, you typically have one pool per database
// But the pool handles connection credentials correctly
adminDB := sql.Open("postgres", "user=admin ...")
customerDB := sql.Open("postgres", "user=readonly ...")
```

---

### 🔐 Wait... Isn't Skipping Auth Dangerous?

**Great question!** You might wonder: "If we skip authentication, can a hacker just connect?"

**Short answer:** The 10,000 users are NOT logging into the database. They're logging into your **Go application**.

---

#### The Two Layers of Security (Bank Analogy 🏦)

Think of your system like a bank:

```
┌─────────────────────────────────────────────────────────┐
│                     THE INTERNET                        │
│                                                         │
│   👤 User/Hacker                                        │
│      ↓                                                  │
│   ┌─────────────────────────────────────────┐          │
│   │     🏦 YOUR GO SERVER (The Bank Lobby)   │          │
│   │                                          │          │
│   │   "Show me your ID card (JWT/Password)" │          │
│   │                                          │          │
│   │   ❌ No ID? → REJECTED (never sees DB)  │          │
│   │   ✅ Valid ID? → Proceed                │          │
│   └──────────────────┬──────────────────────┘          │
│                      ↓                                  │
│   ┌─────────────────────────────────────────┐          │
│   │      🗄️ DATABASE (The Vault)            │          │
│   │                                          │          │
│   │   Only trusts the Go Server              │          │
│   │   (Has "Master Key" = pooled connection)│          │
│   └─────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────┘
```

---

#### Two Different Authentications

| Layer | Who Authenticates | How | When |
|-------|-------------------|-----|------|
| **User → Go App** | Each user | JWT / Password / Cookie | Every request |
| **Go App → Database** | The Go server | Connection string (pooled) | Once at startup |

---

#### What Happens When a Hacker Attacks?

```
1. 🦹 Hacker sends request to your server
2. 🔍 Go server asks: "Where is your JWT token?"
3. ❌ No valid token → REJECTED
4. 🚫 Request NEVER touches the database pool
```

**The "skipped authentication" only happens between your trusted Go server and the database.**

---

#### Why One "Master Account" is Safe

| Scenario | What Happens |
|----------|--------------|
| Without Pool | Go App says to DB: "Hi, it's me again. Here's my password. Let me in." (Every request) |
| With Pool | Go App says once: "I'll leave 50 phone lines open. I'll shout queries whenever customers ask." |

**The database doesn't know "User #123" exists.** It only knows "The Go App" is asking for data.

---

#### Security Roles Breakdown

| Entity | Role | Responsibility |
|--------|------|----------------|
| **User / Hacker** | The "Client" | Must provide Password/JWT to Go App |
| **Go Application** | The "Gatekeeper" | Verifies user, then borrows pool connection |
| **Database** | The "Warehouse" | Only trusts Go App (not end users) |

---

#### ⚠️ The REAL Security Risk: SQL Injection

Since your Go App has a "Master Key" to the database, the danger isn't the connection—it's **what you tell the database to do**.

**The Attack:**
```go
// ❌ DANGEROUS - User input directly in query
userInput := "1; DROP TABLE users; --"
query := "SELECT * FROM books WHERE id = " + userInput
db.Query(query)  // Deletes your users table!
```

**The Fix:**
```go
// ✅ SAFE - Parameterized query
userInput := "1; DROP TABLE users; --"
db.Query("SELECT * FROM books WHERE id = ?", userInput)
// Database treats entire input as a single value, not SQL code
```

---

#### 🎯 Summary

> **Pooling skips the login between Your Server ↔ Database**
> 
> **But your server still checks every user's ID before letting them use that "open door"**

```
User → [JWT Check] → Go Server → [Pool] → Database
        ↑                           ↑
    ALWAYS checked              ALREADY open
```

---

## 🔴 Problem 5: Memory Pressure (Go-Specific)

### What It Is

10,000 goroutines each allocating slices to hold DB results can cause:

- Out of Memory (OOM) errors
- Garbage Collection pauses (everything freezes briefly)

### The Fix

1. **Select only needed columns** (not `SELECT *`)
2. **Use `sync.Pool`** to reuse objects instead of creating new ones
3. **Stream results** instead of loading all into memory

---

## 📊 Quick Comparison: Pagination Strategies

| Feature | Offset/Limit | Keyset (Cursor) |
|---------|--------------|-----------------|
| **Performance** | ❌ Slows down on deep pages | ✅ Consistent speed |
| **Simplicity** | ✅ Easy to implement | ⚠️ Needs sorted unique column |
| **Deep Paging** | ❌ Expensive | ✅ Cheap |
| **Data Drift** | ⚠️ Items can be skipped/duplicated | ✅ No skipping |

---

## 🎯 Summary Table

| Problem | What Breaks | Fix |
|---------|-------------|-----|
| Deep Paging | DB scans millions of rows | Cursor pagination (`WHERE id > X`) |
| Thundering Herd | Cache expires → DB hammered | `singleflight` package |
| SQLite Limits | Only 1 write at a time | Use PostgreSQL in production |
| Connection Exhaustion | "Too many open files" | Connection pooling |
| Memory Pressure | OOM / GC pauses | `sync.Pool`, stream results |

---

## 💡 Interview-Ready Answer

> "At 10,000 concurrent requests, the main bottlenecks are:
>
> 1. **Deep pagination** - OFFSET scans grow linearly, so I'd switch to cursor-based pagination
> 2. **Cache stampedes** - When cache expires, all requests hit DB. I'd use Go's `singleflight` package
> 3. **Database choice** - SQLite can't handle concurrent writes, so I'd use PostgreSQL
> 4. **Connection limits** - I'd configure connection pooling and increase system file descriptor limits
>
> In my project, I currently use OFFSET pagination with SQLite, which works for development. For production, I'd implement cursor pagination and switch to PostgreSQL."

---

## 🔗 Relates to My Project

**Current state:**

- Using OFFSET/LIMIT pagination ✅ (works for small scale)
- Using SQLite ✅ (fine for development)
- No caching yet ❌ (Day 17 topic)

**Would need for production:**

- Cursor pagination for deep pages
- PostgreSQL for concurrent writes
- Cache with stampede protection
- Connection pool configuration
