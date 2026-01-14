package models

import "time"

// =============================================================================
// DATA MODELS - Struct definitions only, NO database logic
// =============================================================================
// These structs define the shape of your data.
// Database operations are in internal/db/db.go
// =============================================================================

// Entry represents a mood/activity log entry
type Entry struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Text      string    `json:"text"`
	Mood      int       `json:"mood"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents a registered user account
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" means NEVER include in JSON (security!)
	CreatedAt    time.Time `json:"created_at"`
}

// =============================================================================
// WHY THIS STRUCTURE?
// =============================================================================
// 1. Models define WHAT your data looks like (structs)
// 2. DB layer (db/db.go) handles HOW to store/retrieve it (queries)
// 3. Handlers handle HTTP requests and call DB functions
//
// This separation means:
// - Change database? Only modify db/db.go
// - Change API format? Only modify handlers
// - Change data shape? Models are the single source of truth
// =============================================================================
