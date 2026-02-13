package main

import "fmt"

func main() {
	fmt.Println("=== Maps (Key-Value Pairs) ===")

	// Creating maps
	fmt.Println("\n--- Creating Maps ---")

	// Using make
	var scores map[string]int
	scores = make(map[string]int)
	fmt.Printf("Empty map: %v (nil: %t)\n", scores, scores == nil)

	// Map literal
	ages := map[string]int{
		"Alice":   25,
		"Bob":     30,
		"Charlie": 35,
	}
	fmt.Printf("Ages map: %v\n", ages)

	// Adding and accessing elements
	fmt.Println("\n--- Adding and Accessing ---")
	scores["Alice"] = 95
	scores["Bob"] = 87
	scores["Charlie"] = 92

	fmt.Printf("Alice's score: %d\n", scores["Alice"])
	fmt.Printf("All scores: %v\n", scores)

	// Checking if key exists
	fmt.Println("\n--- Checking Key Existence ---")
	score, exists := scores["Alice"]
	if exists {
		fmt.Printf("Alice's score: %d\n", score)
	} else {
		fmt.Println("Alice not found")
	}

	score, exists = scores["David"]
	if exists {
		fmt.Printf("David's score: %d\n", score)
	} else {
		fmt.Println("David not found")
	}

	// Deleting elements
	fmt.Println("\n--- Deleting Elements ---")
	fmt.Printf("Before delete: %v\n", scores)
	delete(scores, "Bob")
	fmt.Printf("After deleting Bob: %v\n", scores)

	// Iterating over maps
	fmt.Println("\n--- Iterating Over Maps ---")
	fmt.Println("All ages:")
	for name, age := range ages {
		fmt.Printf("%s is %d years old\n", name, age)
	}

	// Just keys
	fmt.Println("\nJust the names:")
	for name := range ages {
		fmt.Printf("- %s\n", name)
	}

	// Just values
	fmt.Println("\nJust the ages:")
	for _, age := range ages {
		fmt.Printf("- %d\n", age)
	}

	// Map of slices
	fmt.Println("\n--- Map of Slices ---")
	hobbies := map[string][]string{
		"Alice":   {"reading", "swimming", "coding"},
		"Bob":     {"gaming", "cooking"},
		"Charlie": {"photography", "hiking", "music", "painting"},
	}

	fmt.Println("Everyone's hobbies:")
	for person, hobbyList := range hobbies {
		fmt.Printf("%s: %v\n", person, hobbyList)
	}

	// Nested maps
	fmt.Println("\n--- Nested Maps ---")
	students := map[string]map[string]int{
		"Alice": {
			"Math":    95,
			"Science": 88,
			"English": 92,
		},
		"Bob": {
			"Math":    78,
			"Science": 85,
			"English": 90,
		},
	}

	fmt.Println("Student grades:")
	for student, grades := range students {
		fmt.Printf("%s's grades:\n", student)
		for subject, grade := range grades {
			fmt.Printf("  %s: %d\n", subject, grade)
		}
	}
}
