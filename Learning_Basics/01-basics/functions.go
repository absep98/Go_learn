package main

import "fmt"

// Function with parameters and return value
func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Function with multiple parameters
func add(a, b int) int {
	return a + b
}

// Function with multiple return values
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}
	return a / b, nil
}

// Function with named return values
func calculate(x, y int) (sum int, product int) {
	sum = x + y
	product = x * y
	return // naked return
}

// Variadic function (accepts variable number of arguments)
func sum(numbers ...int) int {
	total := 0
	for _, num := range numbers {
		total += num
	}
	return total
}

func main() {
	fmt.Println("=== Go Functions ===")

	// Simple function call
	message := greet("Go Developer")
	fmt.Println(message)

	// Function with multiple parameters
	result := add(5, 3)
	fmt.Printf("5 + 3 = %d\n", result)

	// Function with multiple return values
	quotient, err := divide(10, 2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("10 / 2 = %.2f\n", quotient)
	}

	// Error handling example
	_, err = divide(10, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Named return values
	s, p := calculate(4, 6)
	fmt.Printf("4 + 6 = %d, 4 * 6 = %d\n", s, p)

	// Variadic function
	total := sum(1, 2, 3, 4, 5)
	fmt.Printf("Sum of 1,2,3,4,5 = %d\n", total)

	// Anonymous function (lambda)
	multiply := func(x, y int) int {
		return x * y
	}
	fmt.Printf("Anonymous function: 3 * 4 = %d\n", multiply(3, 4))
}
