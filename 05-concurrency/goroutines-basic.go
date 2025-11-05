package main

import (
	"fmt"
	"time"
)

// Simple function to demonstrate goroutines
func sayHello(name string) {
	for i := 1; i <= 3; i++ {
		fmt.Printf("Hello from %s - %d\n", name, i)
		time.Sleep(100 * time.Millisecond) // Small delay to see concurrent execution
	}
	fmt.Printf("%s finished!\n", name)
}

// Function that takes longer to demonstrate concurrency
func countNumbers(name string, max int) {
	for i := 1; i <= max; i++ {
		fmt.Printf("%s: %d\n", name, i)
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Printf("%s counting done!\n", name)
}

func main() {
	fmt.Println("=== Go Goroutines ===")

	// FIRST: Without goroutines (sequential)
	fmt.Println("\n--- Without Goroutines (Sequential) ---")
	start := time.Now()

	sayHello("Alice")
	sayHello("Bob")

	fmt.Printf("Sequential execution took: %v\n", time.Since(start))

	// SECOND: With goroutines (concurrent)
	fmt.Println("\n--- With Goroutines (Concurrent) ---")
	start = time.Now()

	go sayHello("Charlie") // 'go' keyword starts a goroutine
	go sayHello("Diana")   // This runs concurrently with Charlie

	// Wait a bit to let goroutines finish
	// (We'll learn a better way with channels soon!)
	time.Sleep(1 * time.Second)

	fmt.Printf("Concurrent execution took: %v\n", time.Since(start))

	// THIRD: Multiple goroutines with different tasks
	fmt.Println("\n--- Multiple Different Goroutines ---")

	go countNumbers("Counter-1", 3)
	go countNumbers("Counter-2", 4)
	go sayHello("Greeter")

	// Wait for all goroutines to finish
	time.Sleep(2 * time.Second)

	fmt.Println("\n🎉 All goroutines completed!")
	fmt.Println("Notice how they all ran AT THE SAME TIME!")
}
