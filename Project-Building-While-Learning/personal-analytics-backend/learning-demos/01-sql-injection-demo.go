package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

/*
🔥 SQL INJECTION DEMO - SEE THE ATTACK HAPPEN

This file shows you EXACTLY why SQL injection is dangerous
and why parameterized queries save you.

Run this file to see:
1. SAFE code (parameterized queries with ?)
2. VULNERABLE code (string concatenation)
3. A REAL ATTACK that deletes data
*/

func main() {
	fmt.Println("🔐 SQL Injection Security Demo\n")

	// Create a test database
	db, err := sql.Open("sqlite", ":memory:") // in-memory for demo
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create test table
	setupDatabase(db)

	fmt.Println("============================================================")
	fmt.Println("PART 1: SAFE CODE (Parameterized Queries)")
	fmt.Println("============================================================")
	demonstrateSafeCode(db)

	fmt.Println("\n============================================================")
	fmt.Println("PART 2: VULNERABLE CODE (String Concatenation)")
	fmt.Println("============================================================")
	demonstrateVulnerableCode(db)

	fmt.Println("\n============================================================")
	fmt.Println("PART 3: THE ATTACK - Watch Your Data Get Destroyed")
	fmt.Println("============================================================")
	demonstrateAttack(db)
}

func setupDatabase(db *sql.DB) {
	// Create users table
	createTable := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		email TEXT
	)`

	_, err := db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	// Insert test data
	db.Exec("INSERT INTO users (username, email) VALUES ('alice', 'alice@example.com')")
	db.Exec("INSERT INTO users (username, email) VALUES ('bob', 'bob@example.com')")
	db.Exec("INSERT INTO users (username, email) VALUES ('charlie', 'charlie@example.com')")

	fmt.Println("✅ Database created with 3 users\n")
}

func demonstrateSafeCode(db *sql.DB) {
	fmt.Println("📌 User sends: alice")
	userInput := "alice"

	// ✅ SAFE: Using ? placeholders (parameterized query)
	// It tells the database: "I'm going to give you a command, but I’m leaving a blank spot for the data. I'll send the data to you separately in a moment."
	// This is called a Parameterized Query.
	query := "SELECT id, username, email FROM users WHERE username = ?"

	fmt.Println("🔍 Query:", query)
	fmt.Println("🔍 Parameters:", userInput)

	var id int
	var username, email string
	// It’s a two-step process: QueryRow finds the data, and Scan pulls it out.
	// In Go, when you pass a variable to a function like Scan(id), Go makes a copy of that variable.
	// If Scan worked on a copy, it would fill the copy with the user's ID, and then the function would end and the copy would be deleted.
	// Your original id variable in main would stay 0.
	err := db.QueryRow(query, userInput).Scan(&id, &username, &email)

	if err != nil {
		fmt.Println("❌ Error:", err)
	} else {
		fmt.Printf("✅ Found user: %s (%s)\n", username, email)
	}

	fmt.Println("\n📌 User tries to hack: alice' OR '1'='1")
	hackerInput := "alice' OR '1'='1"

	// ✅ SAFE: The hack attempt is treated as literal text
	err = db.QueryRow(query, hackerInput).Scan(&id, &username, &email)

	if err != nil {
		fmt.Println("✅ PROTECTED! No user found (hack failed)")
		fmt.Println("   The database looks for username = \"alice' OR '1'='1\" (literal string)")
	}
}

func demonstrateVulnerableCode(db *sql.DB) {
	fmt.Println("📌 User sends: alice")
	userInput := "alice"

	// ❌ VULNERABLE: String concatenation
	query := fmt.Sprintf("SELECT id, username, email FROM users WHERE username = '%s'", userInput)

	fmt.Println("🔍 Final query:", query)

	var id int
	var username, email string
	err := db.QueryRow(query).Scan(&id, &username, &email)

	if err != nil {
		fmt.Println("❌ Error:", err)
	} else {
		fmt.Printf("✅ Found user: %s (%s)\n", username, email)
	}

	fmt.Println("\n📌 User tries to hack: alice' OR '1'='1")
	hackerInput := "alice' OR '1'='1"

	// ❌ VULNERABLE: Hacker controls the SQL
	query = fmt.Sprintf("SELECT id, username, email FROM users WHERE username = '%s'", hackerInput)

	fmt.Println("🔍 Final query:", query)
	fmt.Println("💥 THE QUERY BECOMES:")
	fmt.Println("   SELECT id, username, email FROM users WHERE username = 'alice' OR '1'='1'")
	fmt.Println("   This means: 'username is alice' OR (always true)")
	fmt.Println("   Result: Returns ALL users!")

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n🚨 HACKER GOT ALL USERS:")
	for rows.Next() {
		rows.Scan(&id, &username, &email)
		fmt.Printf("   - %s (%s)\n", username, email)
	}
}

func demonstrateAttack(db *sql.DB) {
	// Show current users
	fmt.Println("📊 Current database state:")
	showAllUsers(db)

	fmt.Println("\n💀 ATTACKER INPUT: '; DROP TABLE users; --")
	fmt.Println("   ('; = close the string)")
	fmt.Println("   (DROP TABLE users = delete everything)")
	fmt.Println("   (-- = comment out rest)")

	attackInput := "'; DROP TABLE users; --"

	// ❌ VULNERABLE CODE
	query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", attackInput)

	fmt.Println("\n💣 The query becomes:")
	fmt.Println("   SELECT * FROM users WHERE username = ''; DROP TABLE users; --'")
	fmt.Println("   This executes TWO commands:")
	fmt.Println("   1. SELECT * FROM users WHERE username = ''")
	fmt.Println("   2. DROP TABLE users")

	// Execute the malicious query
	_, err := db.Exec(query)
	if err != nil {
		fmt.Println("\n⚠️  Attack executed (error expected):", err)
	}

	fmt.Println("\n🔥 Checking if table still exists...")
	showAllUsers(db)
}

func showAllUsers(db *sql.DB) {
	rows, err := db.Query("SELECT id, username, email FROM users")
	if err != nil {
		fmt.Println("❌ TABLE DESTROYED! Users table no longer exists!")
		fmt.Println("   Your entire database is gone because of string concatenation.")
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var username, email string
		rows.Scan(&id, &username, &email)
		fmt.Printf("   User %d: %s (%s)\n", id, username, email)
		count++
	}
	fmt.Printf("   Total users: %d\n", count)
}

/*
🎓 KEY LESSONS:

1. ✅ ALWAYS USE PARAMETERIZED QUERIES (?)
   query := "SELECT * FROM users WHERE username = ?"
   db.Query(query, userInput)

2. ❌ NEVER CONCATENATE USER INPUT INTO SQL
   query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", userInput)

3. WHY IT MATTERS:
   - One mistake = entire database deleted
   - One mistake = all user data stolen
   - One mistake = you get fired + company sued

4. INTERVIEW QUESTION (COMMON):
   "How do you prevent SQL injection?"

   ANSWER: "Use parameterized queries with placeholders.
            Never concatenate user input into SQL strings.
            The database driver handles escaping automatically."

5. YOUR CODE:
   ✅ Your current code IS SAFE (using ?)
   You just needed to understand WHY it's safe!

🔥 RUN THIS FILE:
   go run learning-demos/01-sql-injection-demo.go
*/

// What the Database receives:

// Box 1 (The Instruction): SELECT * FROM users WHERE username = ?

// Box 2 (The Data): alice'; DROP TABLE users; --

// The Database engine takes the data in Box 2 and "escapes" it. It adds extra slashes or quotes so the ; is treated like a normal character (like a comma in a sentence) instead of a "Stop" command.

// The Database search becomes: "Find a user where the username is exactly: alice'; DROP TABLE users; --"

// Since no user has that crazy name, it just returns "Not Found." The table is safe.
