package main // Every Go program starts with a package declaration. main is special - it's the entry point.

import "fmt" // We import the fmt package for formatted I/O (printing to console).

func main() { // This is the main function where your program starts executing.
	// Your first Go program!
	fmt.Println("Hello, Go World! 🎉")
	fmt.Println("Welcome to your Go learning journey!")

	// Let's explore some basic concepts

	// Variables and types
	var name string = "Go Learner"
	age := 25 // Short variable declaration
	isLearning := true

	fmt.Printf("Name: %s, Age: %d, Learning: %t\n", name, age, isLearning)

	// Constants
	const pi = 3.14159
	fmt.Printf("Pi is approximately: %.2f\n", pi)

	// Multiple variable declaration
	var (
		firstName = "John"
		lastName  = "Doe"
		score     = 95.5
	)

	fmt.Printf("Student: %s %s, Score: %.1f\n", firstName, lastName, score)
}
