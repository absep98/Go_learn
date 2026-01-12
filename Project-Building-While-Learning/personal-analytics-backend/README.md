# Personal Analytics Backend

A production-ready Go backend with JWT authentication, protected API endpoints, database persistence, and comprehensive testing.

## Why This Exists

This service is the backend engine for a personal analytics platform. Instead of building yet another TODO app, this focuses on real backend problems: authentication, data persistence, security, and scalability—the kind of work that companies actually pay for.

## Tech Stack

- **Language:** Go 1.25.3
- **Database:** SQLite (modernc.org/sqlite)
- **Authentication:** JWT tokens (github.com/golang-jwt/jwt/v5)
- **Security:** bcrypt password hashing
- **Testing:** PowerShell test scripts

## What I Built

### Week 1 (Jan 2-7): Foundation

- HTTP server with `net/http` standard library
- SQLite database connection and schema design
- Environment-based configuration
- CRUD operations (POST/GET /entries)
- Input validation and error handling
- SQL injection prevention
- Request logging middleware

### Week 2 (Jan 8-12): Authentication System

- User registration with email validation
- bcrypt password hashing (never store plain text)
- JWT token generation on login
- Authentication middleware for protected routes
- User-specific data isolation
- Comprehensive test suite (18 tests)
- Error message improvements

## Authentication Flow

```
1. REGISTRATION
   User → POST /register {email, password}
        → Validate email format (@ required)
        → Hash password with bcrypt
        → Save to database
        → Return success + user_id

2. LOGIN
   User → POST /login {email, password}
        → Get user from database
        → Compare password with hash (bcrypt)
        → Generate JWT token (includes user_id)
        → Return token

3. ACCESSING PROTECTED ROUTES
   User → GET/POST /entries (Header: Authorization: Bearer <token>)
        → Middleware checks token
        → Verify signature with JWT_SECRET
        → Extract user_id from token
        → Pass to handler in context
        → Handler uses authenticated user_id
        → Return user-specific data
```

**Key Security Points:**

- Passwords never stored in plain text (bcrypt hashing)
- JWT tokens signed with secret (tamper-proof)
- User_id from verified token only (no client manipulation)
- Each user only sees their own entries

## What I Built (Week 1)

### Days 1-2: Foundation

- HTTP server with `net/http` standard library
- SQLite database connection and schema design
- Environment-based configuration
- Clean project structure (`cmd/` and `internal/` separation)

### Days 3-4: Core Features

- **POST /entries** - Create mood/activity entries with full validation
- **GET /entries** - Retrieve all entries as structured JSON
- Input validation: userID, text length, mood range (1-10), category
- Database persistence with parameterized queries (SQL injection safe)

### Days 5-6: Security & Consolidation

- Learned SQL injection prevention (parameterized vs string concatenation)
- Mastered JSON struct tags for API mapping
- Implemented consistent error response patterns
- Added request logging for debugging
- Code review and cleanup

## Architecture

```
User Request
    ↓
main.go - Route Registration
    ↓
LoggingMiddleware (logs all requests)
    ↓
AuthMiddleware (for protected routes)
    - Verify JWT token
    - Extract user_id
    - Put in context
    ↓
Handler Layer (internal/handlers/)
    - HTTP method validation
    - JSON parsing with struct tags
    - Input validation (sequential checks)
    - Get user_id from context (protected routes)
    ↓
Database Layer (internal/db/)
    - Parameterized SQL queries (SQL injection safe)
    - CRUD operations
    ↓
SQLite Storage (data.db)
    - users table (email, password_hash)
    - entries table (user_id, text, mood, category)
```

## API Endpoints

See [API-ENDPOINTS.md](API-ENDPOINTS.md) for detailed documentation.

### Public Endpoints (No Authentication)

**POST /register** - Create new user account
**POST /login** - Get JWT token

### Protected Endpoints (Requires JWT Token)

**GET /entries** - Retrieve user's entries
**POST /entries** - Create new entry

### Utility Endpoints

**GET /health** - Health check
**GET /ping** - Connection test

## Quick Start

### Prerequisites

- Go 1.25.3 or higher
- Git

### Installation

```bash
# Clone repository
git clone https://github.com/absep98/Go_learn.git
cd Go_learn/Project-Building-While-Learning/personal-analytics-backend

# Install dependencies
go mod download

# Set environment variables (Windows PowerShell)
$env:JWT_SECRET = "your-secret-key-here"
$env:PORT = "8080"
$env:DB_PATH = "./data.db"

# Or create .env file
# JWT_SECRET=your-secret-key-here
# PORT=8080
# DB_PATH=./data.db

# Run server
go run .\cmd\server\main.go
```

Server will start on `http://localhost:8080`

### Running Tests

```powershell
# Terminal 1: Start server
go run .\cmd\server\main.go

# Terminal 2: Run all tests
.\test-all.ps1

# Or run individual test suites
.\test-register.ps1
.\test-login.ps1
.\test-middleware.ps1
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| JWT_SECRET | Yes | - | Secret key for signing JWT tokens |
| PORT | No | 8080 | Server port |
| DB_PATH | No | ./data.db | SQLite database file path |

## Usage Examples

### 1. Register a New User

```powershell
$body = @{
    email = "user@example.com"
    password = "password123"
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:8080/register -Method Post -Body $body -ContentType "application/json"
```

### 2. Login and Get Token

```powershell
$body = @{
    email = "user@example.com"
    password = "password123"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri http://localhost:8080/login -Method Post -Body $body -ContentType "application/json"
$token = $response.token
```

### 3. Create Entry (with token)

```powershell
$headers = @{
    "Authorization" = "Bearer $token"
}

$body = @{
    text = "Had a great day!"
    mood = 8
    category = "personal"
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:8080/entries -Method Post -Headers $headers -Body $body -ContentType "application/json"
```

### 4. Get Entries (with token)

```powershell
$headers = @{
    "Authorization" = "Bearer $token"
}

Invoke-RestMethod -Uri http://localhost:8080/entries -Method Get -Headers $headers
```

## Project Structure

```
personal-analytics-backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point, route registration
├── internal/
│   ├── handlers/
│   │   ├── auth.go              # Register/Login handlers
│   │   ├── entries.go           # CRUD operations for entries
│   │   ├── health.go            # Health check
│   │   ├── middleware.go        # JWT authentication middleware
│   │   └── loggingMiddleware.go # Request logging
│   ├── db/
│   │   └── db.go                # Database layer (SQLite)
│   └── models/                  # (Future: data models)
├── learning-demos/
│   ├── 01-sql-injection-demo.go # Security demo
│   └── 02-json-struct-tags-practice.go # JSON mapping practice
├── test-all.ps1                 # Comprehensive test script
├── test-register.ps1            # Registration tests
├── test-login.ps1               # Login tests
├── test-middleware.ps1          # Auth middleware tests
├── ARCHITECTURE-MAP.md          # Visual architecture guide
├── CHEAT-SHEET.md               # Quick syntax reference
├── TESTING.md                   # Test documentation
├── README.md                    # This file
├── .env                         # Environment variables (not in git)
├── data.db                      # SQLite database (not in git)
└── go.mod                       # Go dependencies
```

## Database Schema

### users table

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
```

### entries table

```sql
CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    text TEXT NOT NULL,
    mood INTEGER NOT NULL,
    category TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
)
```

## Testing

**Test Coverage:** 18 comprehensive tests

- Registration validation (4 tests)
- Login authentication (4 tests)
- Middleware protection (3 tests)
- Entry CRUD operations (4 tests)
- Data isolation (3 tests)

See [TESTING.md](TESTING.md) for detailed test results.

**All tests passing ✅**

## What I Learned

### Week 1

- Go HTTP server fundamentals
- Database design and SQLite integration
- JSON struct tags and API design
- SQL injection prevention (parameterized queries)
- Input validation patterns
- Error handling best practices

### Week 2

- JWT token generation and verification
- bcrypt password hashing
- Middleware pattern for authentication
- Context for passing data between middleware and handlers
- User data isolation
- Comprehensive testing strategies
- Better error messages for debugging

## Next Steps (Week 3)

- Implement UPDATE and DELETE operations
- Add pagination for GET /entries
- Caching layer for improved performance
- Rate limiting
- Background job processing (async)
- Metrics and monitoring

## Why This Matters for Interviews

This project demonstrates:

- ✅ **Authentication systems** - JWT, bcrypt, middleware
- ✅ **Security awareness** - SQL injection prevention, password hashing
- ✅ **API design** - RESTful endpoints, proper status codes
- ✅ **Database design** - Schema, relationships, queries
- ✅ **Testing** - Comprehensive test coverage
- ✅ **Clean architecture** - Separation of concerns, middleware pattern
- ✅ **Production readiness** - Logging, error handling, validation

## Resources

- [ARCHITECTURE-MAP.md](ARCHITECTURE-MAP.md) - Visual guide to how everything connects
- [CHEAT-SHEET.md](CHEAT-SHEET.md) - Quick reference for common patterns
- [API-ENDPOINTS.md](API-ENDPOINTS.md) - Detailed endpoint documentation
- [TESTING.md](TESTING.md) - Test results and coverage
- [Week1-Review.md](Week1-Review.md) - Week 1 learning reflection

## License

This is a learning project. Feel free to use as reference!

## Contact

GitHub: [@absep98](https://github.com/absep98)

}

```

**Validation Rules:**
- `user_id` must be > 0
- `text` cannot be empty
- `mood` must be 1-10
- `category` cannot be empty

### GET /entries
Retrieves all entries ordered by creation time.

**Response (200 OK):**
```json
{
  "success": true,
  "count": 2,
  "entries": [
    {
      "id": 2,
      "user_id": 101,
      "text": "Productive day!",
      "mood": 8,
      "category": "work",
      "created_at": "2026-01-07T10:30:00Z"
    },
    {
      "id": 1,
      "user_id": 101,
      "text": "Morning workout",
      "mood": 9,
      "category": "health",
      "created_at": "2026-01-07T07:00:00Z"
    }
  ]
}
```

## What I Learned

### 1. 3-Layer Architecture

Separation between handlers (HTTP logic), database layer (queries), and storage (SQLite). This makes code testable, maintainable, and easier to modify.

### 2. JSON Struct Tags

Tags like `json:"user_id"` map between JSON keys (snake_case) and Go fields (PascalCase). Without tags, Go can't match `{"user_id": 101}` to `UserID int`.

### 3. SQL Injection Prevention

**Safe:** `db.Exec("INSERT INTO ... VALUES (?, ?, ?)", val1, val2, val3)`
**Vulnerable:** `db.Exec(fmt.Sprintf("INSERT INTO ... VALUES ('%s')", userInput))`

Parameterized queries prevent attackers from injecting malicious SQL.

### 4. Go Patterns

- Multiple return values for error handling: `result, err := function()`
- `defer` for cleanup: `defer rows.Close()`
- Early returns for validation: `if invalid { return error }`

## Tech Stack

- **Language:** Go 1.25.3
- **Database:** SQLite (modernc.org/sqlite driver)
- **Libraries:** `net/http`, `encoding/json`, `godotenv`
- **Tools:** PowerShell for testing, Git for version control

## Project Structure

```
personal-analytics-backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point, route registration
├── internal/
│   ├── handlers/
│   │   └── entries.go          # HTTP handlers, validation
│   ├── db/
│   │   └── db.go               # Database connection, CRUD
│   └── models/
│       └── entry.go            # Data models (planned for Week 2)
├── learning-demos/
│   ├── 01-sql-injection-demo.go    # Security education
│   └── 02-json-struct-tags-practice.go
├── .env                        # Environment variables
├── data.db                     # SQLite database file
└── README.md
```

## How to Run

```bash
# Install dependencies
go mod tidy

# Start the server
go run cmd/server/main.go

# Server starts on http://localhost:8080
```

## Testing

```powershell
# Test POST endpoint
$body = @{
    user_id = 101
    text = "Great day!"
    mood = 9
    category = "personal"
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:8080/entries -Method Post -Body $body -ContentType "application/json"

# Test GET endpoint
Invoke-RestMethod -Uri http://localhost:8080/entries
```

## Roadmap

**Week 2 (Jan 8-14):** Authentication & Authorization

- JWT token generation and validation
- User-specific data isolation
- Protected routes with middleware

**Week 3 (Jan 15-21):** Concurrency & Background Processing

- Goroutines for async tasks
- Worker pools with channels
- AI summary generation (mock or real)

**Week 4 (Jan 22-28):** Scaling & Production Readiness

- Caching layer (in-memory or Redis)
- Rate limiting
- Metrics and monitoring
- Clean documentation and diagrams

## Interview-Ready Stories

**"Tell me about a backend project you built"**
> "I built a Go REST API with CRUD operations, database persistence, and security best practices. It handles entry creation with validation, uses parameterized queries to prevent SQL injection, and follows a 3-layer architecture for separation of concerns."

**"How do you prevent SQL injection?"**
> "I use parameterized queries with placeholders. Instead of concatenating user input into SQL strings, I pass values separately so the database driver handles escaping automatically."

**"What's the benefit of separating handlers and database logic?"**
> "Testability, reusability, and maintainability. I can test handlers without a real database, reuse database functions across different interfaces, and switching databases only requires changing one layer."

---

**Built during Week 1 (Jan 2-7, 2026) as part of transitioning from SDET to Backend Developer.**
