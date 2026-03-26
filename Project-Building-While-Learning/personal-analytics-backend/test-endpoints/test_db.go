package main

import (
	"context"
	"fmt"
	"log"
	"personal-analytics-backend/internal/db"
)

func main() {
	fmt.Println("=== DATABASE TEST ===")

	// Initialize database
	err := db.InitDB("./test_data.db")
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer db.CloseDB()

	ctx := context.Background()

	// Insert some test entries
	fmt.Println("📝 Inserting test entries...")

	id1, err := db.InsertEntry(ctx, 101, "Feeling great today! Got a lot done.", 5, "general")
	if err != nil {
		log.Printf("Error inserting entry 1: %v", err)
	} else {
		fmt.Printf("✅ Inserted entry with ID: %d\n", id1)
	}

	id2, err := db.InsertEntry(ctx, 101, "Bit stressed with work deadlines", 3, "work")
	if err != nil {
		log.Printf("Error inserting entry 2: %v", err)
	} else {
		fmt.Printf("✅ Inserted entry with ID: %d\n", id2)
	}

	id3, err := db.InsertEntry(ctx, 102, "Had a productive coding session", 4, "general")
	if err != nil {
		log.Printf("Error inserting entry 3: %v", err)
	} else {
		fmt.Printf("✅ Inserted entry with ID: %d\n", id3)
	}

	// Read all entries
	fmt.Println("\n📖 Reading all entries from database...")
	entries, err := db.GetAllEntries(ctx)
	if err != nil {
		log.Fatalf("Error reading entries: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("No entries found")
	} else {
		fmt.Printf("Found %d entries:\n\n", len(entries))
		for i, entry := range entries {
			fmt.Printf("Entry #%d:\n", i+1)
			fmt.Printf("  ID: %v\n", entry.ID)
			fmt.Printf("  User ID: %v\n", entry.UserID)
			fmt.Printf("  Text: %v\n", entry.Text)
			fmt.Printf("  Mood: %v\n", entry.Mood)
			fmt.Printf("  Created: %v\n", entry.CreatedAt)
			fmt.Println()
		}
	}

	fmt.Println("✅ Test complete!")
}
