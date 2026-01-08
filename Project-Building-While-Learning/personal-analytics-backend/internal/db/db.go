package db

import (
	"database/sql"
	"log"

	// The underscore _ means "blank import"
	// // You DON'T call any functions from it
	// So why import it if you don't use it?
	// The Magic:
	// This package registers itself with database/sql when imported.
	// How it works:
	// You write: sql.Open("sqlite", "data.db")
	// database/sql says: "Who handles 'sqlite'?"
	// modernc.org/sqlite raises its hand: "I do!"
	// Connection established
	_ "modernc.org/sqlite"
)

// DB is the global database connection
// It's the middleman between your Go code and the database file.
// "Create a variable called DB that will hold the connection to the database."
// This is like having a phone line open to the database. Once connected,
//  you use this DB variable to talk to the database from anywhere in your code.

var DB *sql.DB

// InitDB initializes the database connection and creates tables
func InitDB(dbPath string) error {
	var err error

	// Open SQLite database (creates file if not exists)
	DB, err = sql.Open("sqlite", dbPath) // or ":memory:" as in-memory DB
	if err != nil {
		return err
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		return err
	}

	log.Println("Database connected successfully")

	// Create tables
	err = createTables()
	if err != nil {
		return err
	}

	return nil
}

// createTables creates all required tables if they don't exist
func createTables() error {
	// Users table - stores user accounts
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(usersTable)
	if err != nil {
		return err
	}

	// Entries table - stores mood/activity data
	entriesTable := `
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		text TEXT,
		mood INTEGER,
		category TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(entriesTable)
	if err != nil {
		return err
	}

	log.Println("Tables created successfully")
	return nil
}

// InsertEntry inserts a new entry into the database
// Puts new data INTO the database (like adding a new row to an Excel sheet)
// Takes 5 inputs: userID (which user), text (what they wrote), mood (their mood score), tags (list of tags), category (entry type)
func InsertEntry(userID int, text string, mood int, category string) (int64, error) {
	// The ? marks are placeholders (like blanks in a form)
	query := `INSERT INTO entries (user_id, text, mood, category) VALUES (?, ?, ?, ?)`
	// "Execute the query and fill in the ? marks with actual values."
	result, err := DB.Exec(query, userID, text, mood, category)
	if err != nil {
		return 0, err
	}

	// Get the ID of the inserted row
	// "If something broke, return 0 as ID and the error message."
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("Inserted entry with ID: %d", id)
	return id, nil
}

// GetAllEntries retrieves all entries from the database
// string = field name (like "id", "text", "mood")
// interface{} = any type of value (number, text, etc.)
/*
{
  "id": 1,
  "text": "Happy",
  "mood": 5
}
*/
func GetAllEntries() ([]map[string]interface{}, error) {
	query := `SELECT id, user_id, text, mood, category, tags, created_at FROM entries ORDER BY created_at DESC`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	// "When this function ends, close the rows cursor"
	defer rows.Close()

	var entries []map[string]interface{}

	for rows.Next() {
		var id, userID, mood int
		var text, category, createdAt string
		err := rows.Scan(&id, &userID, &text, &mood, &category, &createdAt)
		if err != nil {
			return nil, err
		}

		entry := map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"text":       text,
			"mood":       mood,
			"category":   category,
			"created_at": createdAt,
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetEntriesByUser retrieves all entries for a specific user
// Used when user is authenticated - only show their own entries
func GetEntriesByUser(userID int64) ([]map[string]interface{}, error) {
	query := `SELECT id, user_id, text, mood, category, created_at 
	          FROM entries 
	          WHERE user_id = ?
	          ORDER BY created_at DESC`

	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []map[string]interface{}

	for rows.Next() {
		var id, userIDResult int64
		var text, category, createdAt string
		var mood int

		err := rows.Scan(&id, &userIDResult, &text, &mood, &category, &createdAt)
		if err != nil {
			return nil, err
		}

		entry := map[string]interface{}{
			"id":         id,
			"user_id":    userIDResult,
			"text":       text,
			"mood":       mood,
			"category":   category,
			"created_at": createdAt,
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// CreateUser inserts a new user into the database
// Takes email and hashed password (NOT plain password!)
func CreateUser(email string, passwordHash string) (int64, error) {
	query := `INSERT INTO users (email, password_hash) VALUES (?, ?)`

	result, err := DB.Exec(query, email, passwordHash)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("User created with ID: %d", id)
	return id, nil
}

// GetUserByEmail retrieves a user by their email
// Returns user_id and password_hash for login verification
func GetUserByEmail(email string) (int64, string, error) {
	query := `SELECT id, password_hash FROM users WHERE email = ?`

	var userID int64
	var passwordHash string

	err := DB.QueryRow(query, email).Scan(&userID, &passwordHash)
	if err != nil {
		return 0, "", err
	}

	return userID, passwordHash, nil
}

// CloseDB closes the database connection
func CloseDB() {
	// Why Close DB Connection When Server Closes?
	// Great question! Let me explain with real-world consequences:

	// What Happens If You DON'T Close?
	// Problem 1: File Locks
	// SQLite locks the data.db file when connected.

	// Without closing:
	// Server crashes → db.db still "locked"
	// Try to restart server → ERROR: database is locked
	// You'd have to manually kill processes or restart your computer.

	// Problem 2: Uncommitted Writes
	// Database might be in the middle of writing data.

	// Without closing:
	// User submits entry → data being written
	// Server crashes
	// data.db corrupted (partial write)
	// With proper close:
	// 	Server shutting down → DB.Close() called
	// All pending writes finish
	// File closes cleanly

	// 	Problem 3: Resource Leaks
	// Database connections use system resources (memory, file handles).
	// 	Start server → uses 10MB RAM for DB
	// Stop server (no close) → 10MB still allocated
	// Start again → another 10MB
	// ...
	// Eventually: Out of memory
	if DB != nil {
		DB.Close()
	}
}
