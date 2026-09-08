package main

import "fmt"

func main() {
	// If-else statements
	age := 18
	if age >= 18 {
		fmt.Println("You are an adult")
	} else {
		fmt.Println("You are a minor")
	}
	
	// If with initialization
	if score := 85; score >= 90 {
		fmt.Println("Grade: A")
	} else if score >= 80 {
		fmt.Println("Grade: B")
	} else {
		fmt.Println("Grade: C or below")
	}
	
	// For loops
	fmt.Println("\nCounting 1 to 5:")
	for i := 1; i <= 5; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()
	
	// While-style loop
	fmt.Println("\nCountdown:")
	count := 5
	for count > 0 {
		fmt.Print(count, " ")
		count--
	}
	fmt.Println()
	
	// Switch statement
	day := "Monday"
	switch day {
	case "Monday":
		fmt.Println("Start of work week")
	case "Friday":
		fmt.Println("TGIF!")
	case "Saturday", "Sunday":
		fmt.Println("Weekend!")
	default:
		fmt.Println("Regular day")
	}
}