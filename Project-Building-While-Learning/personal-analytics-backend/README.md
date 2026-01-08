# Personal Analytics Backend

A production-ready Go backend with CRUD operations, database persistence, and security best practices.

## Why This Exists

This service is the backend engine for a personal analytics platform. Instead of building yet another TODO app, this focuses on real backend problems: data persistence, validation, security, and scalability—the kind of work that companies actually pay for.

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
HTTP Request
    ↓
Handler Layer (internal/handlers/)
    - HTTP method validation
    - JSON parsing with struct tags
    - Input validation (sequential checks)
    ↓
Database Layer (internal/db/)
    - Parameterized SQL queries
    - CRUD operations (Create, Read)
    ↓
SQLite Storage (data.db)
    - Persistent file-based database
```

## API Endpoints

### POST /entries
Creates a new mood/activity entry.

**Request:**
```json
{
  "user_id": 101,
  "text": "Productive day!",
  "mood": 8,
  "category": "work"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Entry created successfully",
  "id": 1
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
