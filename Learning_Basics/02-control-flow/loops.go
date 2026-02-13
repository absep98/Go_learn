package main

import "fmt"

func main() {
	fmt.Println("=== Go Control Flow ===")

	// If-else statements
	fmt.Println("\n--- If-Else ---")
	age := 18

	if age >= 18 {
		fmt.Println("You are an adult!")
	} else {
		fmt.Println("You are a minor.")
	}

	// If with initialization
	if score := 85; score >= 90 {
		fmt.Println("Grade: A")
	} else if score >= 80 {
		fmt.Println("Grade: B")
	} else if score >= 70 {
		fmt.Println("Grade: C")
	} else {
		fmt.Println("Grade: F")
	}

	// For loops
	fmt.Println("\n--- For Loops ---")

	// Traditional for loop
	fmt.Println("Counting 1 to 5:")
	for i := 1; i <= 5; i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// While-style loop
	fmt.Println("Countdown:")
	count := 5
	for count > 0 {
		fmt.Printf("%d ", count)
		count--
	}
	fmt.Println("Blast off! 🚀")

	// Infinite loop with break
	fmt.Println("Finding first even number:")
	num := 1
	for {
		if num%2 == 0 {
			fmt.Printf("Found even number: %d\n", num)
			break
		}
		num++
	}

	// Continue example
	fmt.Println("Odd numbers from 1 to 10:")
	for i := 1; i <= 10; i++ {
		if i%2 == 0 {
			continue // Skip even numbers
		}
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// Range over slice
	fmt.Println("\n--- Range ---")
	fruits := []string{"apple", "banana", "cherry"}

	fmt.Println("Fruits with indices:")
	for index, fruit := range fruits {
		fmt.Printf("%d: %s\n", index, fruit)
	}

	fmt.Println("Just the fruits:")
	for _, fruit := range fruits {
		fmt.Printf("🍎 %s\n", fruit)
	}
}
