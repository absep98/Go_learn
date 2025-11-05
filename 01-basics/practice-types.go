package main

import "fmt"

func main() {
	// Practice with types
	fmt.Println("=== Type Practice ===")

	// Try changing these values and see what happens:
	var score int = 95
	var percentage float64 = float64(score) / 100.0

	fmt.Printf("Score: %d/100\n", score)
	fmt.Printf("Percentage: %.2f\n", percentage)

	// String operations
	firstName := "John"
	lastName := "Doe"
	fullName := firstName + " " + lastName

	fmt.Printf("Full name: %s\n", fullName)
	fmt.Printf("Name length: %d characters\n", len(fullName))

	// Rune (character) example
	letter := 'G'
	fmt.Printf("Letter: %c (Unicode: %d)\n", letter, letter)

	// TODO: Add your own variables here!
	// Try creating:
	// - Your age as an int
	// - Your height as a float64
	// - Your favorite emoji as a rune
	// - Whether you like programming as a bool
	var age int = 26
	var height float64 = 5.9
	var favoriteEmoji rune = 'A'
	var likesProgramming bool = true

	fmt.Printf("My age is %d\n", age)
	fmt.Printf("My height is %.1f\n", height)
	fmt.Printf("My favorite emoji is %c\n", favoriteEmoji)
	fmt.Printf("Do I like programming? %t\n", likesProgramming)
}
