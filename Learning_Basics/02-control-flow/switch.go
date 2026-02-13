package main

import "fmt"

func main() {
	fmt.Println("=== Go Switch Statements ===")

	// Basic switch
	fmt.Println("\n--- Basic Switch ---")
	day := "Monday"

	switch day {
	case "Monday":
		fmt.Println("Start of the work week! 💼")
	case "Tuesday", "Wednesday", "Thursday":
		fmt.Println("Midweek grind! ⚡")
	case "Friday":
		fmt.Println("TGIF! 🎉")
	case "Saturday", "Sunday":
		fmt.Println("Weekend vibes! 🏖️")
	default:
		fmt.Println("Invalid day!")
	}

	// Switch with expression
	fmt.Println("\n--- Switch with Expression ---")
	score := 85

	switch {
	case score >= 90:
		fmt.Println("Excellent! 🌟")
	case score >= 80:
		fmt.Println("Good job! 👍")
	case score >= 70:
		fmt.Println("Not bad! 👌")
	case score >= 60:
		fmt.Println("You passed! ✅")
	default:
		fmt.Println("Better luck next time! 😅")
	}

	// Type switch
	fmt.Println("\n--- Type Switch ---")
	var value interface{} = "Hello, Go!"

	switch v := value.(type) {
	case string:
		fmt.Printf("It's a string: %s\n", v)
	case int:
		fmt.Printf("It's an integer: %d\n", v)
	case bool:
		fmt.Printf("It's a boolean: %t\n", v)
	default:
		fmt.Printf("Unknown type: %T\n", v)
	}

	// Switch with fallthrough
	fmt.Println("\n--- Switch with Fallthrough ---")
	number := 2

	switch number {
	case 1:
		fmt.Println("One")
		fallthrough
	case 2:
		fmt.Println("Two or came from One")
		fallthrough
	case 3:
		fmt.Println("Three or came from previous cases")
	default:
		fmt.Println("Other number")
	}
}
