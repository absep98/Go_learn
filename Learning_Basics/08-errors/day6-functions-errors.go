package main

import (
	"errors"
	"fmt"
	"strconv"
)

// ============================================
// PART 1: BASIC FUNCTION WITH ERROR HANDLING
// ============================================

// Task: Write a function that parses an int from string, returns (int, error)
func parseInteger(input string) (int, error) {
	// strconv.Atoi converts string to int
	// It returns (int, error) - exactly what we need!
	result, err := strconv.Atoi(input)

	if err != nil {
		// Return zero value for int and the error
		return 0, err
	}

	// Return the parsed integer and nil (no error)
	return result, nil
}

// ============================================
// PART 2: ENHANCED VERSION WITH CUSTOM ERRORS
// ============================================

// More sophisticated version with custom error messages
func parseIntegerEnhanced(input string) (int, error) {
	// Check for empty string first
	if input == "" {
		return 0, errors.New("input cannot be empty")
	}

	// Try to parse the integer
	result, err := strconv.Atoi(input)
	if err != nil {
		// Return a more descriptive error
		return 0, fmt.Errorf("failed to parse '%s' as integer: %v", input, err)
	}

	return result, nil
}

// ============================================
// PART 3: FUNCTION WITH VALIDATION
// ============================================

// Parse integer with range validation
func parseIntegerInRange(input string, min, max int) (int, error) {
	// First parse the integer
	result, err := parseInteger(input)
	if err != nil {
		return 0, err
	}

	// Then validate the range
	if result < min || result > max {
		return 0, fmt.Errorf("number %d is outside valid range [%d, %d]", result, min, max)
	}

	return result, nil
}

// ============================================
// PART 4: DEMONSTRATING ERROR HANDLING PATTERNS
// ============================================

func demonstrateErrorHandling() {
	fmt.Println("=== ERROR HANDLING DEMONSTRATION ===\n")

	// Test cases - mix of valid and invalid inputs
	testInputs := []string{
		"42",                    // Valid
		"123",                   // Valid
		"abc",                   // Invalid - not a number
		"",                      // Invalid - empty
		"999999999999999999999", // Invalid - too large
		"-50",                   // Valid negative
	}

	fmt.Println("1. BASIC PARSING:")
	for _, input := range testInputs {
		result, err := parseInteger(input)

		// The Go way: check error first
		if err != nil {
			fmt.Printf("❌ Input '%s' failed: %v\n", input, err)
		} else {
			fmt.Printf("✅ Input '%s' parsed successfully: %d\n", input, result)
		}
	}

	fmt.Println("\n2. ENHANCED PARSING WITH CUSTOM ERRORS:")
	for _, input := range testInputs {
		result, err := parseIntegerEnhanced(input)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
		} else {
			fmt.Printf("✅ Successfully parsed: %d\n", result)
		}
	}

	fmt.Println("\n3. PARSING WITH RANGE VALIDATION:")
	rangeInputs := []string{"5", "50", "150", "abc"}
	for _, input := range rangeInputs {
		result, err := parseIntegerInRange(input, 1, 100) // Valid range: 1-100
		if err != nil {
			fmt.Printf("❌ %v\n", err)
		} else {
			fmt.Printf("✅ Valid number in range: %d\n", result)
		}
	}
}

// ============================================
// PART 5: PRACTICAL EXAMPLE - CALCULATOR
// ============================================

// Safe calculator that handles string inputs
func safeCalculator() {
	fmt.Println("\n=== SAFE CALCULATOR EXAMPLE ===")

	// Simulate user inputs
	inputs := []struct {
		a, b, operation string
	}{
		{"10", "5", "+"},
		{"20", "4", "/"},
		{"abc", "5", "+"},  // Invalid first number
		{"10", "xyz", "*"}, // Invalid second number
		{"15", "0", "/"},   // Division by zero
	}

	for _, input := range inputs {
		result, err := calculate(input.a, input.b, input.operation)
		if err != nil {
			fmt.Printf("❌ Error calculating %s %s %s: %v\n",
				input.a, input.operation, input.b, err)
		} else {
			fmt.Printf("✅ %s %s %s = %d\n",
				input.a, input.operation, input.b, result)
		}
	}
}

// Calculator function with comprehensive error handling
func calculate(aStr, bStr, operation string) (int, error) {
	// Parse first number
	a, err := parseInteger(aStr)
	if err != nil {
		return 0, fmt.Errorf("invalid first number '%s': %v", aStr, err)
	}

	// Parse second number
	b, err := parseInteger(bStr)
	if err != nil {
		return 0, fmt.Errorf("invalid second number '%s': %v", bStr, err)
	}

	// Perform operation
	switch operation {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, errors.New("division by zero is not allowed")
		}
		return a / b, nil
	default:
		return 0, fmt.Errorf("unknown operation '%s'", operation)
	}
}

// ============================================
// PART 6: ERROR HANDLING BEST PRACTICES
// ============================================

func demonstrateBestPractices() {
	fmt.Println("\n=== ERROR HANDLING BEST PRACTICES ===")

	// ❌ BAD: Ignoring errors
	// result, _ := parseInteger("abc")  // DON'T DO THIS!

	// ✅ GOOD: Always check errors
	result, err := parseInteger("123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return // Handle the error appropriately
	}
	fmt.Printf("Success: %d\n", result)

	// ✅ GOOD: Early return pattern
	value, err := parseIntegerInRange("50", 1, 100)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}
	// Continue with valid value
	fmt.Printf("Valid value: %d\n", value)
}

// ============================================
// MAIN FUNCTION - RUN ALL EXAMPLES
// ============================================

func main() {
	fmt.Println("📚 DAY 6: FUNCTIONS, RETURN VALUES, AND ERRORS")
	fmt.Println("====================================================")

	// Run all demonstrations
	demonstrateErrorHandling()
	safeCalculator()
	demonstrateBestPractices()

	fmt.Println("\n🎯 KEY TAKEAWAYS:")
	fmt.Println("1. Always check errors with 'if err != nil'")
	fmt.Println("2. Use early returns for error cases")
	fmt.Println("3. Provide helpful error messages")
	fmt.Println("4. Functions can return multiple values (value, error)")
	fmt.Println("5. Use fmt.Errorf for formatted error messages")
}
