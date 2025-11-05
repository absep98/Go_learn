package main

import "fmt"

// TODO: Create your own functions here!

// 1. Create a function that takes a name and age, returns a greeting message
func introduce(name string, age int) string {
	// Your code here - return something like "Hi, I'm John and I'm 25 years old"
	return fmt.Sprintf("Hi, I'm %s and I'm %d years old", name, age) // Replace this
}

// 2. Create a function that calculates area of a rectangle
func rectangleArea(length, width float64) float64 {
	// Your code here
	return length * width // Replace this
}

// 3. Create a function that checks if a number is even
// Return the result and a message
func isEven(number int) (bool, string) {
	if number%2 == 0 {
		return true, fmt.Sprintf("%d is even", number)
	}
	return false, fmt.Sprintf("%d is odd", number)
}

// 4. Create a function that finds the maximum of multiple numbers
func findMax(numbers ...int) int {
	if len(numbers) == 0 {
		return 0 // Handle empty case
	}
	max := numbers[0] // Start with first number
	for _, num := range numbers {
		if num > max {
			max = num
		}
	}
	return max
}

func main() {
	fmt.Println("=== Function Practice ===")

	// Test your functions here:

	// Test introduce function
	greeting := introduce("Alice", 25)
	fmt.Println(greeting)

	// Test rectangle area
	area := rectangleArea(5.0, 3.0)
	fmt.Printf("Rectangle area: %.2f\n", area)

	// Test even checker
	result, message := isEven(7)
	fmt.Printf("Is even: %t, Message: %s\n", result, message)

	// Test max finder
	maximum := findMax(3, 1, 4, 1, 5, 9, 2, 6)
	fmt.Printf("Maximum number: %d\n", maximum)
}
