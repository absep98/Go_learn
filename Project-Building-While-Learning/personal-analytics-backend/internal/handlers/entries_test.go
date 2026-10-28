package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"personal-analytics-backend/internal/db"
	"personal-analytics-backend/internal/redis"
	"personal-analytics-backend/internal/testdb"
	"strings"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestMain(m *testing.M) {
	// DB env vars present - set up real DB for integrations tests, in a database
	// of our own ("<DB_NAME>_handlers") so the parallel db package can't interfere.
	dsn, migrateDSN, ok, err := testdb.DSNs("handlers")
	if err != nil {
		panic("failed to prepare test DB: " + err.Error())
	}

	if ok {
		err := db.InitDB(dsn, migrateDSN)
		if err != nil {
			panic("Failed to init test DB : " + err.Error())
		}

		_, err = db.DB.Exec("TRUNCATE TABLE entries, users RESTART IDENTITY CASCADE")

		if err != nil {
			panic("failed to clean test DB: " + err.Error())
		}

		redis.Client = goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})
	}
	os.Exit(m.Run())
}

func TestCreateEntry(t *testing.T) {
	h := &Handler{
		Store: &FakeStore{InsertEntryFn: func() (int64, error) {
			return 1, nil
		}},
		Cache: &FakeCache{},
	}

	body := strings.NewReader(`{"text":"feeling good","mood":8,"category":"mood"}`)
	req := httptest.NewRequest(http.MethodPost, "/entries", body)
	ctx := context.WithValue(req.Context(), userIDKey, int64(1))
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.CreateEntry(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var resp CreateEntryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}

	if resp.ID != 1 {
		t.Errorf("expected ID 1, got %d", resp.ID)
	}
}
