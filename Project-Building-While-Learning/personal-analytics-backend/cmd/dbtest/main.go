package main

import (
	"fmt"
	"log"

	"personal-analytics-backend/internal/db"
	"personal-analytics-backend/internal/models"
)

func main() {
	// Initialize database
	err := db.InitDB("./test_data.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	fmt.Println("\n=== Testing Database Operations ===\n")

	// Test 1: Create entries
	fmt.Println("1. Creating entries...")

	entry1, err := models.CreateEntry(1, "Had a great day at work!", 8)
	if err != nil {
		log.Fatalf("Failed to create entry 1: %v", err)
	}
	fmt.Printf("   Created: ID=%d, Mood=%d, Text='%s'\n", entry1.ID, entry1.Mood, entry1.Text)

	entry2, err := models.CreateEntry(1, "Feeling tired but productive", 6)
	if err != nil {
		log.Fatalf("Failed to create entry 2: %v", err)
	}
	fmt.Printf("   Created: ID=%d, Mood=%d, Text='%s'\n", entry2.ID, entry2.Mood, entry2.Text)

	entry3, err := models.CreateEntry(2, "Different user's entry", 7)
	if err != nil {
		log.Fatalf("Failed to create entry 3: %v", err)
	}
	fmt.Printf("   Created: ID=%d, UserID=%d, Text='%s'\n", entry3.ID, entry3.UserID, entry3.Text)

	// Test 2: Read single entry by ID
	fmt.Println("\n2. Reading entry by ID...")

	fetchedEntry, err := models.GetEntryByID(1)
	if err != nil {
		log.Fatalf("Failed to get entry by ID: %v", err)
	}
	fmt.Printf("   Fetched: ID=%d, Text='%s', CreatedAt=%s\n",
		fetchedEntry.ID, fetchedEntry.Text, fetchedEntry.CreatedAt.Format("2006-01-02 15:04:05"))

	// Test 3: Get all entries
	fmt.Println("\n3. Getting all entries...")

	allEntries, err := models.GetAllEntries()
	if err != nil {
		log.Fatalf("Failed to get all entries: %v", err)
	}
	fmt.Printf("   Total entries: %d\n", len(allEntries))
	for _, e := range allEntries {
		fmt.Printf("   - ID=%d, UserID=%d, Mood=%d, Text='%s'\n", e.ID, e.UserID, e.Mood, e.Text)
	}

	// Test 4: Get entries by user ID
	fmt.Println("\n4. Getting entries for User 1...")

	user1Entries, err := models.GetEntriesByUserID(1)
	if err != nil {
		log.Fatalf("Failed to get entries by user ID: %v", err)
	}
	fmt.Printf("   User 1 has %d entries\n", len(user1Entries))
	for _, e := range user1Entries {
		fmt.Printf("   - ID=%d, Mood=%d, Text='%s'\n", e.ID, e.Mood, e.Text)
	}

	fmt.Println("\n=== All Tests Passed! ===")
	fmt.Println("\nDatabase operations working correctly:")
	fmt.Println("✅ CREATE - Insert new entries")
	fmt.Println("✅ READ   - Get entry by ID")
	fmt.Println("✅ READ   - Get all entries")
	fmt.Println("✅ READ   - Get entries by user ID")
}
