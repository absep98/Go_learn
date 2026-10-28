package handlers

/*
=== REFRESH TOKEN ROTATION TESTS ===

A refresh token is single-use: /refresh must consume it and issue a new one.

The original handler did GET then DEL as two separate cache calls. Each Redis
command is atomic, but the pair is not — N concurrent requests replaying the
same stolen token can all GET it before any of them DELs it, and every one of
them is issued a fresh access + refresh token (the attacker now holds N live
sessions). The fix reads-and-deletes in ONE command (Redis GETDEL, 6.2+).

memCache mimics Redis semantics: every single command is atomic (mutex), but
nothing makes a sequence of commands atomic. readBarrier makes the race
deterministic: every concurrent Get must finish reading before any of them is
allowed to continue to Del — the worst-case interleaving, every time.
*/

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type memCache struct {
	mu          sync.Mutex
	data        map[string]string
	readBarrier *sync.WaitGroup // nil = no barrier
}

func newMemCache() *memCache { return &memCache{data: map[string]string{}} }

var errCacheMiss = errors.New("cache miss")

func (m *memCache) Set(ctx context.Context, key string, value interface{}, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value.(string)
	return nil
}

func (m *memCache) Get(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	v, ok := m.data[key]
	m.mu.Unlock()

	if m.readBarrier != nil {
		m.readBarrier.Done() // this reader has read...
		m.readBarrier.Wait() // ...now wait until every concurrent reader has too
	}

	if !ok {
		return "", errCacheMiss
	}
	return v, nil
}

func (m *memCache) Del(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *memCache) GetDel(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	if !ok {
		return "", errCacheMiss
	}
	delete(m.data, key)
	return v, nil
}

func refreshRequest(token string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"`+token+`"}`))
}

func TestRefresh_ConcurrentReplayOfOneTokenSucceedsOnce(t *testing.T) {
	const replays = 10
	cache := newMemCache()
	cache.data["refresh_token:stolen-token"] = "7"
	cache.readBarrier = &sync.WaitGroup{}
	cache.readBarrier.Add(replays)
	h := &Handler{Cache: cache, JWTSecret: "test-secret"}

	var wg sync.WaitGroup
	codes := make(chan int, replays)
	for i := 0; i < replays; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			h.Refresh(rec, refreshRequest("stolen-token"))
			codes <- rec.Code
		}()
	}
	wg.Wait()
	close(codes)

	succeeded := 0
	for code := range codes {
		if code == http.StatusOK {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Fatalf("single-use refresh token was accepted %d times out of %d concurrent replays; want exactly 1", succeeded, replays)
	}
}

func TestRefresh_RotatesTokenAndRejectsTheOldOne(t *testing.T) {
	cache := newMemCache()
	cache.data["refresh_token:first"] = "7"
	h := &Handler{Cache: cache, JWTSecret: "test-secret"}

	rec := httptest.NewRecorder()
	h.Refresh(rec, refreshRequest("first"))
	if rec.Code != http.StatusOK {
		t.Fatalf("first refresh: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var resp LoginResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" || resp.RefreshToken == "first" {
		t.Fatalf("expected a new access token and a rotated refresh token, got %+v", resp)
	}

	rec = httptest.NewRecorder()
	h.Refresh(rec, refreshRequest("first"))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("reusing the old refresh token: expected 401, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.Refresh(rec, refreshRequest(resp.RefreshToken))
	if rec.Code != http.StatusOK {
		t.Errorf("using the rotated refresh token: expected 200, got %d", rec.Code)
	}
}

func logoutRequest(token string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(`{"refresh_token":"`+token+`"}`))
}

func TestLogout_RevokesRefreshToken(t *testing.T) {
	cache := newMemCache()
	cache.data["refresh_token:session-1"] = "7"
	h := &Handler{Cache: cache, JWTSecret: "test-secret"}

	rec := httptest.NewRecorder()
	h.Logout(rec, logoutRequest("session-1"))
	if rec.Code != http.StatusOK {
		t.Fatalf("logout: expected 200, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.Refresh(rec, refreshRequest("session-1"))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("refresh after logout: expected 401, got %d", rec.Code)
	}
}

func TestLogout_UnknownTokenIsUnauthorized(t *testing.T) {
	h := &Handler{Cache: newMemCache(), JWTSecret: "test-secret"}

	rec := httptest.NewRecorder()
	h.Logout(rec, logoutRequest("never-issued"))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
