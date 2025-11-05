package main

import "fmt"

func main() {
	fmt.Println("=== UNDERSTANDING ANONYMOUS STRUCT SLICE ===\n")

	// METHOD 1: The long way (what you might expect)
	fmt.Println("1. LONG WAY - Declare struct type first:")

	// First, declare a named struct type
	type CalculationInput struct {
		a, b, operation string
	}

	// Then create a slice of that type
	var longWayInputs []CalculationInput

	// Then add values one by one
	longWayInputs = append(longWayInputs, CalculationInput{a: "10", b: "5", operation: "+"})
	longWayInputs = append(longWayInputs, CalculationInput{a: "20", b: "4", operation: "/"})

	fmt.Printf("Long way result: %+v\n", longWayInputs)

	// METHOD 2: The short way (anonymous struct slice)
	fmt.Println("\n2. SHORT WAY - Anonymous struct slice:")

	shortWayInputs := []struct {
		a, b, operation string
	}{
		{"10", "5", "+"}, // First struct
		{"20", "4", "/"}, // Second struct
	}

	fmt.Printf("Short way result: %+v\n", shortWayInputs)

	// METHOD 3: Let's see what's happening step by step
	fmt.Println("\n3. STEP BY STEP BREAKDOWN:")

	// Step A: Declare the slice type (but don't initialize)
	var stepByStepInputs []struct {
		a, b, operation string
	}

	fmt.Printf("Empty slice: %+v\n", stepByStepInputs)

	// Step B: Add structs to the slice
	stepByStepInputs = append(stepByStepInputs, struct {
		a, b, operation string
	}{
		a:         "10",
		b:         "5",
		operation: "+",
	})

	fmt.Printf("After adding one: %+v\n", stepByStepInputs)

	// METHOD 4: Show how to access the data
	fmt.Println("\n4. ACCESSING THE DATA:")

	inputs := []struct {
		a, b, operation string
	}{
		{"10", "5", "+"},
		{"20", "4", "/"},
		{"abc", "5", "+"},
	}

	for i, input := range inputs {
		fmt.Printf("inputs[%d].a = %s\n", i, input.a)
		fmt.Printf("inputs[%d].b = %s\n", i, input.b)
		fmt.Printf("inputs[%d].operation = %s\n", i, input.operation)
		fmt.Println("---")
	}

	// METHOD 5: Compare with regular slice of strings
	fmt.Println("\n5. COMPARISON WITH REGULAR SLICE:")

	// If we used a regular slice, we'd lose structure
	flatInputs := []string{"10", "5", "+", "20", "4", "/"}
	fmt.Printf("Flat slice: %v\n", flatInputs)
	fmt.Println("Problem: Hard to know which values go together!")

	// With struct slice, the relationship is clear
	fmt.Printf("Structured slice: %+v\n", inputs)
	fmt.Println("Benefit: Each calculation is clearly grouped!")
}
