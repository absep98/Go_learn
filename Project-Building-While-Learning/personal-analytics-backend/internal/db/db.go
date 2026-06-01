package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"personal-analytics-backend/internal/models"
	"encoding/json"
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

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// DB is the global database connection
// It's the middleman between your Go code and the database file.
// "Create a variable called DB that will hold the connection to the database."
// This is like having a phone line open to the database. Once connected,
//  you use this DB variable to talk to the database from anywhere in your code.

var DB *sql.DB

// InitDB initializes the database connection and creates tables
func InitDB(dsn string, migrateDSN string) error {
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
	// Run Migrations now instead of creating tables
	err = RunMigrations(migrateDSN)
	if err != nil {
		return err
	}
	slog.Info("Database connected", "path", dsn)

	return nil
}

// Run Migrations
func RunMigrations(dsn string) error {
	m, err := migrate.New(
		"file://internal/db/migrations",
		dsn,
	)

	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}

	slog.Info("Database migrations applied")
	return nil
}

type PostgresStore struct{}

// InsertEntry inserts a new entry into the database
// Puts new data INTO the database (like adding a new row to an Excel sheet)
// Takes 5 inputs: userID (which user), text (what they wrote), mood (their mood score), tags (list of tags), category (entry type)
func (p *PostgresStore) InsertEntry(ctx context.Context, userID int, text string, mood int, category string) (int64, error) {
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
func (p *PostgresStore) GetEntryByID(ctx context.Context, entryID int, userId int64) (*models.Entry, error) {

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
func (p *PostgresStore) GetAllEntries(ctx context.Context) ([]models.Entry, error) {
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
func (p *PostgresStore) GetEntriesByUser(ctx context.Context, userID int64) ([]models.Entry, error) {
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

func (p *PostgresStore) GetAllEntriesPaginated(ctx context.Context, userID int64, limit int, offset int) ([]models.Entry, error) {

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

func (p *PostgresStore) GetEntryCount(ctx context.Context, userID int64) (int, error) {

	query := `SELECT COUNT(*) FROM entries WHERE user_id = $1`

	var count int
	err := DB.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (p *PostgresStore) GetEntriesFiltered(ctx context.Context, userID int64, category string, moodMin int, limit int, offset int) ([]models.Entry, error) {

	query := `SELECT id, user_id, text, mood, category, created_at FROM entries WHERE user_id = $1`
	args := []interface{}{userID}
	paramCount := 1

	if category != "" {
		paramCount++
		query += fmt.Sprintf(" AND category = $%d", paramCount)
		args = append(args, category)
	}

	if moodMin > 0 {
		paramCount++
		query += fmt.Sprintf(" AND mood >= $%d", paramCount)
		args = append(args, moodMin)
	}

	paramCount++
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", paramCount)
	args = append(args, limit)

	paramCount++
	query += fmt.Sprintf(" OFFSET $%d", paramCount)
	args = append(args, offset)

	rows, err := DB.QueryContext(ctx, query, args...)
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

// CreateUser inserts a new user into the database
// Takes email and hashed password (NOT plain password!)
func (p *PostgresStore) CreateUser(ctx context.Context, email string, passwordHash string) (int64, error) {
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
func (p *PostgresStore) GetUserByEmail(ctx context.Context, email string) (int64, string, error) {
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
func (p *PostgresStore) UpdateEntry(ctx context.Context, entryId int, userID int64, text string, mood int, category string) (int64, error) {

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

func (p *PostgresStore) DeleteEntry(ctx context.Context, entryId int, userID int64) (int64, error) {

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


func InsertFailedJob(job_type string, payload interface{}, errMsg string) error {

	query := `INSERT INTO failed_jobs (job_type, payload, error) VALUES($1, $2, $3) RETURNING id`
	
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var id int64
	err = DB.QueryRowContext(context.Background(), query, job_type, payloadJSON, errMsg).Scan(&id)
	// "Execute the query and fill in the ? marks with actual values."
	if err != nil {
		return err
	}

	slog.Debug("Failed job inserted", "entry_id", id)
	return nil
}


func GetFailedJobs() ([]models.FailedJob, error) {

	query := `SELECT id, job_type, payload, error, created_at FROM failed_jobs ORDER BY created_at DESC`

	rows, err := DB.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	failedJobs := make([]models.FailedJob, 0)

	for rows.Next() {
		var failedJob  models.FailedJob

		err := rows.Scan(&failedJob.ID, &failedJob.JobType, &failedJob.Payload, &failedJob.Error, &failedJob.CreatedAt)
		if err != nil {
			return nil, err
		}

		failedJobs  = append(failedJobs, failedJob)
	}
	return failedJobs, nil
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
