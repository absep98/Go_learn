package main

import "fmt"

func main() {
	fmt.Println("🚨 NIL MAP DEMONSTRATION\n")

	// ===== WRONG WAY =====
	fmt.Println("❌ WRONG: Using var (creates nil map)")
	var scores map[string]int

	fmt.Printf("scores == nil: %t\n", scores == nil)
	fmt.Printf("scores: %v\n\n", scores)

	// ✅ Can READ from nil map (returns zero value)
	fmt.Println("✅ Reading from nil map (safe):")
	value := scores["math"]
	fmt.Printf("scores['math'] = %d (zero value)\n\n", value)

	// ❌ CANNOT WRITE to nil map (PANICS!)
	fmt.Println("❌ Writing to nil map (PANICS!):")
	fmt.Println("Attempting: scores['math'] = 100")
	// Uncomment to see panic:
	// scores["math"] = 100
	// panic: assignment to entry in nil map

	// ===== RIGHT WAY =====
	fmt.Println("\n✅ CORRECT: Using make() (creates empty map)")
	scores = make(map[string]int) // Initialize with make()

	fmt.Printf("scores == nil: %t\n", scores == nil)
	fmt.Printf("scores: %v\n\n", scores)

	// ✅ Now we can WRITE to it
	fmt.Println("✅ Writing to initialized map:")
	scores["math"] = 100
	scores["english"] = 95
	scores["science"] = 88

	fmt.Printf("scores: %v\n", scores)

	// ===== ALTERNATIVE WAY =====
	fmt.Println("\n✅ ALTERNATIVE: Declare with literal syntax")
	grades := map[string]int{
		"math":    100,
		"english": 95,
		"science": 88,
	}

	fmt.Printf("grades: %v\n", grades)

	// ===== SUMMARY =====
	fmt.Println("\n📝 SUMMARY:")
	fmt.Println("1. var m map[string]int       → NIL map (❌ can't write)")
	fmt.Println("2. m := make(map[string]int)  → Empty map (✅ can write)")
	fmt.Println("3. m := map[string]int{...}   → Initialized map (✅ can write)")
}
