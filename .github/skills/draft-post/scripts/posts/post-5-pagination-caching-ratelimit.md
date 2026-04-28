---
title: What Happens When Your API Has 10,000 Rows? I Added Pagination and Caching to Find Out
subtitle: Part 5 of Learning Go in Public
tags: go, buildinpublic, backend, webdev
series: Learning Go in Public
---

In [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o), I finished full CRUD and wrote my first Go tests. The API worked — but I was the only user, hitting it from my terminal with 10 test entries.

Then I asked: *what happens when there are 10,000 rows?*

Two days, two features. Both taught me things I didn't expect.

## Day 16: Pagination

`GET /entries` was returning every row in the database for a user. Fine when I have 10 test entries. A disaster with 10,000.

Pagination is simple in principle — you add `?page=1&limit=10` to the URL and only fetch that slice. The implementation taught me something about Go's "fail gracefully" philosophy:

```go
page := 1   // default
limit := 10 // default

pageStr := r.URL.Query().Get("page")
if pageStr != "" {
    if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
        page = p
    }
}

limitStr := r.URL.Query().Get("limit")
if limitStr != "" {
    if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
        limit = l
    }
}
```

Notice what's NOT there: no `400 Bad Request` if you pass `?page=abc`. Invalid params just silently fall back to defaults. I went back and forth on this — should I return an error?

Decided no. Pagination params are optional hints, not required inputs. If you pass garbage, you get the first page. The API keeps working. This feels more correct for a read endpoint.

The database query uses `LIMIT` and `OFFSET`:

```sql
SELECT * FROM entries 
WHERE user_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3
```

Where `OFFSET = (page - 1) * limit`. The response now includes `total`, `page`, `limit`, and `entries` — so clients can calculate how many pages exist.

## Day 17: In-Memory Caching

Pagination meant a COUNT query on every GET request to return the total. `SELECT COUNT(*) FROM entries WHERE user_id = ?` is cheap with 10 rows, expensive at scale.

I built a simple in-memory cache:

```go
type CacheEntry struct {
    value      interface{}
    Expiration time.Time
}

type Cache struct {
    data map[string]CacheEntry
    mu   sync.RWMutex  // ← this is the key part
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, exists := c.data[key]
    if !exists || time.Now().After(entry.Expiration) {
        return nil, false
    }
    return entry.value, true
}
```

The `sync.RWMutex` was new to me. Regular `Mutex` blocks everything — only one goroutine at a time. `RWMutex` is smarter: multiple goroutines can *read* simultaneously, but *writing* requires exclusive access.

For a cache that's read far more often than written, this matters.

Cache invalidation: when a user creates or deletes an entry, I delete their cached count:

```go
cache.Delete(fmt.Sprintf("count:user:%d", userID))
```

The cache key includes the user ID so different users never see each other's counts. Simple, but it took me a few minutes to think through why I needed that.

## What I Learned

**Pagination defaults beat errors.** For optional hints like `page` and `limit`, failing gracefully is better UX than returning a 400. Save strict validation for required inputs.

**`sync.RWMutex` vs `sync.Mutex`.** Read-heavy data structures should use `RWMutex`. Multiple goroutines can read simultaneously — only writes need exclusive access.

**Cache keys need scope.** `count:user:42` not just `count`. Otherwise all users share the same cached value — a subtle bug that only shows up with multiple users.

## The API Now

```
GET /entries?page=1&limit=10   → paginated, cached COUNT
POST /entries                  → creates entry, invalidates cache
DELETE /entries?id=X           → deletes entry, invalidates cache
```

The API no longer dumps every row on every request. The database isn't hit for the count on repeated reads. Small changes, but the kind that matter when traffic is real.

## What's Next

In Part 6, I'll cover **rate limiting and Redis** — the week I started thinking about abuse prevention and why in-memory solutions break the moment you run more than one server instance.

*This is Part 5 of the "Learning Go in Public" series. [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o)*
