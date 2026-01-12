# Development Guide

**Quick reference for common patterns, functions, and Go idioms used in this project.**

---

## 🔥 THE 5 CORE PATTERNS

### Pattern 1: Create a Handler

Every HTTP handler follows this structure:

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Check HTTP method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 2. Parse JSON request
    var req MyRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        errorResponse(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // 3. Validate input
    if req.Field == "" {
        errorResponse(w, http.StatusBadRequest, "field is required")
        return
    }

    // 4. Database operation
    result, err := db.MyDBFunction(req.Field)
    if err != nil {
        errorResponse(w, http.StatusInternalServerError, "Database error")
        return
    }

    // 5. Success response
    respondJSON(w, http.StatusOK, MyResponse{Success: true, Data: result})
}
```

**Key Points:**

- Sequential validation with early returns
- Consistent error responses
- Clear status codes
- Log important operations

---

### Pattern 2: Database Query (SQL Injection Safe)

```go
func MyDBFunction(param string) (result, error) {
    // ALWAYS use ? placeholders (parameterized queries)
    query := `SELECT * FROM table WHERE field = ?`

    rows, err := DB.Query(query, param)  // param goes here, NOT in query string
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Process rows...
    return result, nil
}
```

**CRITICAL Security Rules:**

- ✅ Use `?` placeholders
- ✅ Pass values as separate arguments
- ❌ NEVER concatenate strings into SQL: `"WHERE id = " + userInput` (vulnerable!)

---

### Pattern 3: Hash Password (Registration)

```go
// Hash password before storing
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    log.Printf("Error hashing password: %v", err)
    return err
}

// Save string(hash) to database (never save plain password!)
userID, err := db.CreateUser(email, string(hash))
```

**Why bcrypt?**

- One-way hashing (can't reverse)
- Automatically salted
- Computationally expensive (slows brute force)
- Industry standard

---

### Pattern 4: Verify Password (Login)

```go
// Get hash from database
userID, passwordHash, err := db.GetUserByEmail(email)
if err != nil {
    // User doesn't exist - don't reveal this info!
    errorResponse(w, http.StatusUnauthorized, "Invalid email or password")
    return
}

// Compare submitted password with stored hash
err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
if err != nil {
    // Wrong password - same error message as above (security!)
    errorResponse(w, http.StatusUnauthorized, "Invalid email or password")
    return
}

// Success - generate JWT token
```

**Security Note:** Same error message for "user not found" and "wrong password" prevents enumeration attacks.

---

### Pattern 5: Generate JWT Token (Login)

```go
// Step 1: Create claims (data inside token)
claims := jwt.MapClaims{
    "user_id": userID,
    "exp": time.Now().Add(24 * time.Hour).Unix(),  // Expires in 24 hours
}

// Step 2: Create token object with signing method
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// Step 3: Sign token with secret key
secret := os.Getenv("JWT_SECRET")
tokenString, err := token.SignedString([]byte(secret))
if err != nil {
    log.Printf("Error generating token: %v", err)
    return "", err
}

// Return tokenString to user
return tokenString, nil
```

**JWT Structure:** `header.payload.signature`

- Header: Algorithm (HS256)
- Payload: Claims (user_id, expiration)
- Signature: Proves token hasn't been tampered with

---

## 📦 Handler Functions Reference

### `respondJSON(w, status, data)`

**Purpose:** Send any data as JSON with HTTP status code

**Usage:**

```go
respondJSON(w, http.StatusOK, map[string]interface{}{
    "success": true,
    "count": 5,
    "data": results,
})
```

**When to use:** All success responses, structured data

---

### `errorResponse(w, status, message)`

**Purpose:** Send standardized error message

**Usage:**

```go
errorResponse(w, http.StatusBadRequest, "text cannot be empty")
```

**When to use:** All error scenarios (validation, auth, database errors)

**Pattern:** Use `errorResponse` for errors, `respondJSON` for everything else

---

## 🗄️ Database Functions Reference

### `db.InitDB(dbPath)`

**Purpose:** Initialize database connection and create tables
**When:** Once at server startup in main.go
**Returns:** `error` if connection fails

```go
err := db.InitDB("./data.db")
if err != nil {
    log.Fatalf("Failed to initialize database: %v", err)
}
defer db.CloseDB()  // Close on shutdown
```

---

### `db.CreateUser(email, passwordHash)`

**Purpose:** Insert new user into database
**Returns:** `(userID int64, error)`

```go
userID, err := db.CreateUser("user@example.com", hashedPassword)
if err != nil {
    // Check for duplicate email
    if strings.Contains(err.Error(), "UNIQUE constraint failed") {
        errorResponse(w, http.StatusConflict, "Email already registered")
        return
    }
    errorResponse(w, http.StatusInternalServerError, "Failed to create user")
    return
}
```

---

### `db.GetUserByEmail(email)`

**Purpose:** Retrieve user for login verification
**Returns:** `(userID int64, passwordHash string, error)`

```go
userID, hash, err := db.GetUserByEmail(email)
if err != nil {
    errorResponse(w, http.StatusUnauthorized, "Invalid email or password")
    return
}
```

---

### `db.InsertEntry(userID, text, mood, category)`

**Purpose:** Save new entry to database
**Returns:** `(entryID int64, error)`

```go
id, err := db.InsertEntry(int(userID), "Great day", 8, "work")
if err != nil {
    errorResponse(w, http.StatusInternalServerError, "Failed to save entry")
    return
}
```

---

### `db.GetEntriesByUser(userID)`

**Purpose:** Retrieve all entries for specific user
**Returns:** `([]map[string]interface{}, error)`

```go
entries, err := db.GetEntriesByUser(userID)
if err != nil {
    errorResponse(w, http.StatusInternalServerError, "Failed to retrieve entries")
    return
}
```

---

## 🔒 HTTP Status Codes Quick Reference

| Code | Name | When to Use |
|------|------|-------------|
| 200 | OK | Successful GET request |
| 201 | Created | POST successfully created resource |
| 400 | Bad Request | Validation failed, invalid input |
| 401 | Unauthorized | Missing/invalid authentication |
| 404 | Not Found | Resource doesn't exist |
| 405 | Method Not Allowed | Wrong HTTP method (GET instead of POST) |
| 409 | Conflict | Resource already exists (duplicate email) |
| 500 | Internal Server Error | Database error, server crash |

**Decision Tree:**

- Client sent bad data → 400
- Client not authenticated → 401
- Resource not found → 404
- Server/database error → 500
- Success (read) → 200
- Success (created) → 201

---

## 🛡️ Middleware Patterns

### Protect a Route

```go
http.HandleFunc("/entries", handlers.AuthMiddleware(handlers.MyHandler))
//                           ↑ Wraps handler with auth check
```

### Chain Multiple Middleware

```go
http.HandleFunc("/entries",
    handlers.LoggingMiddleware(
        handlers.AuthMiddleware(handlers.MyHandler)))
// Executes: Logging → Auth → Handler
```

### Get User ID in Protected Handler

```go
// Middleware puts user_id in context
userIDValue := r.Context().Value("user_id")
if userIDValue == nil {
    errorResponse(w, http.StatusUnauthorized, "User not authenticated")
    return
}

userID, ok := userIDValue.(int64)
if !ok {
    errorResponse(w, http.StatusInternalServerError, "Invalid authentication data")
    return
}

// Now use userID
log.Printf("Request from user: %d", userID)
```

---

## 📋 Common Go Patterns

### Multiple Return Values

**Every database/external call returns `(result, error)`:**

```go
result, err := someFunction()
if err != nil {
    // Handle error first
    log.Printf("Error: %v", err)
    return
}
// Use result
```

**Pattern:** Always check error before using result

---

### Defer for Cleanup

```go
file, err := os.Open("data.txt")
if err != nil {
    return err
}
defer file.Close()  // Runs when function exits, even if error occurs

// Use file...
// file.Close() will be called automatically
```

**Common uses:**

- `defer db.CloseDB()`
- `defer rows.Close()`
- `defer file.Close()`

---

### JSON Struct Tags

```go
type User struct {
    ID    int    `json:"user_id"`    // JSON "user_id" → Go field ID
    Name  string `json:"name"`       // Exact match required
    Email string `json:"email,omitempty"`  // Omit if empty
}
```

**Rules:**

- Tag value must EXACTLY match JSON key
- Case-sensitive matching
- `omitempty` excludes zero values from JSON output
- No tag = Go field name must match JSON key exactly

---

### Type Conversions

**Go doesn't auto-convert even between int types:**

```go
// Common conversions
var i64 int64 = 100
var i int = int(i64)       // int64 → int

var str string = "hello"
var bytes []byte = []byte(str)  // string → []byte
var backToStr = string(bytes)   // []byte → string

// JWT claims come as float64
userID, ok := claims["user_id"].(float64)  // Type assertion
if ok {
    userIDInt := int64(userID)  // Convert to int64
}
```

**Why `[]byte()`?** Many libraries (bcrypt, JWT signing) need bytes, not strings.

---

## 🏗️ Request Flow (Memorize This!)

```
1. User sends HTTP request
2. Server receives at main.go route
3. LoggingMiddleware logs the request
4. AuthMiddleware checks JWT token (if protected route)
5. Handler function executes:
   - Validates HTTP method
   - Parses JSON body
   - Validates input
   - Calls database function
   - Returns JSON response
6. Response sent back to user
```

---

## 🔑 File Responsibilities

| File | Purpose |
|------|---------|
| main.go | Route registration, server startup |
| middleware.go | JWT verification, request logging |
| auth.go | Registration, login, JWT generation |
| entries.go | CRUD operations for entries |
| db.go | Database layer, SQL queries |

---

## 💡 Quick Answers

**Q: How do I add a new endpoint?**

1. Create handler function in appropriate file
2. Register route in main.go
3. Add middleware if authentication needed

**Q: How do I add a new field to entries?**

1. Add field to struct with JSON tag
2. Add validation in handler
3. Update database INSERT query
4. Update database SELECT query

**Q: How do I debug authentication issues?**

1. Check if JWT_SECRET is set
2. Check token format: "Bearer <token>"
3. Check token expiration
4. Look at server logs for specific error

**Q: Where do I add validation?**
In handler functions, after parsing JSON, before database call.

**Q: How do I test endpoints?**
Run `test-all.ps1` or use individual test scripts.

---

## 🚀 Common Tasks

### Start Server

```powershell
go run .\cmd\server\main.go
```

### Run All Tests

```powershell
.\test-all.ps1
```

### Check for Errors

```powershell
go build .\cmd\server\main.go
```

### Format Code

```powershell
go fmt ./...
```

---

## 🎓 Learning Resources

- **Official Go Docs:** <https://go.dev/doc/>
- **JWT Spec:** <https://jwt.io/>
- **bcrypt:** <https://pkg.go.dev/golang.org/x/crypto/bcrypt>

---

## 🔍 Debugging Checklist

**Server won't start:**

- [ ] JWT_SECRET environment variable set?
- [ ] Port 8080 already in use?
- [ ] Go dependencies installed? (`go mod download`)

**Authentication failing:**

- [ ] Token in correct format: "Bearer <token>"?
- [ ] Token not expired? (24 hour limit)
- [ ] JWT_SECRET matches what was used to sign token?

**Database errors:**

- [ ] DB_PATH environment variable correct?
- [ ] Database file permissions okay?
- [ ] Tables created? (InitDB runs on startup)

**Validation errors:**

- [ ] JSON field names match struct tags exactly?
- [ ] Required fields present?
- [ ] Data types correct? (int vs string)

---

**Remember:** You don't need to memorize this. Bookmark this file and reference it when needed. Professional developers look up syntax constantly!
