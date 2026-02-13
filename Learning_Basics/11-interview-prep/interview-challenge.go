package main

import "fmt"

/*
=================================================================
           🔥 TRUE GO INTERVIEW CHALLENGE 🔥
    Answer these WITHOUT running the code first!
=================================================================

INSTRUCTIONS:
1. Read each question carefully
2. WRITE DOWN your answer before running
3. Then run the code to check
4. This simulates real interview pressure!

Time yourself: Can you answer all in 15 minutes?
*/

func main() {
	fmt.Println("🔥 GO INTERVIEW CHALLENGE - ANSWER BEFORE RUNNING!")
	fmt.Println("=" * 60)

	// ==============================================
	// QUESTION 1: SLICE GOTCHA (5 points)
	// ==============================================
	fmt.Println("Q1: What will this print?")
	s1 := []int{1, 2, 3, 4, 5}
	s2 := s1[1:3]  // s2 = [2, 3]
	s3 := s1[2:4]  // s3 = [3, 4]
	
	s2[1] = 99     // Changes s2[1], but what else?
	
	fmt.Printf("s1: %v\n", s1)
	fmt.Printf("s2: %v\n", s2) 
	fmt.Printf("s3: %v\n", s3)
	
	// YOUR ANSWER:
	// s1: [?, ?, ?, ?, ?]
	// s2: [?, ?]
	// s3: [?, ?]

	// ==============================================
	// QUESTION 2: LOOP VARIABLE CLOSURE (5 points)
	// ==============================================
	fmt.Println("\nQ2: What will this print?")
	funcs := make([]func(), 3)
	
	for i := 0; i < 3; i++ {
		funcs[i] = func() {
			fmt.Printf("%d ", i)
		}
	}
	
	for _, f := range funcs {
		f()
	}
	fmt.Println()
	
	// YOUR ANSWER: Will print "? ? ?"

	// ==============================================
	// QUESTION 3: INTERFACE NIL TRAP (5 points)
	// ==============================================
	fmt.Println("\nQ3: What will this print?")
	type MyInterface interface {
		DoSomething()
	}
	
	type MyStruct struct{}
	func (m *MyStruct) DoSomething() {}
	
	var s *MyStruct = nil
	var i MyInterface = s
	
	fmt.Printf("s == nil: %t\n", s == nil)
	fmt.Printf("i == nil: %t\n", i == nil)
	
	// YOUR ANSWER:
	// s == nil: ?
	// i == nil: ?

	// ==============================================
	// QUESTION 4: APPEND BEHAVIOR (5 points)
	// ==============================================
	fmt.Println("\nQ4: What will this print?")
	a := make([]int, 2, 4)  // len=2, cap=4
	a[0] = 1
	a[1] = 2
	
	b := append(a, 3)  // What happens to 'a'?
	c := append(a, 4)  // What happens now?
	
	fmt.Printf("a: %v\n", a)
	fmt.Printf("b: %v\n", b)
	fmt.Printf("c: %v\n", c)
	
	// YOUR ANSWER:
	// a: [?, ?]
	// b: [?, ?, ?]
	// c: [?, ?, ?]

	// ==============================================
	// QUESTION 5: MAP BEHAVIOR (5 points)
	// ==============================================
	fmt.Println("\nQ5: What will this print?")
	var m map[string]int
	
	// What happens when we try to read from nil map?
	value, ok := m["key"]
	fmt.Printf("value: %d, ok: %t\n", value, ok)
	
	// What happens when we try to write to nil map?
	// PREDICTION: Will this panic? YES/NO (write your answer)
	// Uncomment ONLY if you want to test (will crash the program):
	// m["key"] = 1
	
	// YOUR ANSWER:
	// value: ?, ok: ?
	// Writing to nil map will: (panic/work/compile error)?

	// ==============================================
	// QUESTION 6: STRUCT COMPARISON (5 points)
	// ==============================================
	fmt.Println("\nQ6: What will this print?")
	type Person struct {
		Name string
		Age  int
		Tags []string  // Slice field
	}
	
	p1 := Person{Name: "Alice", Age: 25, Tags: []string{"dev"}}
	p2 := Person{Name: "Alice", Age: 25, Tags: []string{"dev"}}
	
	// PREDICTION: Will this compile? YES/NO
	// If YES, what result? TRUE/FALSE
	// Uncomment to test:
	// fmt.Printf("p1 == p2: %t\n", p1 == p2)
	
	// YOUR ANSWER: 
	// Compiles: yes/no
	// If yes, result: true/false
	// If no, why: ?

	// ==============================================
	// QUESTION 7: TYPE SWITCH EDGE CASE (5 points)
	// ==============================================
	fmt.Println("\nQ7: What will this print?")
	var x interface{} = "hello"
	
	switch v := x.(type) {
	case string:
		fmt.Printf("String: %s (len=%d)\n", v, len(v))
	case int:
		fmt.Printf("Integer: %d\n", v)
	default:
		fmt.Printf("Unknown: %v\n", v)
	}
	
	// Now what if x = nil?
	x = nil
	switch v := x.(type) {
	case string:
		fmt.Printf("String: %s\n", v)
	case nil:
		fmt.Printf("Nil value\n")
	default:
		fmt.Printf("Unknown: %v\n", v)
	}
	
	// YOUR ANSWER:
	// First switch prints: ?
	// Second switch prints: ?

	fmt.Println("\n🎯 CHALLENGE COMPLETE!")
	fmt.Println("Now check your answers against the output!")
}