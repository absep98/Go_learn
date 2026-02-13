package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Print("Type something: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	fmt.Println("You said:", input)

	// IF example
	if len(input) == 0 {
		fmt.Println("You typed nothing.")
	} else if len(input) < 5 {
		fmt.Println("Short input.")
	} else {
		fmt.Println("Long input.")
	}

	// FOR example (repeat input 3 times)
	for i := 0; i < 3; i++ {
		fmt.Println("Echo", i, ":", input)
	}

	// SWITCH example
	switch input {
	case "hi", "hello":
		fmt.Println("Greeting detected.")
	case "bye":
		fmt.Println("Goodbye detected.")
	default:
		fmt.Println("No specific keyword detected.")
	}
}
