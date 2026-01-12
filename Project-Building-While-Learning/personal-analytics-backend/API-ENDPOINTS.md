# API Endpoints Documentation

**Base URL:** `http://localhost:8080`

---

## 🔓 Public Endpoints (No Authentication Required)

### POST /register

**Description:** Create a new user account

**Authentication:** None required

**Request Body:**

```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

**Validation Rules:**

- Email must contain `@` symbol
- Password must be at least 6 characters
- Email must be unique

**Success Response (201 Created):**

```json
{
    "success": true,
    "message": "User registered successfully",
    "user_id": 1
}
```

**Error Responses:**

**400 Bad Request** - Validation failure

```json
{
    "success": false,
    "message": "email is required"
}
```

```json
{
    "success": false,
    "message": "invalid email format"
}
```

```json
{
    "success": false,
    "message": "password must be at least 6 characters"
}
```

**409 Conflict** - Duplicate email

```json
{
    "success": false,
    "message": "Email already registered"
}
```

---

### POST /login

**Description:** Authenticate user and receive JWT token

**Authentication:** None required

**Request Body:**

```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

**Validation Rules:**

- Email and password required
- Must match existing user credentials

**Success Response (200 OK):**

```json
{
    "success": true,
    "message": "Login successful",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Token Details:**

- Algorithm: HS256
- Expires: 24 hours after generation
- Contains: user_id claim

**Error Responses:**

**400 Bad Request** - Missing fields

```json
{
    "success": false,
    "message": "email is required"
}
```

**401 Unauthorized** - Invalid credentials

```json
{
    "success": false,
    "message": "Invalid email or password"
}
```

**Note:** Same error for wrong password and non-existent user (security best practice)

---

## 🔒 Protected Endpoints (Require JWT Token)

**All protected endpoints require:**

```
Authorization: Bearer <your-jwt-token>
```

### GET /entries

**Description:** Retrieve all entries for authenticated user

**Authentication:** Required (JWT token)

**Request Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:** None

**Success Response (200 OK):**

```json
{
    "success": true,
    "count": 2,
    "entries": [
        {
            "id": 1,
            "user_id": 1,
            "text": "Had a great day!",
            "mood": 8,
            "category": "personal",
            "created_at": "2026-01-12T10:30:00Z"
        },
        {
            "id": 2,
            "user_id": 1,
            "text": "Productive work session",
            "mood": 9,
            "category": "work",
            "created_at": "2026-01-12T14:20:00Z"
        }
    ]
}
```

**Empty Database Response (200 OK):**

```json
{
    "success": true,
    "count": 0,
    "entries": []
}
```

**Error Responses:**

**401 Unauthorized** - No token provided

```json
{
    "success": false,
    "message": "Authorization header required"
}
```

**401 Unauthorized** - Invalid token

```json
{
    "success": false,
    "message": "Invalid or expired token"
}
```

**401 Unauthorized** - Wrong format

```json
{
    "success": false,
    "message": "Invalid authorization format. Use: Bearer <token>"
}
```

---

### POST /entries

**Description:** Create a new entry for authenticated user

**Authentication:** Required (JWT token)

**Request Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
    "text": "Had a great day!",
    "mood": 8,
    "category": "personal"
}
```

**Validation Rules:**

- `text`: Required, cannot be empty
- `mood`: Required, must be integer 1-10
- `category`: Required, cannot be empty
- `user_id`: Automatically extracted from JWT token (not in request body)

**Success Response (201 Created):**

```json
{
    "success": true,
    "message": "Entry created successfully",
    "id": 5
}
```

**Error Responses:**

**400 Bad Request** - Validation failures

```json
{
    "success": false,
    "message": "text cannot be empty"
}
```

```json
{
    "success": false,
    "message": "mood must be between 1 and 10"
}
```

```json
{
    "success": false,
    "message": "category cannot be empty"
}
```

**401 Unauthorized** - Missing/invalid token (same as GET /entries)

**405 Method Not Allowed** - Wrong HTTP method

```json
Method not allowed
```

---

## 🔧 Utility Endpoints

### GET /health

**Description:** Health check for monitoring

**Authentication:** None required

**Success Response (200 OK):**

```
OK
```

---

### GET /ping

**Description:** Simple connectivity test

**Authentication:** None required

**Success Response (200 OK):**

```json
{
    "message": "pong"
}
```

---

## 📊 HTTP Status Codes Reference

| Code | Name | Usage |
|------|------|-------|
| 200 | OK | Successful GET/POST |
| 201 | Created | Resource created (POST /register, POST /entries) |
| 400 | Bad Request | Validation failed, invalid input |
| 401 | Unauthorized | Missing/invalid authentication |
| 405 | Method Not Allowed | Wrong HTTP method |
| 409 | Conflict | Duplicate resource (email already exists) |
| 500 | Internal Server Error | Server/database error |

---

## 🧪 Testing Examples

### PowerShell

**Register:**

```powershell
$body = @{
    email = "test@example.com"
    password = "password123"
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:8080/register -Method Post -Body $body -ContentType "application/json"
```

**Login:**

```powershell
$body = @{
    email = "test@example.com"
    password = "password123"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri http://localhost:8080/login -Method Post -Body $body -ContentType "application/json"
$token = $response.token
```

**Create Entry:**

```powershell
$headers = @{ "Authorization" = "Bearer $token" }
$body = @{
    text = "Great day!"
    mood = 8
    category = "personal"
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:8080/entries -Method Post -Headers $headers -Body $body -ContentType "application/json"
```

**Get Entries:**

```powershell
$headers = @{ "Authorization" = "Bearer $token" }
Invoke-RestMethod -Uri http://localhost:8080/entries -Method Get -Headers $headers
```

### cURL

**Register:**

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

**Login:**

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

**Create Entry:**

```bash
curl -X POST http://localhost:8080/entries \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{"text":"Great day!","mood":8,"category":"personal"}'
```

**Get Entries:**

```bash
curl -X GET http://localhost:8080/entries \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## 🔐 Security Notes

1. **JWT Token Expiration:** Tokens expire after 24 hours
2. **Password Storage:** Passwords hashed with bcrypt (never stored plain text)
3. **User Isolation:** Users only see their own entries (user_id from token)
4. **SQL Injection:** All queries use parameterized statements
5. **Error Messages:** Authentication errors don't reveal user existence

---

## 📝 Response Format Consistency

All responses follow this pattern:

**Success:**

```json
{
    "success": true,
    "message": "...",
    "data": { }
}
```

**Error:**

```json
{
    "success": false,
    "message": "..."
}
```
