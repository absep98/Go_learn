---
title: Tests, UPDATE, DELETE — And the Refactor I Didn't Plan
subtitle: Part 4 of Learning Go in Public
tags: go, buildinpublic, testing, backend
series: Learning Go in Public
---

In [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46), I built authentication — JWT tokens, bcrypt hashing, middleware that wraps handlers like Russian dolls. The API had four endpoints and was fully protected.

But it could only create and read entries. No updating. No deleting. And zero tests.

This post is about the week I added the missing CRUD operations, wrote my first Go tests, and accidentally refactored the entire project structure — all because one test file exposed a design flaw I couldn't unsee.

## UPDATE: Harder Than I Expected

Adding a PATCH endpoint sounded simple. It wasn't. The tricky part wasn't the database query — it was **ownership validation**.

When a user sends `PATCH /entries?id=5`, I can't just update entry 5. I need to verify that entry 5 *belongs to this user*. Otherwise any authenticated user could edit anyone's entries.

Here's the handler:

```go
func UpdateEntry(w http.ResponseWriter, r *http.Request) {
    // Parse entry ID from query string
    entryId, err := strconv.Atoi(r.URL.Query().Get("id"))
    
    // Get user ID from context (set by AuthMiddleware)
    userID := r.Context().Value("user_id").(int64)
    
    // Validate all fields...
    if req.Text == "" { /* 400 */ }
    if req.Mood < 1 || req.Mood > 10 { /* 400 */ }
    
    // The key line — WHERE clause checks BOTH id AND user_id
    rowsAffected, err := db.UpdateEntry(ctx, entryId, userID, 
        req.Text, req.Mood, req.Category)
    
    if rowsAffected == 0 {
        // Entry doesn't exist OR doesn't belong to this user
        errorResponse(w, http.StatusNotFound, 
            "Entry not found or access denied")
        return
    }
}
```

The database query uses `WHERE id = ? AND user_id = ?`. If the entry exists but belongs to someone else, `rowsAffected` is 0. Same response as "entry doesn't exist." This is intentional — you don't want to leak information about other users' data.

I wrote an 8-test PowerShell script to verify every edge case: update your own entry, try to update someone else's, update with missing fields, update with mood out of range. All passing.

## DELETE: The Same Pattern, Simpler

Once UPDATE worked, DELETE took 20 minutes. Same ownership pattern:

```go
func DeleteEntry(w http.ResponseWriter, r *http.Request) {
    entryId, _ := strconv.Atoi(r.URL.Query().Get("id"))
    userID := r.Context().Value("user_id").(int64)
    
    rowsAffected, err := db.DeleteEntry(ctx, entryId, userID)
    
    if rowsAffected == 0 {
        errorResponse(w, http.StatusNotFound, 
            "Entry not found or access denied")
        return
    }
    
    respondJSON(w, http.StatusOK, CreateEntryResponse{
        Success: true,
        Message: "Entry deleted successfully",
    })
}
```

Same `WHERE id = ? AND user_id = ?`. Same ownership check. Same ambiguous error message. I didn't realize it at the time, but this was my first Go **pattern** — I was copying the structure from UPDATE and it just worked.

Seven tests. All green.

## The Accidental Refactor

Here's where things got interesting. I was writing tests when I noticed something ugly: my `models/entry.go` had database functions mixed in with struct definitions. The `Entry` struct and `InsertEntry()` were in the same file.

```
models/
  entry.go    ← struct definitions AND database functions??
  models.go   ← more structs
```

This was wrong. The whole point of separating `models/` from `db/` is separation of concerns. Models define *what* data looks like. DB defines *how* it's stored and retrieved.

So I refactored:

```
models/
  models.go   ← ALL struct definitions (Entry, User, requests, responses)
db/
  db.go       ← ALL database operations (Insert, Update, Delete, Get)
```

I moved every database function from `models/entry.go` into `db/db.go`. Deleted the old file. Updated all imports.

This wasn't planned. It happened because writing tests forced me to think about where things belong. When you import a package for testing, messy boundaries become obvious.

## My First Real Tests

I had been testing with PowerShell scripts — `curl`-style commands that hit a running server. Those are integration tests. Useful, but slow and fragile.

Go has `testing` and `httptest` built into the standard library. No Jest, no Mocha, no test runner to install. You just:

1. Create a file ending in `_test.go`
2. Write functions starting with `Test`
3. Run `go test`

Here's my first handler test:

```go
func TestCreateEntry(t *testing.T) {
    body := strings.NewReader(
        `{"text":"feeling good","mood":8,"category":"mood"}`,
    )
    
    // Create a fake HTTP request
    req := httptest.NewRequest(http.MethodPost, "/entries", body)
    
    // Inject user_id into context (normally AuthMiddleware does this)
    ctx := context.WithValue(req.Context(), "user_id", int64(1))
    req = req.WithContext(ctx)
    
    // Record the response
    rec := httptest.NewRecorder()
    
    // Call the handler directly — no server needed
    CreateEntry(rec, req)
    
    // Assert
    if rec.Code != http.StatusCreated {
        t.Errorf("expected 201, got %d", rec.Code)
    }
}
```

Two things blew my mind:

**`httptest.NewRequest`** — you create a fake HTTP request with any method, URL, and body. No actual network call. No server running. It's just a struct.

**`httptest.NewRecorder`** — captures whatever the handler writes. Status code, headers, body. You inspect it after the handler returns.

The test for invalid input was almost as satisfying:

```go
func TestCreateEntry_MissingText(t *testing.T) {
    body := strings.NewReader(`{"text":"","mood":8,"category":"mood"}`)
    req := httptest.NewRequest(http.MethodPost, "/entries", body)
    ctx := context.WithValue(req.Context(), "user_id", int64(1))
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()
    
    CreateEntry(rec, req)
    
    if rec.Code != http.StatusBadRequest {
        t.Errorf("expected 400, got %d", rec.Code)
    }
}
```

Empty text → 400. Invalid mood → 400. Missing auth → 401. Each scenario is three lines of setup, one function call, one assertion.

Coming from JavaScript testing where you install Jest, configure Babel, mock fetch, set up test databases... Go's approach felt almost too simple. But it works.

## What I Learned This Week

**The ownership pattern is reusable.** `WHERE id = ? AND user_id = ?` with `rowsAffected == 0` handles both "not found" and "not yours" in one query. I've used this pattern three times now (GET, UPDATE, DELETE) and it's always the same.

**Refactoring happens when you test.** I didn't plan to restructure the project. But the moment I tried to write a test that imported `models`, the messy boundaries became obvious. Tests are a design feedback tool, not just a verification tool.

**Go testing is minimal and that's the point.** No assertions library. No mocking framework. Just `if got != want { t.Errorf() }`. It forces you to write tests that are readable without any DSL knowledge.

## Where the API Stands Now

```
POST   /register       → Create account (public)
POST   /login          → Get JWT token (public)
POST   /entries        → Create entry (protected)
GET    /entries        → List entries (protected)
PATCH  /entries?id=X   → Update entry (protected + ownership)
DELETE /entries?id=X   → Delete entry (protected + ownership)
GET    /health         → Health check (public)
```

Six real endpoints. Full CRUD. Authentication. Authorization with ownership validation. 18 tests passing. And a project structure that finally makes sense.

## What's Next

In Part 5, I'll cover the week I added **pagination, caching, and rate limiting** — the features that made me think about what happens when this API isn't just me hitting it from PowerShell anymore.

*This is Part 4 of the "Learning Go in Public" series. [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46)*
