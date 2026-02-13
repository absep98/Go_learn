package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// --- Problem 1

func runInteractivePrompt() {
	fmt.Println("\n--- Problem 1: Interactive Console ---")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Say something: ")

		if !scanner.Scan() {
			break
		}

		input := scanner.Text()

		if input == "stop" {
			fmt.Println("Exiting interactive prompt.")
			break
		}

		fmt.Println("You said: ", input)
	}
}

// --- Problem 2
func inputClassification() {
	fmt.Println("\n--- Problem 2 : Input Classification ---")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Println("Error reading input.")
		return
	}
	text := scanner.Text()
	length := len(text)
	if length == 0 {
		fmt.Println("Empty")
	} else if length < 5 {
		fmt.Println("Short")
	} else if length >= 5 && length <= 10 {
		fmt.Println("Medium")
	} else {
		fmt.Println("Long")
	}
}

func keywordDetector() {
	fmt.Println("\n--- Problem 3 : Keyword Detector ---")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Println("Error reading input.")
		return
	}

	word := scanner.Text()
	switch word {
	case "hi", "hello":
		{
			fmt.Println("Greeting")
		}
	case "bye":
		{
			fmt.Println("Farewell")
		}
	default:
		{
			fmt.Println("Unknown keyword")
		}
	}
}

func timeOfDayDetector() {
	fmt.Println("\n--- Problem 4 : Time of Day Detector ---")

	hour := time.Now().Hour()
	if hour < 12 {
		fmt.Println("Morning")
	}
	if hour < 17 {
		fmt.Println("Evening")
	} else {
		fmt.Println("Night")
	}
}

func typeSwitchFunction(value interface{}) {
	fmt.Println("\n--- Problem 5 : Type Switch Function ---")

	switch v := value.(type) {
	case string:
		fmt.Printf("It's a string: %s\n", v)
	case int:
		fmt.Printf("It's an integer: %d\n", v)
	case bool:
		fmt.Printf("It's a Boolean: %t\n", v)
	default:
		fmt.Printf("Unknown type: %T\n", v)
	}

}

func repeatWordNTimes() {
	fmt.Println("\n--- Problem 6 : Repeat Word N times ---")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Println("Error in reading input.")
		return
	}
	word := strings.TrimSpace(scanner.Text())
	scanner.Scan()
	nString := scanner.Text()

	n, err := strconv.Atoi(nString)
	if err != nil {
		fmt.Printf("Invalid input for N: '%s' is not a valid integer.\n", nString)
		return
	}
	for i := 0; i < n; i++ {
		fmt.Printf("%d: %s\n", i+1, word)
	}
}

func loopOverChar() {
	fmt.Println("\n--- Problem 7 : Loop over Characters ---")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Println("Error in reading input.")
		return
	}
	word := strings.TrimSpace(scanner.Text())

	for i, ch := range word {
		fmt.Printf("%d %c\n", i, ch)
	}
}

// --- Main Function (The Orchestrator) ---

func main() {
	// Call Problem 1's function
	runInteractivePrompt() // Commented out so you can focus on Problem 2 testing

	inputClassification()

	keywordDetector()

	timeOfDayDetector()

	typeSwitchFunction(10)

	repeatWordNTimes()

	loopOverChar()

	for i := range 5 {
		fmt.Println(i)
	}

}
