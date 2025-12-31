package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

/*
=================================================================
            GO INTERVIEW PREPARATION TEST
    Comprehensive coverage of all learned topics
=================================================================

INSTRUCTIONS:
1. Answer each question/solve each problem
2. Run sections individually to test your solutions
3. Think through edge cases and gotchas
4. Expected time: 45-60 minutes (interview pace)

SCORING:
- 25+ correct: Ready for Go interviews! 🚀
- 20-24: Good foundation, review weak areas
- <20: Need more practice on fundamentals

Total Questions: 30
*/

// =================================================================
// SECTION 1: BASIC SYNTAX & TYPES (5 questions)
// =================================================================

func section1Basics() {
	fmt.Println("=== SECTION 1: BASICS ===")

	// Q1: What will this print?
	var a int
	var b string
	var c bool
	fmt.Printf("Q1 Answer: a=%d, b='%s', c=%t\n", a, b, c)
	// Expected: a=0, b='', c=false

	// Q2: Fix the compilation error
	// var x := 5  // What's wrong here?
	var x int = 5
	fmt.Printf("Q2 Answer: x=%d\n", x)

	// Q3: Type conversion challenge
	var i int = 42
	var f float64 = float64(i)
	var s string = strconv.Itoa(i)
	fmt.Printf("Q3 Answer: i=%d, f=%.1f, s='%s'\n", i, f, s)

	// Q4: Constants behavior
	const pi = 3.14159
	const greeting = "Hello"
	// const dynamic = time.Now()  // Why does this fail?
	fmt.Printf("Q4 Answer: pi=%f, greeting=%s\n", pi, greeting)

	// Q5: Short variable declaration rules
	name := "Go"
	age := 13
	// name := "Python"  // Why does this fail?
	name = "Python" // But this works
	fmt.Printf("Q5 Answer: name=%s, age=%d\n", name, age)
}

// =================================================================
// SECTION 2: FUNCTIONS & ERROR HANDLING (5 questions)
// =================================================================

// Q6: Multiple return values
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

// Q7: Named return values - spot the issue
func calculate(x int) (result int, doubled int) {
	result = x // Bug: this should be result = x
	doubled = x * 2
	return // naked return
}

// Q8: Variadic function
func sum(numbers ...int) int {
	total := 0
	for _, num := range numbers {
		total += num
	}
	return total
}

// Q9: First-class functions
func applyOperation(a, b int, op func(int, int) int) int {
	return op(a, b)
}

// Q10: Function as return value
func getOperation(opType string) func(int, int) int {
	switch opType {
	case "add":
		return func(a, b int) int { return a + b }
	case "multiply":
		return func(a, b int) int { return a * b }
	default:
		return func(a, b int) int { return 0 }
	}
}

func section2Functions() {
	fmt.Println("\n=== SECTION 2: FUNCTIONS ===")

	// Q6 Test
	result, err := divide(10, 3)
	if err != nil {
		fmt.Printf("Q6 Error: %v\n", err)
	} else {
		fmt.Printf("Q6 Answer: %.2f\n", result)
	}

	// Q7 Test - what's the bug?
	res, doubled := calculate(5)
	fmt.Printf("Q7 Answer: result=%d, doubled=%d (Bug: result should be 5)\n", res, doubled)

	// Q8 Test
	total := sum(1, 2, 3, 4, 5)
	fmt.Printf("Q8 Answer: %d\n", total)

	// Q9 Test
	multiply := func(x, y int) int { return x * y }
	result9 := applyOperation(6, 7, multiply)
	fmt.Printf("Q9 Answer: %d\n", result9)

	// Q10 Test
	addOp := getOperation("add")
	result10 := addOp(3, 4)
	fmt.Printf("Q10 Answer: %d\n", result10)
}

// =================================================================
// SECTION 3: CONTROL FLOW MASTERY (5 questions)
// =================================================================

func section3ControlFlow() {
	fmt.Println("\n=== SECTION 3: CONTROL FLOW ===")

	// Q11: Short variable declaration in if
	if x := 10; x > 5 {
		fmt.Printf("Q11 Answer: x=%d is greater than 5\n", x)
	}
	// fmt.Println(x) // What happens if we uncomment this?

	// Q12: Switch without expression (boolean switch)
	score := 85
	switch {
	case score >= 90:
		fmt.Println("Q12 Answer: A grade")
	case score >= 80:
		fmt.Println("Q12 Answer: B grade") // This will execute
	case score >= 70:
		fmt.Println("Q12 Answer: C grade")
	}

	// Q13: Type switch challenge
	values := []interface{}{42, "hello", 3.14, true}
	for i, v := range values {
		switch v := v.(type) {
		case int:
			fmt.Printf("Q13 Answer[%d]: Integer %d\n", i, v)
		case string:
			fmt.Printf("Q13 Answer[%d]: String '%s'\n", i, v)
		case float64:
			fmt.Printf("Q13 Answer[%d]: Float %.2f\n", i, v)
		default:
			fmt.Printf("Q13 Answer[%d]: Unknown type\n", i)
		}
	}

	// Q14: Range with index and value
	numbers := []int{10, 20, 30}
	for i, v := range numbers {
		fmt.Printf("Q14 Answer: numbers[%d] = %d\n", i, v)
	}

	// Q15: Multiple case values
	day := "Saturday"
	switch day {
	case "Saturday", "Sunday":
		fmt.Println("Q15 Answer: Weekend!")
	case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday":
		fmt.Println("Q15 Answer: Weekday")
	}
}

// =================================================================
// SECTION 4: SLICE INTERNALS - ADVANCED (5 questions)
// =================================================================

func section4SliceInternals() {
	fmt.Println("\n=== SECTION 4: SLICE INTERNALS ===")

	// Q16: Slice capacity and append behavior
	s1 := make([]int, 3, 5) // len=3, cap=5
	fmt.Printf("Q16 Answer: len=%d, cap=%d\n", len(s1), cap(s1))
	s1 = append(s1, 4, 5)
	fmt.Printf("Q16 Answer after append: len=%d, cap=%d\n", len(s1), cap(s1))

	// Q17: Slice sharing memory - TRICKY!
	original := []int{1, 2, 3, 4, 5}
	slice1 := original[1:3] // [2, 3]
	slice2 := original[2:4] // [3, 4]

	slice1[1] = 99 // Changes original[2] to 99

	fmt.Printf("Q17 Answer: original=%v, slice1=%v, slice2=%v\n", original, slice1, slice2)
	// Expected: original=[1 2 99 4 5], slice1=[2 99], slice2=[99 4]

	// Q18: Full slice expression a[i:j:k]
	arr := []int{0, 1, 2, 3, 4, 5}
	s3 := arr[1:3:4] // start=1, end=3, cap=4-1=3
	fmt.Printf("Q18 Answer: s3=%v, len=%d, cap=%d\n", s3, len(s3), cap(s3))

	// Q19: Copy behavior
	src := []int{1, 2, 3}
	dst := make([]int, 2)
	copied := copy(dst, src)
	fmt.Printf("Q19 Answer: src=%v, dst=%v, copied=%d\n", src, dst, copied)

	// Q20: Append causes reallocation
	small := []int{1, 2}
	fmt.Printf("Q20 Before: len=%d, cap=%d, ptr=%p\n", len(small), cap(small), small)
	small = append(small, 3, 4, 5)
	fmt.Printf("Q20 After: len=%d, cap=%d, ptr=%p\n", len(small), cap(small), small)
}

// =================================================================
// SECTION 5: STRUCT MASTERY (5 questions)
// =================================================================

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Q21: Value receiver vs pointer receiver
func (p Person) GetInfo() string {
	return fmt.Sprintf("%s (%d years)", p.Name, p.Age)
}

func (p *Person) UpdateAge(newAge int) {
	p.Age = newAge
}

// Q22: Struct embedding
type Employee struct {
	Person  // Embedded struct
	Salary  int
	Company string
}

func section5Structs() {
	fmt.Println("\n=== SECTION 5: STRUCTS ===")

	// Q21: Method receivers
	p := Person{Name: "Alice", Age: 25}
	fmt.Printf("Q21 Answer: %s\n", p.GetInfo())

	p.UpdateAge(26) // Go automatically handles &p
	fmt.Printf("Q21 Answer after update: %s\n", p.GetInfo())

	// Q22: Embedding
	emp := Employee{
		Person:  Person{Name: "Bob", Age: 30},
		Salary:  50000,
		Company: "TechCorp",
	}

	// Can access embedded fields directly
	fmt.Printf("Q22 Answer: %s works at %s, age %d\n", emp.Name, emp.Company, emp.Age)

	// Q23: JSON marshaling with tags
	jsonData, _ := json.Marshal(p)
	fmt.Printf("Q23 Answer: JSON = %s\n", jsonData)

	// Q24: Zero value of struct
	var emptyPerson Person
	fmt.Printf("Q24 Answer: Zero value = %+v\n", emptyPerson)

	// Q25: Struct comparison
	p1 := Person{Name: "Alice", Age: 25}
	p2 := Person{Name: "Alice", Age: 25}
	fmt.Printf("Q25 Answer: p1 == p2 is %t\n", p1 == p2)
}

// =================================================================
// SECTION 6: MAP MASTERY (5 questions)
// =================================================================

func section6Maps() {
	fmt.Println("\n=== SECTION 6: MAPS ===")

	// Q26: Map creation and zero value
	var m1 map[string]int
	m2 := make(map[string]int)
	m3 := map[string]int{"go": 1, "python": 2}

	fmt.Printf("Q26 Answer: m1=%v (nil), m2=%v (empty), m3=%v\n", m1, m2, m3)

	// Q27: Key existence check
	value, exists := m3["go"]
	fmt.Printf("Q27 Answer: 'go' exists=%t, value=%d\n", exists, value)

	value2, exists2 := m3["java"]
	fmt.Printf("Q27 Answer: 'java' exists=%t, value=%d\n", exists2, value2)

	// Q28: Delete from map
	delete(m3, "python")
	fmt.Printf("Q28 Answer after delete: %v\n", m3)

	// Q29: Map of functions
	operations := map[string]func(int, int) int{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}

	result := operations["add"](5, 3)
	fmt.Printf("Q29 Answer: 5 + 3 = %d\n", result)

	// Q30: Frequency counter (common interview pattern)
	text := "hello world"
	freq := make(map[rune]int)
	for _, char := range text {
		if char != ' ' {
			freq[char]++
		}
	}
	fmt.Printf("Q30 Answer: Character frequency = %v\n", freq)
}

// =================================================================
// BONUS: GOTCHA QUESTIONS (Interview Favorites)
// =================================================================

func bonusGotchas() {
	fmt.Println("\n=== BONUS: COMMON GOTCHAS ===")

	// Gotcha 1: Loop variable in goroutines
	fmt.Println("Gotcha 1: What would this print?")
	// for i := 0; i < 3; i++ {
	//     go func() { fmt.Println("Wrong:", i) }()  // Prints 3, 3, 3
	// }

	// Correct version:
	for i := 0; i < 3; i++ {
		go func(val int) {
			fmt.Printf("Correct: %d ", val)
		}(i)
	}
	fmt.Println()

	// Gotcha 2: Slice append with capacity
	s := make([]int, 0, 3)
	s = append(s, 1, 2, 3)
	fmt.Printf("Gotcha 2: len=%d, cap=%d\n", len(s), cap(s))

	// Gotcha 3: Interface nil check
	var p *Person = nil
	var i interface{} = p
	fmt.Printf("Gotcha 3: i == nil is %t (tricky!)\n", i == nil)
}

// =================================================================
// MAIN FUNCTION - RUN THE TEST
// =================================================================

func main() {
	fmt.Println("🚀 GO INTERVIEW PREPARATION TEST")
	fmt.Println("==================================")

	section1Basics()
	section2Functions()
	section3ControlFlow()
	section4SliceInternals()
	section5Structs()
	section6Maps()
	bonusGotchas()

	fmt.Println("\n🎯 INTERVIEW READY CHECKLIST:")
	fmt.Println("✅ Zero values of all types")
	fmt.Println("✅ Multiple return values & error handling")
	fmt.Println("✅ Variadic functions & first-class functions")
	fmt.Println("✅ Control flow (if, for, switch, type switch)")
	fmt.Println("✅ Slice internals (len, cap, append, sharing)")
	fmt.Println("✅ Struct methods (value vs pointer receivers)")
	fmt.Println("✅ Struct embedding & JSON tags")
	fmt.Println("✅ Map operations & patterns")
	fmt.Println("✅ Common gotchas & edge cases")

	fmt.Println("\n💡 Next Level Topics for Senior Roles:")
	fmt.Println("- Interfaces & polymorphism")
	fmt.Println("- Goroutines & channels")
	fmt.Println("- Context package")
	fmt.Println("- HTTP servers & middleware")
	fmt.Println("- Testing & benchmarking")
}
