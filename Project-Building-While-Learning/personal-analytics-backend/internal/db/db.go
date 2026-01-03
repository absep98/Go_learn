package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// DB is the global database connection
var DB *sql.DB

// InitDB initializes the database connection and creates tables
func InitDB(dbPath string) error {
	var err error

	// Open SQLite database (creates file if not exists)
	DB, err = sql.Open("sqlite", dbPath)
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
	// Entries table - stores mood/activity data
	entriesTable := `
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		text TEXT,
		mood INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(entriesTable)
	if err != nil {
		return err
	}

	log.Println("Tables created successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
