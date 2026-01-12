# 5-Minute Cheat Sheet

**Use this when you forget something. Don't memorize - just bookmark.**

---

## 🔥 THE ONLY 5 PATTERNS YOU NEED

### **1. Create a Handler**

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Check method
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", 405)
        return
    }

    // 2. Parse JSON
    var req MyRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 3. Validate
    if req.Field == "" {
        errorResponse(w, 400, "field is required")
        return
    }

    // 4. Database operation
    result, err := db.MyDBFunction(req.Field)
    if err != nil {
        errorResponse(w, 500, "Database error")
        return
    }

    // 5. Success response
    respondJSON(w, 200, MyResponse{Success: true, Data: result})
}
```

---

### **2. Database Query**

```go
func MyDBFunction(param string) (result, error) {
    query := `SELECT * FROM table WHERE field = ?`
    // Use ? for parameters (prevents SQL injection)

    rows, err := DB.Query(query, param)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Process rows...
    return result, nil
}
```

---

### **3. Hash Password (Registration)**

```go
// Hash it
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    return err
}

// Save string(hash) to database
db.CreateUser(email, string(hash))
```

---

### **4. Verify Password (Login)**

```go
// Get hash from database
userID, hash, err := db.GetUserByEmail(email)

// Compare
err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
if err != nil {
    // Wrong password
    return
}

// Password correct - generate token
```

---

### **5. Generate JWT Token (Login)**

```go
// Create claims
claims := jwt.MapClaims{
    "user_id": userID,
    "exp": time.Now().Add(24 * time.Hour).Unix(),
}

// Create and sign token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

// Return tokenString to user
```

---

## 🎯 QUICK ANSWERS

**Q: How do I protect a route?**

```go
http.HandleFunc("/myroute", handlers.AuthMiddleware(handlers.MyHandler))
```

**Q: How do I get user_id in a protected handler?**

```go
userID := r.Context().Value("user_id").(int64)
```

**Q: What status codes to use?**

- 200 OK - Success (GET)
- 201 Created - Success (POST creating something)
- 400 Bad Request - Validation error
- 401 Unauthorized - Not logged in / invalid token
- 404 Not Found - Resource doesn't exist
- 500 Internal Server Error - Database/server error

**Q: How do I send JSON response?**

```go
respondJSON(w, statusCode, dataStruct)
```

**Q: How do I parse JSON request?**

```go
var req MyStruct
json.NewDecoder(r.Body).Decode(&req)
```

**Q: Why []byte()?**
Libraries need bytes, not strings. Just convert:

```go
[]byte(myString)  // string → bytes
string(myBytes)   // bytes → string
```

---

## 🔑 FILES QUICK REFERENCE

- **main.go** → Routes requests
- **middleware.go** → Checks JWT tokens
- **auth.go** → Register/Login
- **entries.go** → CRUD operations
- **db.go** → Database functions

---

**That's it. These 5 patterns cover 90% of your code.**
