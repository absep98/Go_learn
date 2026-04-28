package handlers

import (
	"context"
	"fmt"
	"personal-analytics-backend/internal/models"
	"time"

	"github.com/lib/pq"
)

type FakeStore struct {
	//  control what it returns in each test
	InsertEntryFn func() (int64, error)
	users         map[string]struct {
		id           int64
		passwordHash string
	}
}

type FakeCache struct{}

func (f *FakeStore) InsertEntry(ctx context.Context, userID int, text string, mood int, category string) (int64, error) {
	return f.InsertEntryFn()
}

func (f *FakeStore) GetEntryByID(ctx context.Context, entryID int, userID int64) (*models.Entry, error) {
	return nil, nil
}

func (f *FakeStore) GetEntriesByUser(ctx context.Context, userID int64) ([]models.Entry, error) {
	return nil, nil
}

func (f *FakeStore) GetAllEntriesPaginated(ctx context.Context, userID int64, limit int, offset int) ([]models.Entry, error) {
	return nil, nil
}

func (f *FakeStore) GetEntryCount(ctx context.Context, userID int64) (int, error) {
	return 0, nil
}

func (f *FakeStore) GetEntriesFiltered(ctx context.Context, userID int64, category string, moodMin int, limit int, offset int) ([]models.Entry, error) {
	return nil, nil
}

func (f *FakeStore) UpdateEntry(ctx context.Context, entryID int, userID int64, text string, mood int, category string) (int64, error) {
	return 0, nil
}

func (f *FakeStore) DeleteEntry(ctx context.Context, entryID int, userID int64) (int64, error) {
	return 0, nil
}

func (f *FakeStore) CreateUser(ctx context.Context, email string, passwordHash string) (int64, error) {
	if f.users == nil {
		f.users = make(map[string]struct {
			id           int64
			passwordHash string
		})
	}

	if _, exists := f.users[email]; exists {
		return 0, &pq.Error{Code: "23505"} // Duplicate email
	}

	id := int64(len(f.users) + 1)
	f.users[email] = struct {
		id           int64
		passwordHash string
	}{id, passwordHash}
	return id, nil
}

func (f *FakeStore) GetUserByEmail(ctx context.Context, email string) (int64, string, error) {
	if user, ok := f.users[email]; ok {
		return user.id, user.passwordHash, nil
	}

	return 0, "", fmt.Errorf("user not found")
}

func (f *FakeCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (f *FakeCache) Get(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("cache miss")
}

func (f *FakeCache) Del(ctx context.Context, key string) error {
	return nil
}
