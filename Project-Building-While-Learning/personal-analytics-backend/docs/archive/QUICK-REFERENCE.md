# Quick Reference - Personal Analytics Backend

**Purpose:** When you forget what a function does, look here first!

---

## 📦 Handler Functions (internal/handlers/entries.go)

### `respondJSON(w, status, data)`

**What it does:** Sends any data back as JSON with HTTP status code
**When to use:** Success responses, complex data
**Example:**

```go
respondJSON(w, http.StatusOK, map[string]interface{}{
    "success": true,
    "count": 5,
})
```

### `errorResponse(w, status, message)`

**What it does:** Sends standardized error message
**When to use:** Validation errors, any failure
**Example:**

```go
errorResponse(w, http.StatusBadRequest, "text cannot be empty")
```

**Pattern:** Use `errorResponse` for errors, `respondJSON` for everything else

---

## 🗄️ Database Functions (internal/db/db.go)

### `db.InitDB(dbPath)`

**What it does:** Opens database connection, creates tables
**When to use:** Once at server startup in main.go
**Returns:** error if connection fails

### `db.InsertEntry(userID, text, mood, category)`

**What it does:** Saves new entry to database
**When to use:** After validation passes in POST handler
**Returns:** (id int64, error)
**Example:**

```go
id, err := db.InsertEntry(101, "Great day", 8, "work")
if err != nil {
    // Handle error
}
```

### `db.GetAllEntries()`

**What it does:** Retrieves all entries from database
**When to use:** GET /entries handler
**Returns:** ([]map[string]interface{}, error)

### `db.CloseDB()`

**What it does:** Closes database connection
**When to use:** Server shutdown (with defer in main.go)

---

## 🔒 HTTP Status Codes (When to Use What)

| Code | Name | When to Use |
|------|------|-------------|
| 200 | OK | GET requests successful |
| 201 | Created | POST created new resource |
| 400 | Bad Request | Validation failed, invalid input |
| 404 | Not Found | Resource doesn't exist |
| 405 | Method Not Allowed | Wrong HTTP method (GET vs POST) |
| 500 | Internal Server Error | Database error, server crashed |

**Quick decision:**

- Client's fault (bad data) → 400
- Server's fault (database error) → 500
- Success with data → 200
- Success created → 201

---

## 🏗️ Request Flow Pattern

**Every handler follows this:**

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Log request
    log.Println("📨 Request received")

    // 2. Check HTTP method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", 405)
        return
    }

    // 3. Parse JSON (if POST/PUT)
    var req MyRequestStruct
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        errorResponse(w, 400, "Invalid JSON")
        return
    }

    // 4. Validate fields (one by one)
    if req.Field1 == "" {
        errorResponse(w, 400, "field1 required")
        return
    }

    // 5. Database operation
    result, err := db.SomeFunction(req.Field1)
    if err != nil {
        errorResponse(w, 500, "Database error")
        return
    }

    // 6. Success response
    respondJSON(w, 200, MyResponse{Success: true, Data: result})
}
```

**Memorize this flow, not individual functions!**

---

## 📋 Common Go Patterns You'll See

### Multiple Return Values

```go
result, err := someFunction()
if err != nil {
    // Handle error
    return
}
// Use result
```

**Pattern:** Functions return (result, error). Always check error first.

### Defer (Cleanup)

```go
file, err := os.Open("data.txt")
if err != nil {
    return err
}
defer file.Close()  // Runs at end, even if error happens
```

**Pattern:** Use defer right after opening/acquiring resource.

### Struct Tags

```go
type User struct {
    ID   int    `json:"user_id"`    // JSON key "user_id" → field ID
    Name string `json:"name"`
}
```

**Pattern:** Tag value must exactly match JSON key from API.

### Parameterized Queries

```go
// ✅ SAFE
query := "SELECT * FROM users WHERE id = ?"
db.Query(query, userInput)

// ❌ UNSAFE
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userInput)
```

**Pattern:** Always use ? placeholders, never concatenate user input.

---

## 🔍 When You Forget Something

### 1. Check Function Signature

Hover over function in VS Code or look at definition:

```go
func respondJSON(w http.ResponseWriter, status int, data interface{})
```

Read it as: "respondJSON takes (response writer, status code, any data)"

### 2. Look at Existing Usage

Search your code for how you used it before:

- `Ctrl+F` → search for function name
- See how it's called in other places
- Copy the pattern

### 3. Read Comments Above Function

Every function should have a comment explaining it:

```go
// errorResponse sends a standardized error message as JSON
// Parameters: w (response), status (HTTP code), message (error text)
func errorResponse(w http.ResponseWriter, status int, message string) {
```

### 4. Check This File (QUICK-REFERENCE.md)

Bookmark this file for instant lookup.

---

## 💡 Learning Strategy

### DON'T Try to Memorize

- ❌ Exact function names
- ❌ Parameter order
- ❌ Syntax details

### DO Understand

- ✅ **Patterns** (request → validate → database → response)
- ✅ **Purpose** (why this function exists)
- ✅ **When to use** (what problem it solves)

### Example

**Bad approach:** "I need to remember respondJSON takes w, status, data"
**Good approach:** "When I want to send JSON back, I use respondJSON"

Then look up the syntax when you need it!

---

## 🎯 Daily Practice

**When coding:**

1. If you forget a function, check this file (2 seconds)
2. Look at existing usage in your code (5 seconds)
3. Read the comment above the function (10 seconds)

**DON'T:**

- Stare at blank screen trying to remember
- Feel bad about looking things up
- Think "real developers don't need references"

**Real developers look things up 50+ times per day!**

---

## 📚 Function Categories (Mental Model)

Think of functions in groups:

**Response Functions** (sending data back):

- `respondJSON` → any data
- `errorResponse` → just errors
- `http.Error` → simple text error

**Database Functions** (data operations):

- `InsertEntry` → create
- `GetAllEntries` → read all
- `GetEntryByID` → read one (Week 2)

**Validation Functions** (checking data):

- `if req.Field == ""` → check empty
- `if req.Field < 1 || req.Field > 10` → check range
- `json.Decode` → check valid JSON

**Group in your mind:** "I need to send a response → which response function?"

---

## � Authentication Functions (Week 2)

### **Password Hashing (Registration)**

```go
// Convert plain password to secure hash
passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    // Handle error
}
// Save string(passwordHash) to database
```

**What it does:** One-way encryption of password (can't reverse)
**When to use:** Registration - before saving to database
**Remember:** `[]byte(password)` just converts string to bytes (library requirement)

---

### **Password Verification (Login)**

```go
// Compare stored hash with user's password
err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))
if err != nil {
    // Password doesn't match → 401 error
    return
}
// Password matches → generate token
```

**What it does:** Checks if password matches the hash
**When to use:** Login - after getting user from database
**Remember:** Hash first, password second (parameter order matters!)
**Pattern:** Just like `strings.Contains(email, "@")` - function returns error if false

---

### **JWT Token Generation (Login)**

```go
// STEP 1: Create claims (data inside token)
claims := jwt.MapClaims{
    "user_id": userID,
    "exp": time.Now().Add(24 * time.Hour).Unix(), // Expires in 24 hours
}

// STEP 2: Create token with claims
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// STEP 3: Sign token with secret key
tokenString, err := token.SignedString([]byte(secret))
if err != nil {
    // Handle error
}
// Return tokenString to user
```

**What it does:** Creates encrypted token containing user_id
**When to use:** Login - after password verification succeeds
**Think of it as:** Creating a "ticket" user shows to prove they logged in
**Parts breakdown:**

- `jwt.MapClaims{...}` → Map (like `map[string]interface{}` you know!)
- `jwt.NewWithClaims(...)` → Function call (like `db.CreateUser(...)`)
- `token.SignedString(...)` → Convert to string (like `string(hash)`)

**NOT SCARY:** It's just 3 function calls, one after another!

---

### **Database Auth Functions**

```go
// Check if user exists and get their hash
userID, passwordHash, err := db.GetUserByEmail(email)
if err != nil {
    // User doesn't exist
}

// Create new user
userID, err := db.CreateUser(email, passwordHash)
```

**Pattern:** Same as other db functions you already know!

---

## 💡 Dealing with "Scary" Library Syntax

### **The Truth About Libraries:**

**You feel scared because:**

- Unfamiliar function names
- New parameter types (`[]byte`, `jwt.MapClaims`)
- Multiple steps chained together

**The reality:**

- It's just function calls (you do this all day!)
- `[]byte()` is just type conversion (like `string()` or `int()`)
- Libraries always look scary at first → become normal after using 2-3 times

### **What Professional Developers Do:**

1. ✅ Google "golang bcrypt compare password example"
2. ✅ Copy working code from docs
3. ✅ Modify for their use case
4. ✅ Add comments explaining what it does
5. ✅ Move on (don't try to memorize!)

**Nobody memorizes library syntax.** We all look it up. Every. Single. Time.

### **Your Strategy:**

When you see "scary" syntax:

1. **Break it down** - What's the function name? What's the parameter?
2. **Compare to known patterns** - Is it like `db.CreateUser(param1, param2)`?
3. **Add to this file** - Write example so you never Google again
4. **Use it 2-3 times** - Then it's not scary anymore

**Example:**

```go
// SCARY at first:
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

// After breaking down:
err := someFunction(parameter1, parameter2)
           ↓            ↓           ↓
      function name   hash    user's password
```

**It's just a function call!** The name is long, but the concept is simple.

---

## �🚀 Next Steps

1. **Bookmark this file** for quick lookup
2. **Add to it** when you learn new functions (Week 2 auth functions)
3. **Use it actively** - don't feel bad about checking!
4. **After 2 weeks** you'll need it less (patterns stick)

---

**Remember:** Professional developers Google things CONSTANTLY. The difference is they know WHAT to look for and WHERE to find it. This file is your "WHERE".
