package handlers

import (
	"context"
	"personal-analytics-backend/internal/models"
	"time"
)

// Store is the interface for all database operations
// any struct that has these methods can be used as a Store
type Store interface {
	InsertEntry(ctx context.Context, userID int, text string, mood int, category string) (int64, error)
	GetEntryByID(ctx context.Context, entryID int, userID int64) (*models.Entry, error)
	GetEntriesByUser(ctx context.Context, userID int64) ([]models.Entry, error)
	GetAllEntriesPaginated(ctx context.Context, userID int64, limit int, offset int) ([]models.Entry, error)
	GetEntryCount(ctx context.Context, userID int64) (int, error)
	GetEntriesFiltered(ctx context.Context, userID int64, category string, moodMin int, limit int, offset int) ([]models.Entry, error)
	UpdateEntry(ctx context.Context, entryID int, userID int64, text string, mood int, category string) (int64, error)
	DeleteEntry(ctx context.Context, entryID int, userID int64) (int64, error)
	CreateUser(ctx context.Context, email string, passwordHash string) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (int64, string, error)
}

type Handler struct {
	Store     Store
	Cache     Cache
	JWTSecret string
}

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}
