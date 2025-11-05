package main

import (
	"fmt"
	"strconv"
)

/*
===========================================
   GO FUNDAMENTALS TEST (01-03)
   Topics: Basics, Control Flow, Data Structures
===========================================

INSTRUCTIONS:
1. Read each problem carefully
2. Implement the functions marked with TODO
3. Run this file: go run test-01-to-03.go
4. All tests should pass!

Your Score: 0/15 problems
*/

// ============================================
// SECTION 1: BASICS - Types & Functions (5 problems)
// ============================================

// Problem 1: Type Conversion Challenge
// TODO: Fix this function - it should convert temperature from Celsius to Fahrenheit
// Formula: F = (C * 9/5) + 32
// Hint: Watch out for integer division!
func celsiusToFahrenheit(celsius int) float64 {
	// Your code here
	F := float64(celsius)*(9.0/5.0) + 32
	return F // if int was there int(F) and so on for other data types mutation accordingly.
}

// Problem 2: Multiple Return Values
// TODO: Create a function that takes a slice of integers and returns:
// - the minimum value
// - the maximum value
// - the average as float64
// If slice is empty, return (0, 0, 0.0)
func findStats(numbers []int) (int, int, float64) {
	// Your code here
	if len(numbers) <= 0 {
		return 0, 0, 0.0
	}
	minn := numbers[0]
	maxx := numbers[0]
	sum := 0
	for i := 0; i < len(numbers); i++ {
		if minn > numbers[i] {
			minn = numbers[i]
		}
		if maxx < numbers[i] {
			maxx = numbers[i]
		}
		sum += numbers[i]
	}
	return minn, maxx, float64(sum) / float64(len(numbers))
}

// Problem 3: Variadic Function with String Manipulation
// TODO: Create a function that takes multiple strings and returns:
// - A single concatenated string with spaces between words
// - The total character count (excluding spaces)
// Example: joinAndCount("Hello", "Go", "World") -> "Hello Go World", 12
func joinAndCount(words ...string) (string, int) {
	// Your code here
	finalString := ""
	characterCount := 0
	length := len(words)
	for ind, str := range words {
		if ind+1 != length {
			finalString += str + " "
		} else {
			finalString += str
		}
		characterCount += len(str)
	}
	return finalString, characterCount
}

// Problem 4: Named Return Values & Error Handling
// TODO: Create a function that divides two integers and returns quotient and remainder
// Return an error if divisor is 0
// Use named return values: (quotient int, remainder int, err error)
func divideWithRemainder(dividend, divisor int) (quotient int, remainder int, err error) {
	// Your code here
	if divisor == 0 {
		return 0, 0, fmt.Errorf("cannot divide by zero")
	}

	return dividend / divisor, dividend % divisor, nil
}

// Problem 5: First-Class Functions
// TODO: Create a function that takes two integers and a function
// The function parameter should be of type: func(int, int) int
// Apply the function to the two integers and return the result
func operation(a, b int) int {
	return a + b
}
func applyOperation(a, b int, operation func(int, int) int) int {
	// Your code here
	return operation(a, b)
}

// ============================================
// SECTION 2: CONTROL FLOW (3 problems)
// ============================================

// Problem 6: FizzBuzz Classic
// TODO: Return a slice of strings from 1 to n where:
// - Multiples of 3: "Fizz"
// - Multiples of 5: "Buzz"
// - Multiples of both: "FizzBuzz"
// - Other numbers: the number as string
// Example: fizzBuzz(5) -> ["1", "2", "Fizz", "4", "Buzz"]
func fizzBuzz(n int) []string {
	// Your code here
	var ans []string

	for i := 1; i <= n; i++ {
		if i%3 == 0 && i%5 == 0 {
			ans = append(ans, "FizzBuzz")
		} else if i%3 == 0 {
			ans = append(ans, "Fizz")
		} else if i%5 == 0 {
			ans = append(ans, "Buzz")
		} else {
			ans = append(ans, strconv.Itoa(i))
		}
	}

	return ans
}

// Problem 7: Switch Statement with Type
// TODO: Create a function that takes an interface{} and returns a string describing its type
// Handle: int, float64, string, bool, and default case
// Example: "Integer: 42", "Float: 3.14", "String: hello", "Boolean: true", "Unknown type"
func describeType(value interface{}) string {
	// Your code here
	var ans string = ""
	switch v := value.(type) {
	case string:
		ans = "String: " + v
	case int:
		ans = "Integer: " + strconv.Itoa(v)
	case bool:
		ans = "Boolean: " + strconv.FormatBool(v)
	case float64:
		ans = "Float: " + strconv.FormatFloat(v, 'f', -1, 64)
	default:
		ans = "Unknown type"
	}
	return ans
}

// Problem 8: Nested Loops - Pattern Printing
// TODO: Generate a 2D slice representing a multiplication table up to n
// Example: multiplicationTable(3) ->
// [[1,2,3], [2,4,6], [3,6,9]]
func multiplicationTable(n int) [][]int {
	// Your code here
	mat := [][]int{}
	for i := 1; i <= n; i++ {
		row := []int{}
		for j := 1; j <= n; j++ {
			row = append(row, i*j)
		}
		mat = append(mat, row)
	}
	return mat
}

// ============================================
// SECTION 3: DATA STRUCTURES (7 problems)
// ============================================

// Problem 9: Slice Manipulation
// TODO: Remove all even numbers from a slice and return the new slice
// Don't modify the original slice
// Example: removeEvens([1,2,3,4,5,6]) -> [1,3,5]
func removeEvens(numbers []int) []int {
	// Your code here
	ans := []int{}
	for _, no := range numbers {
		if no%2 == 0 {
			continue
		}
		ans = append(ans, no)
	}
	return ans
}

// Problem 10: Map Operations
// TODO: Count the frequency of each word in the slice
// Return a map[string]int where key is word and value is count
// Example: wordCount(["go", "is", "go", "great"]) -> {"go": 2, "is": 1, "great": 1}
func wordCount(words []string) map[string]int {
	// Your code here
	var scores map[string]int
	scores = make(map[string]int)

	for _, str := range words {
		scores[str]++
	}
	return scores
}

// Problem 11: Struct Definition & Methods
// TODO: Define a struct called 'Rectangle' with width and height (float64)
// Then implement the Area() and Perimeter() methods below

type Rectangle struct {
	width  float64
	height float64
}

// TODO: Implement Area method for Rectangle
func (r Rectangle) Area() float64 {
	// Your code here
	return r.height * r.width
}

// TODO: Implement Perimeter method for Rectangle
func (r Rectangle) Perimeter() float64 {
	// Your code here
	return 2 * (r.height + r.width)
}

// Problem 12: Struct with Slice Field
// TODO: Define a Student struct with name (string) and grades ([]int)
// Implement the Average() method that returns the average grade as float64

type Student struct {
	name   string
	grades []int
}

// TODO: Implement Average method
func (s Student) Average() float64 {
	// Your code here
	sum := 0.0
	length := len(s.grades)
	for _, grade := range s.grades {
		sum += float64(grade)
	}

	return sum / float64(length)
}

// Problem 13: Map of Structs
// TODO: Create a function that takes a slice of Students and returns
// a map where the key is the student name and value is their average grade
// Only include students with average >= 60
func getPassingStudents(students []Student) map[string]float64 {
	// Your code here
	var ans map[string]float64
	ans = make(map[string]float64)

	for _, student := range students {
		avgg := student.Average()

		if avgg > 60.0 {
			ans[student.name] = avgg
		}
	}

	return ans
}

// Problem 14: Slice of Slices (2D Array)
// TODO: Transpose a matrix (2D slice)
// Example: [[1,2,3], [4,5,6]] -> [[1,4], [2,5], [3,6]]
// Assume all rows have the same length
func transposeMatrix(matrix [][]int) [][]int {
	// Your code here
	n := len(matrix)
	m := len(matrix[0])

	ans := [][]int{}

	for i := 0; i < m; i++ {
		row := []int{}
		for j := 0; j < n; j++ {
			row = append(row, matrix[j][i])
		}
		ans = append(ans, row)
	}
	return ans
}

// Problem 15: CHALLENGE - Combining Everything
// TODO: Create a function that:
// 1. Takes a map[string][]int (person name -> list of scores)
// 2. Returns a slice of names sorted by their average score (highest first)
// 3. Only include people with at least 3 scores
// You'll need: maps, slices, sorting logic, and calculations
func topPerformers(scores map[string][]int) []string {
	// Your code here

	// Step 1: Create a struct to hold name and average
	type performer struct {
		name string
		avg  float64
	}

	// Step 2: Collect all valid performers (≥3 scores)
	var candidates []performer

	for name, grades := range scores {
		// Only include people with at least 3 scores
		if len(grades) < 3 {
			continue
		}

		// Calculate average
		sum := 0.0
		for _, score := range grades {
			sum += float64(score)
		}
		avg := sum / float64(len(grades))

		// Add to candidates
		candidates = append(candidates, performer{name: name, avg: avg})
	}

	// Step 3: Sort candidates by average (highest first) using bubble sort
	for i := 0; i < len(candidates); i++ {
		for j := 0; j < len(candidates)-1-i; j++ {
			if candidates[j].avg < candidates[j+1].avg {
				// Swap
				candidates[j], candidates[j+1] = candidates[j+1], candidates[j]
			}
		}
	}

	// Step 4: Extract names in sorted order
	result := []string{}
	for _, performer := range candidates {
		result = append(result, performer.name)
	}

	return result
}

// ============================================
// TEST CASES - DON'T MODIFY BELOW
// ============================================

func main() {
	fmt.Println("===========================================")
	fmt.Println("   GO FUNDAMENTALS TEST (01-03)")
	fmt.Println("===========================================\n")

	passed := 0
	total := 15

	// Test 1
	fmt.Println("Test 1: Celsius to Fahrenheit")
	if result := celsiusToFahrenheit(0); result == 32.0 {
		fmt.Println("✓ PASS: 0°C = 32°F")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected 32.0, got %.2f\n", result)
	}
	if result := celsiusToFahrenheit(100); result == 212.0 {
		fmt.Println("✓ PASS: 100°C = 212°F")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected 212.0, got %.2f\n", result)
	}
	total++ // This test counts as 2

	// Test 2
	fmt.Println("\nTest 2: Find Stats")
	min, max, avg := findStats([]int{1, 2, 3, 4, 5})
	if min == 1 && max == 5 && avg == 3.0 {
		fmt.Println("✓ PASS: Stats correct for [1,2,3,4,5]")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected (1, 5, 3.0), got (%d, %d, %.2f)\n", min, max, avg)
	}

	// Test 3
	fmt.Println("\nTest 3: Join and Count")
	str, count := joinAndCount("Hello", "Go", "World")
	if str == "Hello Go World" && count == 12 {
		fmt.Println("✓ PASS: Joined correctly and counted 12 chars")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected ('Hello Go World', 12), got ('%s', %d)\n", str, count)
	}

	// Test 4
	fmt.Println("\nTest 4: Divide with Remainder")
	q, r, err := divideWithRemainder(10, 3)
	if err == nil && q == 3 && r == 1 {
		fmt.Println("✓ PASS: 10/3 = 3 remainder 1")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected (3, 1, nil), got (%d, %d, %v)\n", q, r, err)
	}

	// Test 5
	fmt.Println("\nTest 5: Apply Operation")
	multiply := func(a, b int) int { return a * b }
	if result := applyOperation(6, 7, multiply); result == 42 {
		fmt.Println("✓ PASS: Applied multiplication function")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected 42, got %d\n", result)
	}

	// Test 6
	fmt.Println("\nTest 6: FizzBuzz")
	fb := fizzBuzz(15)
	if len(fb) == 15 && fb[2] == "Fizz" && fb[4] == "Buzz" && fb[14] == "FizzBuzz" {
		fmt.Println("✓ PASS: FizzBuzz works correctly")
		passed++
	} else {
		fmt.Printf("✗ FAIL: FizzBuzz incorrect. Got: %v\n", fb)
	}

	// Test 7
	fmt.Println("\nTest 7: Describe Type")
	if describeType(42) == "Integer: 42" && describeType("hello") == "String: hello" {
		fmt.Println("✓ PASS: Type description correct")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Type descriptions incorrect\n")
	}

	// Test 8
	fmt.Println("\nTest 8: Multiplication Table")
	table := multiplicationTable(3)
	if len(table) == 3 && table[0][0] == 1 && table[2][2] == 9 {
		fmt.Println("✓ PASS: Multiplication table correct")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Table incorrect. Got: %v\n", table)
	}

	// Test 9
	fmt.Println("\nTest 9: Remove Evens")
	odds := removeEvens([]int{1, 2, 3, 4, 5, 6})
	if len(odds) == 3 && odds[0] == 1 && odds[1] == 3 && odds[2] == 5 {
		fmt.Println("✓ PASS: Even numbers removed")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected [1,3,5], got %v\n", odds)
	}

	// Test 10
	fmt.Println("\nTest 10: Word Count")
	wc := wordCount([]string{"go", "is", "go", "great"})
	if wc["go"] == 2 && wc["is"] == 1 && wc["great"] == 1 {
		fmt.Println("✓ PASS: Word count correct")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected map correct counts, got %v\n", wc)
	}

	// Test 11
	fmt.Println("\nTest 11: Rectangle Struct")
	rect := Rectangle{width: 5, height: 3}
	if rect.Area() == 15.0 && rect.Perimeter() == 16.0 {
		fmt.Println("✓ PASS: Rectangle methods work")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected Area=15.0, Perimeter=16.0, got Area=%.2f, Perimeter=%.2f\n", rect.Area(), rect.Perimeter())
	}

	// Test 12
	fmt.Println("\nTest 12: Student Average")
	student := Student{name: "Alice", grades: []int{85, 90, 95}}
	if student.Average() == 90.0 {
		fmt.Println("✓ PASS: Student average correct")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected 90.0, got %.2f\n", student.Average())
	}

	// Test 13
	fmt.Println("\nTest 13: Passing Students")
	students := []Student{
		{name: "Alice", grades: []int{85, 90, 95}},
		{name: "Bob", grades: []int{50, 55, 45}},
		{name: "Charlie", grades: []int{70, 75, 80}},
	}
	passing := getPassingStudents(students)
	if len(passing) == 2 && passing["Alice"] == 90.0 && passing["Charlie"] == 75.0 {
		fmt.Println("✓ PASS: Passing students filtered correctly")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected 2 passing students, got %v\n", passing)
	}

	// Test 14
	fmt.Println("\nTest 14: Transpose Matrix")
	matrix := [][]int{{1, 2, 3}, {4, 5, 6}}
	transposed := transposeMatrix(matrix)
	if len(transposed) == 3 && transposed[0][1] == 4 && transposed[2][0] == 3 {
		fmt.Println("✓ PASS: Matrix transposed correctly")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Transpose incorrect. Got: %v\n", transposed)
	}

	// Test 15
	fmt.Println("\nTest 15: Top Performers (CHALLENGE)")
	scoreMap := map[string][]int{
		"Alice":   {90, 85, 95, 88},
		"Bob":     {70, 75}, // Only 2 scores - should be excluded
		"Charlie": {80, 85, 90},
		"David":   {95, 98, 92, 96},
	}
	top := topPerformers(scoreMap)
	if len(top) == 3 && top[0] == "David" && top[1] == "Alice" && top[2] == "Charlie" {
		fmt.Println("✓ PASS: Top performers ranked correctly")
		passed++
	} else {
		fmt.Printf("✗ FAIL: Expected [David, Alice, Charlie], got %v\n", top)
	}

	// Final Score
	fmt.Println("\n===========================================")
	fmt.Printf("   FINAL SCORE: %d/%d (%.1f%%)\n", passed, total, float64(passed)/float64(total)*100)
	fmt.Println("===========================================")

	if passed == total {
		fmt.Println("\n🎉 PERFECT SCORE! You're ready for concurrency!")
	} else if passed >= total*3/4 {
		fmt.Println("\n👍 Good job! Review the failed tests and try again.")
	} else {
		fmt.Println("\n📚 Keep practicing! Review the basics and try again.")
	}
}
