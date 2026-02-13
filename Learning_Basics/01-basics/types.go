package main

import "fmt"

func main() {
	fmt.Println("=== Go Data Types ===")

	// Basic types
	var (
		// Numbers
		age         int     = 25
		temperature float64 = 98.6

		// Text
		name    string = "Alice"
		initial rune   = 'A' // rune is an alias for int32, represents a Unicode code point

		// Boolean
		isActive bool = true

		// Byte (alias for uint8)
		byteValue byte = 255
	)

	fmt.Printf("Integer: %d (type: %T)\n", age, age)
	fmt.Printf("Float: %.1f (type: %T)\n", temperature, temperature)
	fmt.Printf("String: %s (type: %T)\n", name, name)
	fmt.Printf("Rune: %c (type: %T)\n", initial, initial)
	fmt.Printf("Boolean: %t (type: %T)\n", isActive, isActive)
	fmt.Printf("Byte: %d (type: %T)\n", byteValue, byteValue)

	// Type conversion
	fmt.Println("\n=== Type Conversion ===")
	var x int = 42
	var y float64 = float64(x)
	var z string = fmt.Sprintf("%d", x)

	fmt.Printf("int to float64: %d -> %.1f\n", x, y)
	fmt.Printf("int to string: %d -> %s\n", x, z)

	// Zero values (default values when not initialized)
	fmt.Println("\n=== Zero Values ===")
	var (
		defaultInt    int
		defaultFloat  float64
		defaultString string
		defaultBool   bool
	)

	fmt.Printf("Default int: %d\n", defaultInt)
	fmt.Printf("Default float64: %f\n", defaultFloat)
	fmt.Printf("Default string: '%s'\n", defaultString)
	fmt.Printf("Default bool: %t\n", defaultBool)
}
