# Personal Analytics Backend

A backend service for tracking personal data (mood, habits, activities) with AI-assisted insights.

## Why This Exists

This service is designed to be the backend engine for a personal analytics platform. Instead of building yet another TODO app, this focuses on the harder problems: data ingestion, async processing, AI integration, and scalability—the kind of backend work that companies actually pay for.

## Tech Stack

- **Language:** Go 1.25
- **Framework:** net/http (stdlib)
- **Database:** PostgreSQL (planned)
- **Environment:** godotenv for config

## Current Status (Day 1 - Jan 2, 2026)

✅ HTTP server running on configurable port  
✅ Health check endpoint  
✅ Environment-based configuration  
✅ Clean project structure (cmd/internal separation)

## How to Run

```bash
# Install dependencies
go mod tidy

# Run the server
go run ./cmd/server/main.go

# Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/ping
```

## Project Structure

```
personal-analytics-backend/
├── cmd/
│   └── server/          # Main entry point
├── internal/
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models (coming Day 2)
│   └── db/              # Database layer (coming Day 2)
├── .env                 # Environment config
└── go.mod               # Dependencies
```

## Roadmap

**Week 1:** Foundation + CRUD  
**Week 2:** Auth + JWT  
**Week 3:** AI processing (async workers)  
**Week 4:** Caching + scaling considerations

## What I Built Today

Set up a Go backend with a running HTTP server and basic project structure. The server responds to health checks and demonstrates environment-based configuration.
