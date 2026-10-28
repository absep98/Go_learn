package db

/*
=== DB PACKAGE TESTS ===

WHAT WE ARE TESTING:
CRUD operations on Postgres — InsertEntry, GetEntryByID, GetAllEntries, UpdateEntry, DeleteEntry.
All tests run against a real Postgres database configured via env vars — no production data touched.

KEY CONCEPTS USED:

1. TestMain — PACKAGE-LEVEL SETUP
   func TestMain(m *testing.M) runs ONCE before all tests in this package.
   We use it to connect to Postgres, truncate tables for a clean state, and run all tests.
   m.Run() is where all TestXxx functions actually execute.
   Without calling m.Run(), no tests run at all.

   If DB_HOST is not set (e.g. local dev without Postgres), the tests are skipped entirely
   via os.Exit(0). This avoids false failures when Postgres isn't available.

   Why not just call InitDB() inside each test?
   Each call would open a new connection pool — slow and redundant.
   TestMain runs once, all tests share the same DB instance.

2. WHY TRUNCATE BEFORE TESTS
   Unlike SQLite :memory: (which starts empty every run), Postgres persists data between runs.
   TRUNCATE TABLE entries, users RESTART IDENTITY CASCADE wipes all rows and resets
   auto-increment IDs to 1 — ensuring a known, clean state before every test run.
   CASCADE handles the foreign key from entries → users.

3. SHARED DB STATE ACROSS TESTS
   All tests share the same DB instance from TestMain.
   Rows inserted in TestInsertEntry are still there in TestGetAllEntries.
   That's why TestGetAllEntries uses len(entries) >= 2 instead of == 2.

4. ROWS AFFECTED vs ID
   InsertEntry returns (int64, error) — the id of the new row
   UpdateEntry and DeleteEntry return (int64, error) — rows affected (how many rows changed)
   Asserting rowsAffected == 1 confirms exactly one row was modified.

Q: Why test the NotFound case for GetEntryByID?
A: The happy path (found) is one code path. NotFound is a different code path.
   Without this test, a bug in the error-handling branch would go undetected.
   Good tests cover both success AND failure paths.

Q: Why use int(id) when calling GetEntryByID?
A: InsertEntry returns int64. GetEntryByID expects int.
   Go doesn't implicitly convert between numeric types — you must be explicit.
*/

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"personal-analytics-backend/internal/testdb"

	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	// A database of our own ("<DB_NAME>_db"): the handlers package runs in
	// parallel and truncates its own tables — it must never touch ours.
	dsn, migrateDSN, ok, err := testdb.DSNs("db")
	if !ok {
		fmt.Println("DB_HOST not set, skipping DB tests")
		os.Exit(0)
	}
	if err != nil {
		panic("failed to prepare test DB: " + err.Error())
	}

	testDSN, testMigrateDSN = dsn, migrateDSN

	err = InitDB(dsn, migrateDSN)
	if err != nil {
		panic("failed to init test DB: " + err.Error())
	}

	_, err = DB.Exec("TRUNCATE TABLE entries, users RESTART IDENTITY CASCADE")

	if err != nil {
		panic("failed to clean test DB: " + err.Error())
	}
	// run all tests
	os.Exit(m.Run())
}

// store is the real Postgres-backed implementation exercised by every test below.
var store = &PostgresStore{}

// resetTables gives a test that asserts EXACT row counts a clean slate.
// TestMain truncates only once per process, so without this, `go test -count=N`
// (or any re-run in the same process) accumulates rows and those tests fail.
func resetTables(t *testing.T) {
	t.Helper()
	if _, err := DB.Exec("TRUNCATE TABLE entries, users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("reset tables: %v", err)
	}
}

// Connection strings built in TestMain, reused by tests that call InitDB again.
var testDSN, testMigrateDSN string

// InitDB used to log the full DSN — including the database password — at INFO
// level on every start, so the credential ended up in container logs and any
// log shipper. It should log only non-secret connection details.
func TestInitDB_DoesNotLogPassword(t *testing.T) {
	var logs bytes.Buffer
	prevLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(prevLogger)

	prevDB := DB
	if err := InitDB(testDSN, testMigrateDSN); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	prevDB.Close()

	pw := os.Getenv("DB_PASSWORD")
	for _, leak := range []string{"password=" + pw, ":" + pw + "@"} {
		if strings.Contains(logs.String(), leak) {
			t.Fatalf("database password written to logs (found %q):\n%s", leak, logs.String())
		}
	}
}

func TestInsertEntry(t *testing.T) {

	id, err := store.InsertEntry(context.Background(), 1, "feeling good", 8, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

}

func TestGetEntryByID(t *testing.T) {
	id, err := store.InsertEntry(context.Background(), 1, "feeling good", 8, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	entry, err := store.GetEntryByID(context.Background(), int(id), 1)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if entry.Text != "feeling good" {
		t.Errorf("expected text 'feeling good', got %v", entry.Text)
	}

}

func TestGetEntryByID_NotFound(t *testing.T) {

	entry, err := store.GetEntryByID(context.Background(), 9999, 1)

	if err == nil {
		t.Errorf("expected error for missing entry, got nil")
	}

	if entry != nil {
		t.Errorf("expected nil entry, got %v", entry)
	}
}

func TestGetAllEntries(t *testing.T) {

	id, err := store.InsertEntry(context.Background(), 2, "feeling good", 8, "mood")
	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	id, err = store.InsertEntry(context.Background(), 10, "feeling bad", 5, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	entry, err := store.GetAllEntries(context.Background())

	if err != nil {
		t.Errorf("failed to get all entries %v", err)
	}

	if len(entry) < 2 {
		t.Errorf("expected at least 2 entries, got %d", len(entry))
	}
}

func TestUpdateEntry(t *testing.T) {

	id, err := store.InsertEntry(context.Background(), 100, "Feeling good feeling alright", 10, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %v", err)
	}

	if err != nil {
		t.Errorf("expected nil error on insert, got %v", err)
	}

	rowsAffected, err := store.UpdateEntry(context.Background(), int(id), 100, "Not going well", 7, "mood")

	if rowsAffected != 1 {
		t.Errorf("expected 1 row updated, got %d", rowsAffected)
	}

	if err != nil {
		t.Errorf("expected nil error on update, got %v", err)
	}

}

func TestDeleteEntry(t *testing.T) {

	id, err := store.InsertEntry(context.Background(), 100, "Feeling good, feeling alright", 10, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %v", err)
	}

	if err != nil {
		t.Errorf("expected nil error on insert, got %v", err)
	}

	rowAffected, err := store.DeleteEntry(context.Background(), int(id), 100)

	if rowAffected != 1 {
		t.Errorf("expected 1 row updated, got %d", rowAffected)
	}

	if err != nil {
		t.Errorf("expected nil error on delete, got %v", err)
	}

}

func TestGetAllEntriesPaginated(t *testing.T) {

	// Insert 3 entries for userID=1
	store.InsertEntry(context.Background(), 1, "entry 1", 5, "mood") //nolint:errcheck
	store.InsertEntry(context.Background(), 1, "entry 2", 6, "mood") //nolint:errcheck
	store.InsertEntry(context.Background(), 1, "entry 3", 7, "mood") //nolint:errcheck

	// Page 1: limit=2, offset=0 → should return 2 entries
	entries, err := store.GetAllEntriesPaginated(context.Background(), 1, 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	// Page 2: limit=2, offset=2 → should return at least 1 entry
	entries2, err := store.GetAllEntriesPaginated(context.Background(), 1, 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries2) < 1 {
		t.Errorf("expected at least 1 entry on page 2, got %d", len(entries2))
	}
}

func TestGetEntryCount(t *testing.T) {

	store.InsertEntry(context.Background(), 200, "test", 5, "mood") //nolint:errcheck
	store.InsertEntry(context.Background(), 200, "test", 5, "mood") //nolint:errcheck

	count, err := store.GetEntryCount(context.Background(), 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count < 2 {
		t.Errorf("expected at least 2, got %d", count)
	}

	// userID with no entries should return 0
	count2, err := store.GetEntryCount(context.Background(), 9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count2 != 0 {
		t.Errorf("expected 0, got %d", count2)
	}

}

func TestGetFilteredByCategory(t *testing.T) {
	resetTables(t)

	store.InsertEntry(context.Background(), 300, "test", 5, "mood") //nolint:errcheck
	store.InsertEntry(context.Background(), 300, "test", 5, "mood") //nolint:errcheck
	store.InsertEntry(context.Background(), 300, "test", 5, "work") //nolint:errcheck

	entries, err := store.GetEntriesFiltered(context.Background(), 300, "mood", 0, 10, 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Filtered operation didn't returned the correct entries data")
	}
}

func TestGetFilteredByMoodMin(t *testing.T) {
	resetTables(t)
	store.InsertEntry(context.Background(), 400, "low mood", 5, "mood")  //nolint:errcheck
	store.InsertEntry(context.Background(), 400, "mid mood", 7, "mood")  //nolint:errcheck
	store.InsertEntry(context.Background(), 400, "high mood", 9, "mood") //nolint:errcheck

	entries, err := store.GetEntriesFiltered(context.Background(), 400, "", 7, 10, 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries with mood >= 7 and got %d", len(entries))
	}
}
