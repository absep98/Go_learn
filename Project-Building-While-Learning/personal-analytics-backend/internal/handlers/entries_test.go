package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"personal-analytics-backend/internal/db"
	"personal-analytics-backend/internal/redis"
	"strings"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestMain(m *testing.M) {
	host := os.Getenv("DB_HOST")

	if host != "" {
		// DB env vars present - set up real DB for integrations tests
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host,
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		)

		migrateDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), host, os.Getenv("DB_PORT"), os.Getenv("DB_NAME"),
		)

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
	json.NewDecoder(rec.Body).Decode(&resp)

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}

	if resp.ID != 1 {
		t.Errorf("expected ID 1, got %d", resp.ID)
	}
}
