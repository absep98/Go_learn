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
	"context"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	host := os.Getenv("DB_HOST")

	if host == "" {
		fmt.Println("DB_HOST not set, skipping DB tests")
		os.Exit(0)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	err := InitDB(dsn)
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

func TestInsertEntry(t *testing.T) {

	id, err := InsertEntry(context.Background(), 1, "feeling good", 8, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

}

func TestGetEntryByID(t *testing.T) {
	id, err := InsertEntry(context.Background(), 1, "feeling good", 8, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	entry, err := GetEntryByID(context.Background(), int(id), 1)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if entry.Text != "feeling good" {
		t.Errorf("expected text 'feeling good', got %v", entry.Text)
	}

}

func TestGetEntryByID_NotFound(t *testing.T) {

	entry, err := GetEntryByID(context.Background(), 9999, 1)

	if err == nil {
		t.Errorf("expected error for missing entry, got nil")
	}

	if entry != nil {
		t.Errorf("expected nil entry, got %v", entry)
	}
}

func TestGetAllEntries(t *testing.T) {

	id, err := InsertEntry(context.Background(), 2, "feeling good", 8, "mood")
	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	id, err = InsertEntry(context.Background(), 10, "feeling bad", 5, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	entry, err := GetAllEntries(context.Background())

	if err != nil {
		t.Errorf("failed to get all entries %v", err)
	}

	if len(entry) < 2 {
		t.Errorf("expected at least 2 entries, got %d", len(entry))
	}
}

func TestUpdateEntry(t *testing.T) {

	id, err := InsertEntry(context.Background(), 100, "Feeling good feeling alright", 10, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %v", err)
	}

	if err != nil {
		t.Errorf("expected nil error on insert, got %v", err)
	}

	rowsAffected, err := UpdateEntry(context.Background(), int(id), 100, "Not going well", 7, "mood")

	if rowsAffected != 1 {
		t.Errorf("expected 1 row updated, got %d", rowsAffected)
	}

	if err != nil {
		t.Errorf("expected nil error on update, got %v", err)
	}

}

func TestDeleteEntry(t *testing.T) {

	id, err := InsertEntry(context.Background(), 100, "Feeling good, feeling alright", 10, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %v", err)
	}

	if err != nil {
		t.Errorf("expected nil error on insert, got %v", err)
	}

	rowAffected, err := DeleteEntry(context.Background(), int(id), 100)

	if rowAffected != 1 {
		t.Errorf("expected 1 row updated, got %d", rowAffected)
	}

	if err != nil {
		t.Errorf("expected nil error on delete, got %v", err)
	}

}
