package main

import "fmt"

// An interface is a CONTRACT: "any type with this method counts as this type."
// No "implements" keyword needed — Go checks automatically (structural typing).
type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

// Pointer receiver — Area() belongs to *Circle, NOT Circle.
func (c *Circle) Area() float64 {
	return 3.14 * c.Radius * c.Radius
}

func main() {
	c := Circle{Radius: 5}
	fmt.Println(c.Area()) // works — Go auto-does (&c).Area() here

	// var s Shape = c        // COMPILE ERROR — Circle does NOT satisfy Shape
	var s Shape = &c // works — *Circle satisfies Shape
	fmt.Println(s.Area())

	// --- Type assertion / type switch ---
	var x interface{} = 42

	// comma-ok form: never panics, tells you if the assertion worked
	n, ok := x.(string)
	fmt.Println(n, ok) // "" false  (42 is not a string)

	m, ok := x.(int)
	fmt.Println(m, ok) // 42 true

	// type switch: ask "what concrete type is inside this interface?"
	describe(&c)
	describe(42)
}

func describe(x interface{}) {
	switch v := x.(type) {
	case *Circle:
		fmt.Println("it's a circle with radius", v.Radius)
	case int:
		fmt.Println("it's an int:", v)
	default:
		fmt.Println("unknown type")
	}
}

// THE ONE RULE:
//   Pointer receiver method (c *Circle) → only *Circle satisfies the interface.
//   Value receiver method   (c Circle)  → BOTH Circle and *Circle satisfy it.
//
// Calling c.Area() directly on a variable always works (Go auto-addresses),
// but assigning to an interface checks the EXACT method set — no auto-addressing there.
