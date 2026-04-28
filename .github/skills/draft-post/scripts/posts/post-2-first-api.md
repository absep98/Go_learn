---
title: I Stopped Watching Tutorials and Started Building a REST API in Go
subtitle: Part 2 of Learning Go in Public
tags: go, buildinpublic, backend, beginners
series: Learning Go in Public
---

In my [last post](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii), I talked about my first week learning Go — the confusion, the messy first commit, and why I picked Go over sticking with JavaScript.

This post is about the moment things changed: I stopped writing practice files and started building a real project.

## The Decision

After a week of writing `structs.go`, `maps-practice.go`, and `interfaces-tutorial.go`, I was learning syntax but not learning to *build*. I could explain what a goroutine is but couldn't wire up an HTTP endpoint.

So I picked a project: a **personal analytics backend**. A REST API where I could log daily entries — mood, productivity notes, whatever — and query them later. Simple enough to be achievable, real enough to force me to learn.

## Day 1: The Skeleton

My first backend commit had just 3 files that mattered:

**`cmd/server/main.go`** — the entry point:

```go
func main() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found")
    }

    dbPath := os.Getenv("DB_PATH")
    if dbPath == "" {
        dbPath = "./data.db"
    }
    err = db.InitDB(dbPath)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.CloseDB()

    http.HandleFunc("/health", handlers.HealthHandler)
    http.HandleFunc("/ping", handlers.PingHandler)

    log.Printf("Server starting on port %s", port)
    http.ListenAndServe(":"+port, nil)
}
```

Two endpoints. `/health` returns "ok". `/ping` returns the current time. That's it. That was my entire backend.

But here's what I learned just from this:
- `http.HandleFunc` registers a function to a URL path — Go's router is that simple
- `http.ListenAndServe` blocks and listens forever — no framework needed
- `defer db.CloseDB()` — Go's `defer` runs when the function exits, perfect for cleanup

**`internal/db/db.go`** — database connection:

I used SQLite because it's a single file — no Docker, no Postgres setup, no connection strings. Just `./data.db`.

```go
func InitDB(path string) error {
    var err error
    DB, err = sql.Open("sqlite", path)
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }
    // Create tables
    _, err = DB.Exec(`CREATE TABLE IF NOT EXISTS entries (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        text TEXT NOT NULL,
        mood INTEGER CHECK(mood >= 1 AND mood <= 10),
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
    return err
}
```

The `%w` in `fmt.Errorf` wraps the original error — you can unwrap it later. I didn't understand why this mattered until weeks later when I needed to check error types.

## Day 2: The First Real Endpoint

This is where it clicked. I wrote `POST /entries` — my first handler that actually does something:

```go
func CreateEntry(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req CreateEntryRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        respondJSON(w, http.StatusBadRequest, CreateEntryResponse{
            Success: false,
            Message: "Invalid request body",
        })
        return
    }

    // Validate: user_id > 0, text not empty, mood 1-10
    if req.UserID <= 0 { /* return 400 */ }
    if req.Text == "" { /* return 400 */ }
    if req.Mood < 1 || req.Mood > 10 { /* return 400 */ }

    // Insert into database
    id, err := db.InsertEntry(req.UserID, req.Text, req.Mood)
    if err != nil {
        respondJSON(w, http.StatusInternalServerError, CreateEntryResponse{
            Success: false,
            Message: "Failed to create entry",
        })
        return
    }

    respondJSON(w, http.StatusCreated, CreateEntryResponse{
        Success: true,
        Message: "Entry created",
        ID:      id,
    })
}
```

Look at the pattern: **validate → process → respond. Early returns everywhere.** No nesting. No else blocks. If something's wrong, respond with an error and `return` immediately.

This is the Go way, and it's brilliant. Coming from JavaScript where I'd chain `.then().catch()` or wrap everything in try/catch, this felt refreshing. Every failure point is explicit.

## What I Learned About `http.ResponseWriter` vs `*http.Request`

This confused me for days. Why is `w` not a pointer but `r` is?

- `*http.Request` is a **pointer to a struct** — it's huge (headers, URL, cookies, body). Copying it would be wasteful.
- `http.ResponseWriter` is an **interface**. In Go, interfaces are already a small wrapper that points to the real data. You don't need a pointer to a pointer.

The rule I learned: *If you see a type without `*` but you can still call methods that modify things, it's probably an interface.*

## The Comments I Left for Myself

My code from this commit is packed with comments like:

```go
// "Read the JSON data from the request and put it into our req variable."
// Step-by-step:
// r.Body = the data the client sent
// json.NewDecoder(r.Body) = create a JSON reader
// .Decode(&req) = convert JSON into our struct
```

These comments are verbose and probably "wrong" by clean code standards. But they're exactly what I needed at the time — explanations in my own words, written while the concept was clicking.

I'm keeping them. They're documentation of learning, not documentation of code.

## The Moment It Felt Real

I ran `go run cmd/server/main.go`, opened another terminal, and typed:

```bash
curl -X POST http://localhost:8080/entries \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "text": "First entry!", "mood": 8}'
```

And got back:

```json
{"success": true, "message": "Entry created", "id": 1}
```

Data went into a real database. I could query it back with `GET /entries`. It was two endpoints and a SQLite file, but it was *mine*. I built it with zero frameworks, just Go's standard library.

That's when Go stopped being a language I was studying and became a tool I was building with.

## What's Next

In Part 3, I'll cover the scariest part of this project: **adding authentication from scratch**. JWT tokens, bcrypt password hashing, and the middleware pattern that changed how I think about HTTP — all without a single auth library.

*This is Part 2 of the "Learning Go in Public" series. [Part 1 is here.](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii)*
