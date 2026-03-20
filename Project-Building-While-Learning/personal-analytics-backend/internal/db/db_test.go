package db

/*
=== DB PACKAGE TESTS ===

WHAT WE ARE TESTING:
CRUD operations on SQLite — InsertEntry, GetEntryByID, GetAllEntries, UpdateEntry, DeleteEntry.
All tests run against an in-memory SQLite database (":memory:") — no file created, no production data touched.

KEY CONCEPTS USED:

1. TestMain — PACKAGE-LEVEL SETUP
   func TestMain(m *testing.M) runs ONCE before all tests in this package.
   We use it to initialize the in-memory DB and create tables.
   m.Run() is where all TestXxx functions actually execute.
   Without calling m.Run(), no tests run at all.

   Why not just call InitDB() inside each test?
   Each call would create a new DB and tables — slow and redundant.
   TestMain runs once, all tests share the same DB instance.

2. IN-MEMORY SQLITE (":memory:")
   InitDB(":memory:") creates a DB that lives only in RAM.
   It's empty at start, fast, and disappears when tests finish.
   Why not use the production data.db?
   - Tests insert/delete rows — would corrupt real user data
   - Tests need a known state — production DB has unknown existing rows
   - In-memory DB is isolated and reproducible every run

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
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// setup : init in-memory db

	err := InitDB(":memory:")

	if err != nil {
		panic("failed to init test DB: " + err.Error())
	}

	// run all tests
	os.Exit(m.Run())
}

func TestInsertEntry(t *testing.T) {

	id, err := InsertEntry(1, "feeling good", 8, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

}

func TestGetEntryByID(t *testing.T) {
	id, err := InsertEntry(1, "feeling good", 8, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	entry, err := GetEntryByID(int(id), 1)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if entry["text"].(string) != "feeling good" {
		t.Errorf("expected text 'feeling good', got %v", entry["text"])
	}

}

func TestGetEntryByID_NotFound(t *testing.T) {

	entry, err := GetEntryByID(9999, 1)

	if err == nil {
		t.Errorf("expected error for missing entry, got nil")
	}

	if entry != nil {
		t.Errorf("expected nil entry, got %v", entry)
	}
}

func TestGetAllEntries(t *testing.T) {

	id, err := InsertEntry(2, "feeling good", 8, "mood")
	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	id, err = InsertEntry(10, "feeling bad", 5, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	entry, err := GetAllEntries()

	if err != nil {
		t.Errorf("failed to get all entries %v", err)
	}

	if len(entry) < 2 {
		t.Errorf("expected at least 2 entries, got %d", len(entry))
	}
}

func TestUpdateEntry(t *testing.T) {

	id, err := InsertEntry(100, "Feeling good feeling alright", 10, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %v", err)
	}

	if err != nil {
		t.Errorf("expected nil error on insert, got %v", err)
	}

	rowsAffected, err := UpdateEntry(int(id), 100, "Not going well", 7, "mood")

	if rowsAffected != 1 {
		t.Errorf("expected 1 row updated, got %d", rowsAffected)
	}

	if err != nil {
		t.Errorf("expected nil error on update, got %v", err)
	}

}

func TestDeleteEntry(t *testing.T) {

	id, err := InsertEntry(100, "Feeling good, feeling alright", 10, "mood")

	if id <= 0 {
		t.Errorf("expected id > 0, got %v", err)
	}

	if err != nil {
		t.Errorf("expected nil error on insert, got %v", err)
	}

	rowAffected, err := DeleteEntry(int(id), 100)

	if rowAffected != 1 {
		t.Errorf("expected 1 row updated, got %d", rowAffected)
	}

	if err != nil {
		t.Errorf("expected nil error on delete, got %v", err)
	}

}
