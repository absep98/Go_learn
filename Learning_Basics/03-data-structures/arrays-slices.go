package main

import "fmt"

func main() {
	fmt.Println("=== Arrays and Slices ===")

	// Arrays (fixed size)
	fmt.Println("\n--- Arrays ---")
	var numbers [5]int = [5]int{1, 2, 3, 4, 5}
	fmt.Printf("Array: %v\n", numbers)
	fmt.Printf("Length: %d\n", len(numbers))
	fmt.Printf("First element: %d\n", numbers[0])
	fmt.Printf("Last element: %d\n", numbers[len(numbers)-1])

	// Array literal with implicit size
	colors := [...]string{"red", "green", "blue"}
	fmt.Printf("Colors: %v\n", colors)

	// Slices (dynamic arrays)
	fmt.Println("\n--- Slices ---")

	// Creating slices
	var fruits []string // nil slice
	fmt.Printf("Empty slice: %v (nil: %t)\n", fruits, fruits == nil)

	// Using make
	scores := make([]int, 3, 5) // length 3, capacity 5
	fmt.Printf("Slice with make: %v (len: %d, cap: %d)\n", scores, len(scores), cap(scores))

	// Slice literal
	animals := []string{"cat", "dog", "bird"}
	fmt.Printf("Animal slice: %v\n", animals)

	// Appending to slices
	fmt.Println("\n--- Appending ---")
	append(fruits, "apple")
	fmt.Printf("After appending: %v\n", fruits)
	fruits = append(fruits, "apple")
	fruits = append(fruits, "banana", "cherry")
	fmt.Printf("After appending: %v\n", fruits)

	// Slicing operations
	fmt.Println("\n--- Slicing Operations ---")
	fmt.Printf("All animals: %v\n", animals)
	fmt.Printf("First two: %v\n", animals[:2])
	fmt.Printf("From second: %v\n", animals[1:])
	fmt.Printf("Middle: %v\n", animals[1:2])

	// Copying slices
	fmt.Println("\n--- Copying Slices ---")
	original := []int{1, 2, 3}
	copied := make([]int, len(original))
	copy(copied, original)

	fmt.Printf("Original: %v\n", original)
	fmt.Printf("Copied: %v\n", copied)

	// Modifying one doesn't affect the other
	copied[0] = 99
	fmt.Printf("After modifying copy:\n")
	fmt.Printf("Original: %v\n", original)
	fmt.Printf("Copied: %v\n", copied)

	// 2D slice
	fmt.Println("\n--- 2D Slice ---")
	matrix := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	fmt.Println("Matrix:")
	for i, row := range matrix {
		fmt.Printf("Row %d: %v\n", i, row)
	}
}
