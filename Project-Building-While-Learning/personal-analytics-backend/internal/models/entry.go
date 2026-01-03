package models

import (
	"time"

	"personal-analytics-backend/internal/db"
)

// Entry represents a mood/activity log entry
type Entry struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Text      string    `json:"text"`
	Mood      int       `json:"mood"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateEntry inserts a new entry into the database
func CreateEntry(userID int64, text string, mood int) (*Entry, error) {
	query := `INSERT INTO entries (user_id, text, mood) VALUES (?, ?, ?)`

	result, err := db.DB.Exec(query, userID, text, mood)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Fetch the created entry to get the timestamp
	return GetEntryByID(id)
}

// GetEntryByID retrieves a single entry by its ID
func GetEntryByID(id int64) (*Entry, error) {
	query := `SELECT id, user_id, text, mood, created_at FROM entries WHERE id = ?`

	entry := &Entry{}
	err := db.DB.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.UserID,
		&entry.Text,
		&entry.Mood,
		&entry.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// GetAllEntries retrieves all entries (for testing)
func GetAllEntries() ([]Entry, error) {
	query := `SELECT id, user_id, text, mood, created_at FROM entries ORDER BY created_at DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Text,
			&entry.Mood,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetEntriesByUserID retrieves all entries for a specific user
func GetEntriesByUserID(userID int64) ([]Entry, error) {
	query := `SELECT id, user_id, text, mood, created_at FROM entries WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Text,
			&entry.Mood,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
