package main

import "fmt"

// Practice Problem 2: Loops and Logic
// TODO: Complete the functions below

func printNumbers(n int) {
	// TODO: Print numbers from 1 to n
}

func factorial(n int) int {
	// TODO: Calculate and return factorial of n
	// Example: factorial(5) = 5 * 4 * 3 * 2 * 1 = 120
	return 0
}

func sumEvenNumbers(limit int) int {
	// TODO: Return sum of all even numbers from 2 to limit
	// Example: sumEvenNumbers(10) = 2 + 4 + 6 + 8 + 10 = 30
	return 0
}

func main() {
	fmt.Println("Numbers 1 to 5:")
	printNumbers(5)
	
	fmt.Println("Factorial of 5:", factorial(5))        // Should print 120
	fmt.Println("Sum of even numbers up to 10:", sumEvenNumbers(10))  // Should print 30
}