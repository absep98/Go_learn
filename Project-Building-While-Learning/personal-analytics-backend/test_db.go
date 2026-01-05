package main

import (
	"fmt"
	"log"
	"personal-analytics-backend/internal/db"
)

func main() {
	fmt.Println("=== DATABASE TEST ===\n")

	// Initialize database
	err := db.InitDB("./test_data.db")
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer db.CloseDB()

	// Insert some test entries
	fmt.Println("📝 Inserting test entries...")

	id1, err := db.InsertEntry(101, "Feeling great today! Got a lot done.", 5)
	if err != nil {
		log.Printf("Error inserting entry 1: %v", err)
	} else {
		fmt.Printf("✅ Inserted entry with ID: %d\n", id1)
	}

	id2, err := db.InsertEntry(101, "Bit stressed with work deadlines", 3)
	if err != nil {
		log.Printf("Error inserting entry 2: %v", err)
	} else {
		fmt.Printf("✅ Inserted entry with ID: %d\n", id2)
	}

	id3, err := db.InsertEntry(102, "Had a productive coding session", 4)
	if err != nil {
		log.Printf("Error inserting entry 3: %v", err)
	} else {
		fmt.Printf("✅ Inserted entry with ID: %d\n", id3)
	}

	// Read all entries
	fmt.Println("\n📖 Reading all entries from database...")
	entries, err := db.GetAllEntries()
	if err != nil {
		log.Fatalf("Error reading entries: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("No entries found")
	} else {
		fmt.Printf("Found %d entries:\n\n", len(entries))
		for i, entry := range entries {
			fmt.Printf("Entry #%d:\n", i+1)
			fmt.Printf("  ID: %v\n", entry["id"])
			fmt.Printf("  User ID: %v\n", entry["user_id"])
			fmt.Printf("  Text: %v\n", entry["text"])
			fmt.Printf("  Mood: %v\n", entry["mood"])
			fmt.Printf("  Created: %v\n", entry["created_at"])
			fmt.Println()
		}
	}

	fmt.Println("✅ Test complete!")
}
