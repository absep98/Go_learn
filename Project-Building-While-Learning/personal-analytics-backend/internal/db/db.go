package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"personal-analytics-backend/internal/cache"
	"personal-analytics-backend/internal/models"
	"strconv"
	"time"

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
	_ "github.com/lib/pq"
)

// DB is the global database connection
// It's the middleman between your Go code and the database file.
// "Create a variable called DB that will hold the connection to the database."
// This is like having a phone line open to the database. Once connected,
//  you use this DB variable to talk to the database from anywhere in your code.

var DB *sql.DB

// InitDB initializes the database connection and creates tables
func InitDB(dsn string) error {
	var err error

	// Open Postgres database using the connection string (DSN)
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		return err
	}

	slog.Info("Database connected", "path", dsn)

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
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);`

	_, err := DB.Exec(usersTable)
	if err != nil {
		return err
	}

	// Entries table - stores mood/activity data
	entriesTable := `
	CREATE TABLE IF NOT EXISTS entries (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		text TEXT,
		mood INTEGER,
		category TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);`

	_, err = DB.Exec(entriesTable)
	if err != nil {
		return err
	}

	slog.Info("Database tables created")
	return nil
}

// InsertEntry inserts a new entry into the database
// Puts new data INTO the database (like adding a new row to an Excel sheet)
// Takes 5 inputs: userID (which user), text (what they wrote), mood (their mood score), tags (list of tags), category (entry type)
func InsertEntry(ctx context.Context, userID int, text string, mood int, category string) (int64, error) {
	// The ? marks are placeholders (like blanks in a form)
	query := `INSERT INTO entries (user_id, text, mood, category) VALUES ($1, $2, $3, $4) RETURNING id`

	var id int64
	err := DB.QueryRowContext(ctx, query, userID, text, mood, category).Scan(&id)
	// "Execute the query and fill in the ? marks with actual values."
	if err != nil {
		return 0, err
	}

	slog.Debug("Entry inserted", "entry_id", id)
	return id, nil
}

// GetEntryById retrieves the entry from database for a particular ID
func GetEntryByID(ctx context.Context, entryID int, userId int64) (*models.Entry, error) {

	query := `SELECT id, user_id, text, mood, category, created_at
				FROM entries
				WHERE id = $1 and user_id = $2
			`

	var entry models.Entry

	row := DB.QueryRowContext(ctx, query, entryID, userId)

	err := row.Scan(&entry.ID, &entry.UserID, &entry.Text, &entry.Mood, &entry.Category, &entry.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &entry, nil
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
func GetAllEntries(ctx context.Context) ([]models.Entry, error) {
	query := `SELECT id, user_id, text, mood, category, created_at FROM entries ORDER BY created_at DESC`

	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	// "When this function ends, close the rows cursor"
	defer rows.Close()

	var entries []models.Entry

	for rows.Next() {
		var entry models.Entry

		err := rows.Scan(&entry.ID, &entry.UserID, &entry.Text, &entry.Mood, &entry.Category, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetEntriesByUser retrieves all entries for a specific user
// Used when user is authenticated - only show their own entries
func GetEntriesByUser(ctx context.Context, userID int64) ([]models.Entry, error) {
	query := `SELECT id, user_id, text, mood, category, created_at
	          FROM entries
	          WHERE user_id = $1
	          ORDER BY created_at DESC`

	rows, err := DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.Entry

	for rows.Next() {

		var entry models.Entry

		err := rows.Scan(&entry.ID, &entry.UserID, &entry.Text, &entry.Mood, &entry.Category, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func GetAllEntriesPaginated(ctx context.Context, userID int64, limit int, offset int) ([]models.Entry, error) {

	if limit < 1 {
		limit = 10
	}
	queryData := `SELECT id, user_id, text, mood, category, created_at
		FROM entries WHERE
		user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := DB.QueryContext(ctx, queryData, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var entries []models.Entry

	for rows.Next() {

		var entry models.Entry

		err := rows.Scan(&entry.ID, &entry.UserID, &entry.Text, &entry.Mood, &entry.Category, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func GetEntryCount(ctx context.Context, userID int64) (int, error) {

	query := `SELECT COUNT(*) FROM entries WHERE user_id = $1`

	var count int
	err := DB.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetEntriesByUserPaginated allow users to paginate through entries instead of getting all at once.
func GetEntriesByUserPaginated(ctx context.Context, userId int, page int, limit int) (entries []models.Entry, total int) {

	// 1. Build cache key
	cacheKey := fmt.Sprintf("count:user:%d", userId)

	// 2. Try to get from cache
	if cachedCount, found := cache.Get(cacheKey); found {
		total, _ = strconv.Atoi(cachedCount) // Convert string to int (Redis stores strings)
	} else {
		// 3. Cache miss: query database
		queryForCount := `SELECT COUNT(*) FROM entries WHERE user_id = $1`
		err := DB.QueryRowContext(ctx, queryForCount, userId).Scan(&total)

		if err != nil {
			return nil, 0
		}

		// 4. Store in cache for 60 seconds
		cache.Set(cacheKey, total, 60*time.Second)
	}

	queryForEntries := `SELECT id, user_id, text, mood, category, created_at
	 FROM entries WHERE
	 user_id = $1
	 ORDER BY created_at DESC
	 LIMIT $2 OFFset $3
	 `
	rows, err := DB.QueryContext(ctx, queryForEntries, userId, limit, (page-1)*limit)

	if err != nil {
		return nil, 0
	}

	defer rows.Close()

	for rows.Next() {

		var entry models.Entry

		err := rows.Scan(&entry.ID, &entry.UserID, &entry.Text, &entry.Mood, &entry.Category, &entry.CreatedAt)
		if err != nil {
			return nil, 0
		}

		entries = append(entries, entry)
	}
	return entries, total
}

// CreateUser inserts a new user into the database
// Takes email and hashed password (NOT plain password!)
func CreateUser(ctx context.Context, email string, passwordHash string) (int64, error) {
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`

	var id int64

	err := DB.QueryRowContext(ctx, query, email, passwordHash).Scan(&id)
	if err != nil {
		return 0, err
	}

	slog.Debug("User created", "user_id", id)
	return id, nil
}

// GetUserByEmail retrieves a user by their email
// Returns user_id and password_hash for login verification
func GetUserByEmail(ctx context.Context, email string) (int64, string, error) {
	query := `SELECT id, password_hash FROM users WHERE email = $1`

	var userID int64
	var passwordHash string

	err := DB.QueryRowContext(ctx, query, email).Scan(&userID, &passwordHash)
	if err != nil {
		return 0, "", err
	}

	return userID, passwordHash, nil
}

// Update Entry
// Updates an entry only if it belongs to the authenticated user
func UpdateEntry(ctx context.Context, entryId int, userID int64, text string, mood int, category string) (int64, error) {

	query := `UPDATE entries SET text = $1, mood = $2, category = $3 WHERE id = $4 AND user_id = $5`

	// Parameters must match placeholder order: text, mood, category, id, user_id

	result, err := DB.ExecContext(ctx, query, text, mood, category, entryId, userID)
	if err != nil {
		return 0, err
	}

	// Check if any row was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	// Returns 0 if entry not found OR doesn't belong to user
	return rowsAffected, nil
}

func DeleteEntry(ctx context.Context, entryId int, userID int64) (int64, error) {

	query := `DELETE FROM entries WHERE id = $1 and user_id = $2`

	result, err := DB.ExecContext(ctx, query, entryId, userID)
	if err != nil {
		return 0, err
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowAffected, nil
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
