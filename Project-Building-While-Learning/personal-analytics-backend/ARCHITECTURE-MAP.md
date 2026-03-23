# Personal Analytics Backend - Architecture Map

**Purpose:** Visual guide showing how everything connects. Read this when you feel lost.

---

## 🏗️ THE BIG PICTURE

```
USER SENDS REQUEST
        ↓
   [main.go] - Routes requests to handlers
        ↓
   [middleware.go] - Checks JWT token (if protected route)
        ↓
   [handlers/*.go] - Processes request, validates data
        ↓
   [db/db.go] - Saves/retrieves from database
        ↓
   [data.db] - SQLite database file
        ↓
   RESPONSE SENT BACK
```

---

## 📁 FILE-BY-FILE BREAKDOWN

### **cmd/server/main.go** (Entry Point)

**What it does:** Starts server, registers routes

```
Routes:
├── /health      → No auth needed
├── /ping        → No auth needed
├── /register    → No auth needed (anyone can register)
├── /login       → No auth needed (anyone can login)
└── /entries     → 🔒 PROTECTED (needs JWT token)
                   Wrapped with AuthMiddleware
```

**Key line:**

```go
http.HandleFunc("/entries", handlers.AuthMiddleware(func...))
                            ↑ This wraps the handler with protection
```

---

### **internal/handlers/middleware.go** (Security Guard)

**What it does:** Checks if user is logged in before allowing access

```
Request arrives
    ↓
1. Get "Authorization" header
    ↓
2. Extract JWT token (remove "Bearer " prefix)
    ↓
3. Verify token with secret key
    ↓
4. Extract user_id from token
    ↓
5. Put user_id in request context (like a backpack)
    ↓
6. Pass request to next handler
```

**If anything fails → 401 Unauthorized error**

---

### **internal/handlers/auth.go** (Login/Register)

**What it does:** Creates user accounts and issues JWT tokens

#### Register Flow

```
POST /register
{"email": "user@example.com", "password": "secret"}
    ↓
1. Validate email (not empty, has @)
2. Validate password (not empty, >= 6 chars)
3. Hash password with bcrypt
4. Save to database (users table)
5. Return success with user_id
```

#### Login Flow

```
POST /login
{"email": "user@example.com", "password": "secret"}
    ↓
1. Validate email + password not empty
2. Get user from database by email
3. Compare password with stored hash (bcrypt)
4. Generate JWT token with user_id inside
5. Return token to user
```

**User keeps this token and sends it with every request!**

---

### **internal/handlers/entries.go** (Data Operations)

**What it does:** Creates and retrieves user entries

#### Create Entry Flow

```
POST /entries (🔒 protected)
{"text": "Had a great day", "mood": 8, "category": "personal"}
    ↓
1. Middleware already verified user (user_id in context)
2. Validate all fields (text, mood 1-10, category)
3. Save to database with user_id
4. Return success with entry_id
```

#### Get Entries Flow

```
GET /entries (🔒 protected)
    ↓
1. Middleware already verified user (user_id in context)
2. Get user_id from request context
3. Fetch ONLY entries for this user from database
4. Return entries array
```

**Key concept:** Each user only sees THEIR OWN entries!

---

### **internal/db/db.go** (Database Layer)

**What it does:** Talks to SQLite database

#### Tables

```
users table:
├── id (auto-increment)
├── email (unique)
├── password_hash (bcrypt hash)
└── created_at

entries table:
├── id (auto-increment)
├── user_id (links to users.id)
├── text
├── mood
├── category
└── created_at
```

#### Functions

- `InitDB()` - Opens database connection
- `CreateUser()` - Inserts new user
- `GetUserByEmail()` - Finds user for login
- `InsertEntry()` - Creates new entry
- `GetEntriesByUser()` - Gets entries for specific user
- `CloseDB()` - Closes connection when server stops

---

## 🔐 AUTHENTICATION FLOW (The Complete Journey)

### **Registration:**

```
1. User sends email + password
2. Server hashes password with bcrypt
3. Server saves email + hash to database
4. Server returns "success"
```

### **Login:**

```
1. User sends email + password
2. Server finds user by email
3. Server compares password with stored hash
4. If match → generate JWT token
5. Token contains: {"user_id": 123, "exp": 1736445930}
6. Server returns token to user
```

### **Using Protected Routes:**

```
1. User sends request with header:
   Authorization: Bearer eyJhbGci...xyz

2. Middleware extracts token from header

3. Middleware verifies token:
   - Is signature valid? (checks with JWT_SECRET)
   - Has it expired? (checks "exp" claim)

4. If valid → extract user_id from token

5. Put user_id in context (backpack)

6. Pass request to handler

7. Handler gets user_id from context

8. Handler uses user_id to fetch/save data
```

---

## 🔑 KEY CONCEPTS (The "Why" Behind the Code)

### **1. Why Hash Passwords?**

```
Plain password in DB:  "mysecret123"  ← BAD! If DB leaked, all passwords exposed
Hashed password in DB: "$2a$10$xyz..." ← GOOD! Can't reverse to get original
```

### **2. Why JWT Tokens?**

```
Without JWT:
- User logs in → Server creates session in memory
- Problem: Server needs to remember ALL logged-in users
- Doesn't work with multiple servers

With JWT:
- User logs in → Server creates token (like a ticket)
- Token contains user_id (who they are)
- Token is signed (proves it's authentic)
- User sends token with every request
- Server doesn't need to remember anything!
```

### **3. Why Middleware?**

```
Without Middleware:
- Every handler copies same auth code
- Easy to forget checking auth
- Hard to maintain

With Middleware:
- Auth code in ONE place
- Wrap any route to protect it
- Can't forget (route won't work without it)
```

### **4. Why Context?**

```
Problem: Middleware verifies user, but how does handler know WHO the user is?

Solution: Context (like a backpack)
- Middleware: "I verified this is user 123" → puts in context
- Handler: "Let me check the context... oh, user 123!" → reads from context
- Context travels with the request through all layers
```

---

## 📊 DATA FLOW EXAMPLE

### **Creating an Entry (Full Journey):**

```
1. USER ACTION:
   POST /entries
   Header: Authorization: Bearer eyJhbGci...xyz
   Body: {"text": "Great day!", "mood": 9, "category": "work"}

2. MAIN.GO:
   "Request for /entries → route to AuthMiddleware(CreateEntry)"

3. MIDDLEWARE.GO:
   ├─ Get token: "eyJhbGci...xyz"
   ├─ Verify with JWT_SECRET ✓
   ├─ Extract user_id: 123
   ├─ Put in context: {"user_id": 123}
   └─ Call next handler (CreateEntry)

4. ENTRIES.GO (CreateEntry):
   ├─ Get user_id from context: 123
   ├─ Validate: text not empty ✓
   ├─ Validate: mood 1-10 ✓
   ├─ Validate: category not empty ✓
   └─ Call db.InsertEntry(123, "Great day!", 9, "work")

5. DB.GO:
   ├─ SQL: INSERT INTO entries (user_id, text, mood, category) VALUES (?, ?, ?, ?)
   ├─ Execute with parameters: [123, "Great day!", 9, "work"]
   ├─ Get inserted ID: 456
   └─ Return ID to handler

6. ENTRIES.GO:
   └─ Return JSON: {"success": true, "message": "Entry created", "id": 456}

7. USER RECEIVES:
   Status: 201 Created
   Body: {"success": true, "message": "Entry created successfully", "id": 456}
```

---

## 🎯 COMMON PATTERNS (Memorize These, Not Syntax)

### **Pattern 1: Handler Structure**

```
func Handler(w http.ResponseWriter, r *http.Request) {
    1. Check method (POST/GET)
    2. Parse request body (JSON)
    3. Validate each field (early returns on error)
    4. Call database function
    5. Return JSON response
}
```

### **Pattern 2: Database Function**

```
func DBFunction(params...) (result, error) {
    1. Write SQL query with ? placeholders
    2. Execute with DB.Exec() or DB.Query()
    3. Handle errors
    4. Return result
}
```

### **Pattern 3: Error Handling**

```
result, err := someFunction()
if err != nil {
    // Log error
    // Return error response
    return
}
// Continue with result
```

### **Pattern 4: JSON Response**

```
respondJSON(w, statusCode, dataStruct)
```

---

## 💡 WHEN YOU FORGET SOMETHING

### "How do I verify JWT token?"

→ Look at `middleware.go` line 60-85

### "How do I hash a password?"

→ Look at `auth.go` line 95-100

### "How do I get user_id from context?"

→ Look at `entries.go` line 115-125

### "How do I protect a route?"

→ Look at `main.go` line 74 (wrap with AuthMiddleware)

---

## 🚀 TESTING YOUR API

### 1. Register a user

```powershell
.\test-register.ps1
```

### 2. Login

```powershell
.\test-login.ps1
# Saves token in $token variable
```

### 3. Access protected route

```powershell
.\test-middleware.ps1
# Tests with/without token
```

### 4. Create entry

```powershell
# Start server first:
go run .\cmd\server\main.go

# In another terminal:
$token = "paste_your_token_here"
$headers = @{"Authorization" = "Bearer $token"}
$body = @{text="Great day"; mood=9; category="work"} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:8080/entries `
    -Method Post -Headers $headers -Body $body -ContentType "application/json"
```

---

## ✅ WHAT YOU'VE BUILT (Be Proud!)

- ✅ User registration with secure password hashing
- ✅ User login with JWT token generation
- ✅ Protected routes requiring authentication
- ✅ User-specific data isolation
- ✅ Proper error handling and validation
- ✅ RESTful API with JSON responses
- ✅ SQLite database with proper schema

**This is a production-ready authentication system!**

---

## 📚 NEXT STEPS (When Ready)

**Don't do these now. Just know they exist:**

1. Week 2 Day 10: Protect POST /entries endpoint
2. Week 2 Day 11: Add token expiration handling
3. Week 2 Day 12: Testing and refinement
4. Week 2 Day 13: Documentation and review

**For now: Just USE what you built. Understanding comes from using, not reading.**

---

**Remember:** You don't need to memorize syntax. You need to understand:

1. What each file does
2. How data flows between them
3. When to look up syntax (always!)

**Professional developers Google syntax daily. You should too.**
