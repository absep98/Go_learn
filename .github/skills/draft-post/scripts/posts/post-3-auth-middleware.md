---
title: Building Authentication From Scratch in Go — No Libraries, No Magic
subtitle: Part 3 of Learning Go in Public
tags: go, buildinpublic, authentication, backend
series: Learning Go in Public
---

In [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-51h5), I had a working REST API with two endpoints. You could create entries and query them. But anyone could hit the API — no login, no tokens, no protection.

This post is about the day I decided to build authentication from scratch, and the Go pattern that changed how I think about HTTP servers.

## The Plan

I needed three things:
1. **Register** — hash a password, save the user
2. **Login** — verify credentials, return a JWT token
3. **Middleware** — block unauthenticated requests before they reach any handler

No auth libraries. No Passport.js equivalent. Just `golang-jwt/jwt` for token creation and `golang.org/x/crypto/bcrypt` for password hashing.

## Registration: My First Time Hashing a Password

Here's what the register handler looked like:

```go
func Register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Validate
    if req.Email == "" { /* 400 */ }
    if !strings.Contains(req.Email, "@") { /* 400 */ }
    if len(req.Password) < 6 { /* 400 */ }

    // Hash password — NEVER store plain text
    passwordHash, err := bcrypt.GenerateFromPassword(
        []byte(req.Password), bcrypt.DefaultCost,
    )

    // Save to database
    userID, err := db.CreateUser(req.Email, string(passwordHash))
}
```

Two things I learned here:

**`bcrypt.DefaultCost` is 10.** That means it runs the hashing algorithm 2^10 = 1,024 times. Slow on purpose — makes brute-force attacks impractical. I spent 20 minutes reading about why a slower hash function is a *feature*, not a bug.

**`[]byte()` everywhere.** bcrypt's functions take byte slices, not strings. Coming from JavaScript where everything is just a string, I kept forgetting to convert. Go makes you think about data types at every step.

## Login: Creating My First JWT

The login flow taught me more about Go than any tutorial:

```go
// 1. Get user from database
userID, passwordHash, err := db.GetUserByEmail(req.Email)

// 2. Compare password with stored hash
err = bcrypt.CompareHashAndPassword(
    []byte(passwordHash), []byte(req.Password),
)

// 3. Create JWT claims
claims := jwt.MapClaims{
    "user_id": userID,
    "exp":     time.Now().Add(24 * time.Hour).Unix(),
}

// 4. Sign the token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString([]byte(secret))
```

I left this comment in my code at the time:

```go
// Why Unix()? JWT needs simple numbers, not complex Go time objects
// Example: time.Now().Unix() = 1736359530 (just a number)
// Add 24 hours = add 86400 seconds (24 * 60 * 60 = 86400)
```

Again — verbose, maybe obvious to experienced devs. But I was explaining JWT to *myself* while building it. That's the whole point of learning in public.

One security detail I got right: the error message for a wrong password is `"Invalid email or password"` — same as for a non-existent email. You never want to reveal whether an account exists.

## The Middleware Pattern That Changed Everything

This is the part I'm most proud of from this phase. Go doesn't have middleware built in like Express. You build it yourself:

```go
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Get token from "Authorization: Bearer <token>"
        authHeader := r.Header.Get("Authorization")
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // 2. Validate token
        claims, err := validateToken(tokenString)
        if err != nil {
            errorResponseAuth(w, http.StatusUnauthorized, "Invalid token")
            return
        }

        // 3. Extract user_id and add to request context
        userID := claims["user_id"].(float64)
        ctx := context.WithValue(r.Context(), "user_id", int64(userID))

        // 4. Call next handler with updated context
        next(w, r.WithContext(ctx))
    }
}
```

Read that again. **A function that takes a handler and returns a new handler.** The returned handler does the auth check *first*, and only calls the original handler if the token is valid.

In `main.go`, using it looks like this:

```go
// Public routes — no protection
http.HandleFunc("/register", handlers.Register)
http.HandleFunc("/login", handlers.Login)

// Protected route — wrapped with AuthMiddleware
http.HandleFunc("/entries", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        handlers.CreateEntry(w, r)
    } else if r.Method == http.MethodGet {
        handlers.GetEntries(w, r)
    }
}))
```

`AuthMiddleware(handler)` — that's it. One wrapper function. No framework, no plugin system, no decorator syntax. Just a function wrapping another function.

This is the moment Go clicked for me at a deeper level. Functions are values. You can pass them around, wrap them, chain them. Middleware isn't a framework feature — it's just *how functions work*.

## The `context.WithValue` Trick

The most clever part of Go's auth pattern: how you pass the user ID from middleware to the handler.

```go
// Middleware sets it:
ctx := context.WithValue(r.Context(), "user_id", int64(userID))
next(w, r.WithContext(ctx))

// Handler reads it:
userID := r.Context().Value("user_id").(int64)
```

No global variables. No session objects. The user ID travels *with the request itself*. Every handler that needs to know who's making the request just reads it from the context.

Coming from Express where you'd do `req.user = decoded`, this felt more explicit and safer. The data is scoped to the request, not bolted onto a mutable object.

## One Thing That Tripped Me Up

JWT numbers come back as `float64`, not `int64`. This line:

```go
userID, ok := claims["user_id"].(float64)
```

I originally wrote `.(int64)` and couldn't figure out why it kept failing. Turns out JSON (which JWT uses internally) doesn't distinguish between integers and floats — everything is a number. Go's JSON decoder maps all numbers to `float64`.

Took me an hour of debugging. Now I'll never forget it.

## What the API Looked Like After This

```
POST /register  → Create account (public)
POST /login     → Get JWT token (public)
POST /entries   → Create entry (protected)
GET  /entries   → List entries (protected)
GET  /health    → Health check (public)
GET  /ping      → Ping (public)
```

Four real endpoints. Authentication. Authorization. No frameworks. Just Go's standard library plus two packages for JWT and bcrypt.

## What's Next

In Part 4, I'll cover the week I added **testing, UPDATE, and DELETE** — and why writing tests *after* the code changed how I thought about designing handlers.

*This is Part 3 of the "Learning Go in Public" series. [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-51h5)*
