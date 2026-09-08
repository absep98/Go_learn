package main

import "fmt"

func main() {
	// Variable declarations
	var name string = "Alice"
	var age int = 25
	var height float64 = 5.8
	var isStudent bool = true
	
	// Short variable declaration
	city := "New York"
	temperature := 72.5
	
	// Multiple variable declaration
	var x, y, z int = 1, 2, 3
	
	fmt.Println("Name:", name)
	fmt.Println("Age:", age)
	fmt.Println("Height:", height)
	fmt.Println("Is Student:", isStudent)
	fmt.Println("City:", city)
	fmt.Println("Temperature:", temperature)
	fmt.Println("Coordinates:", x, y, z)
	
	// Zero values
	var defaultInt int
	var defaultString string
	var defaultBool bool
	
	fmt.Printf("Default int: %d, string: '%s', bool: %t\n", defaultInt, defaultString, defaultBool)
}