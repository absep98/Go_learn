package main

import "fmt"

/*
=================================================================
      QUICK GO INTERVIEW QUIZ - TRICKY QUESTIONS
      Focus on gotchas and edge cases
=================================================================

Instructions: Answer each question mentally, then run to check

DIFFICULTY LEVELS:
🟢 Junior Go Developer
🟡 Mid-level Go Developer  
🔴 Senior Go Developer
*/

func main() {
	fmt.Println("🧠 GO INTERVIEW QUICK QUIZ")
	fmt.Println("===========================\n")

	// =================================================================
	// 🟢 JUNIOR LEVEL QUESTIONS
	// =================================================================
	
	fmt.Println("🟢 JUNIOR LEVEL:")
	
	// Q1: What will this print?
	var x int
	var y *int
	fmt.Printf("Q1: x=%d, y=%v\n", x, y)
	// Answer: x=0, y=<nil>
	
	// Q2: Fix this compilation error
	// const dynamic = len("hello")  // This works!
	const static = len("hello")
	fmt.Printf("Q2: Length = %d\n", static)
	// Answer: const can use len() at compile time
	
	// Q3: What happens here?
	s := []int{1, 2, 3}
	s = append(s, 4)
	fmt.Printf("Q3: s = %v\n", s)
	// Answer: [1 2 3 4] - append modifies and returns
	
	// =================================================================
	// 🟡 MID-LEVEL QUESTIONS  
	// =================================================================
	
	fmt.Println("\n🟡 MID-LEVEL:")
	
	// Q4: Slice sharing gotcha
	a := []int{1, 2, 3, 4, 5}
	b := a[1:3]  // [2, 3]
	b[0] = 99    // Changes a[1] to 99!
	fmt.Printf("Q4: a=%v, b=%v\n", a, b)
	// Answer: a=[1 99 3 4 5], b=[99 3]
	
	// Q5: Interface nil gotcha
	var p *string = nil
	var i interface{} = p
	fmt.Printf("Q5: p==nil is %t, i==nil is %t\n", p == nil, i == nil)
	// Answer: p==nil is true, i==nil is false (typed nil!)
	
	// Q6: Map initialization
	var m1 map[string]int           // nil map
	m2 := map[string]int{}          // empty map
	m3 := make(map[string]int)      // empty map
	
	// m1["key"] = 1  // This would panic!
	m2["key"] = 1
	m3["key"] = 1
	
	fmt.Printf("Q6: m1=%v, m2=%v, m3=%v\n", m1, m2, m3)
	// Answer: Only m2 and m3 can be written to
	
	// =================================================================
	// 🔴 SENIOR LEVEL QUESTIONS
	// =================================================================
	
	fmt.Println("\n🔴 SENIOR LEVEL:")
	
	// Q7: Method set gotcha
	type Counter struct{ count int }
	
	func (c Counter) Value() int     { return c.count }      // Value receiver
	func (c *Counter) Increment()    { c.count++ }           // Pointer receiver
	
	c1 := Counter{count: 5}
	c1.Increment()  // Go automatically takes address: (&c1).Increment()
	fmt.Printf("Q7: c1.count = %d\n", c1.count)
	
	// Q8: Slice capacity expansion
	s1 := make([]int, 0, 1)
	fmt.Printf("Q8 Before: len=%d, cap=%d\n", len(s1), cap(s1))
	
	s1 = append(s1, 1)  // Uses existing capacity
	fmt.Printf("Q8 Step 1: len=%d, cap=%d\n", len(s1), cap(s1))
	
	s1 = append(s1, 2)  // Exceeds capacity, doubles it
	fmt.Printf("Q8 Step 2: len=%d, cap=%d\n", len(s1), cap(s1))
	// Answer: Capacity typically doubles when exceeded
	
	// Q9: Type assertion panic vs safe check
	var v interface{} = "hello"
	
	// Safe way
	if str, ok := v.(string); ok {
		fmt.Printf("Q9 Safe: Got string '%s'\n", str)
	}
	
	// Panic way (commented to avoid crash)
	// str := v.(string)  // Would panic if v is not string
	
	// Q10: Closure variable capture
	funcs := make([]func(), 3)
	for i := 0; i < 3; i++ {
		// Wrong way - captures loop variable
		// funcs[i] = func() { fmt.Printf("Wrong: %d ", i) }
		
		// Right way - captures value
		funcs[i] = func(val int) func() {
			return func() { fmt.Printf("Q10: %d ", val) }
		}(i)
	}
	
	fmt.Print("Q10 Answer: ")
	for _, f := range funcs {
		f()
	}
	fmt.Println()
	
	// =================================================================
	// RAPID FIRE ROUND
	// =================================================================
	
	fmt.Println("\n⚡ RAPID FIRE:")
	
	// What's the zero value?
	var slice []int
	var mapVar map[string]int  
	var channel chan int
	var function func()
	
	fmt.Printf("RF1: slice==nil: %t\n", slice == nil)
	fmt.Printf("RF2: map==nil: %t\n", mapVar == nil)  
	fmt.Printf("RF3: chan==nil: %t\n", channel == nil)
	fmt.Printf("RF4: func==nil: %t\n", function == nil)
	
	// Append to nil slice (this works!)
	slice = append(slice, 1, 2, 3)
	fmt.Printf("RF5: Append to nil slice: %v\n", slice)
	
	// =================================================================
	// FINAL BOSS QUESTIONS
	// =================================================================
	
	fmt.Println("\n👑 FINAL BOSS:")
	
	// FB1: What does this do?
	nums := []int{1, 2, 3}
	for i, v := range nums {
		nums[i] = v * 2  // Modifying during iteration
	}
	fmt.Printf("FB1: Modified slice: %v\n", nums)
	// Answer: [2 4 6] - safe because range copies values
	
	// FB2: Struct comparison
	type Point struct{ X, Y int }
	p1 := Point{1, 2}
	p2 := Point{1, 2}
	fmt.Printf("FB2: Points equal: %t\n", p1 == p2)
	// Answer: true - structs are comparable if all fields are comparable
	
	// FB3: What happens here?
	defer fmt.Println("FB3: This prints last")
	fmt.Println("FB3: This prints first")
	// Answer: defer executes after function returns
	
	fmt.Println("\n🎯 SCORING GUIDE:")
	fmt.Println("🟢 All Junior (1-3): Basic Go syntax ✓")
	fmt.Println("🟡 All Mid-level (4-6): Good understanding ✓")
	fmt.Println("🔴 All Senior (7-10): Strong fundamentals ✓")
	fmt.Println("⚡ Rapid Fire (RF1-5): Quick recall ✓")
	fmt.Println("👑 Final Boss (FB1-3): Interview ready! 🚀")
}