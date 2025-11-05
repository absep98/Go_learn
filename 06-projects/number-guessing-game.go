package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("🎲 Welcome to the Number Guessing Game! 🎲")
	fmt.Println("I'm thinking of a number between 1 and 100...")
	fmt.Println("Can you guess it?")

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate random number between 1 and 100
	target := rand.Intn(100) + 1
	attempts := 0
	maxAttempts := 7

	for attempts < maxAttempts {
		attempts++
		fmt.Printf("\nAttempt %d/%d: Enter your guess: ", attempts, maxAttempts)

		var guess int
		_, err := fmt.Scanf("%d", &guess)
		if err != nil {
			fmt.Println("Please enter a valid number!")
			attempts-- // Don't count invalid input as an attempt
			continue
		}

		if guess == target {
			fmt.Printf("🎉 Congratulations! You guessed it in %d attempts!\n", attempts)
			fmt.Printf("The number was %d\n", target)
			return
		} else if guess < target {
			fmt.Println("📈 Too low! Try a higher number.")
		} else {
			fmt.Println("📉 Too high! Try a lower number.")
		}

		// Give hints based on how close they are
		diff := target - guess
		if diff < 0 {
			diff = -diff
		}

		if diff <= 5 {
			fmt.Println("🔥 You're very close!")
		} else if diff <= 15 {
			fmt.Println("😊 Getting warmer!")
		} else {
			fmt.Println("🧊 You're quite far off!")
		}
	}

	fmt.Printf("\n😢 Sorry, you've run out of attempts!\n")
	fmt.Printf("The number was %d\n", target)
	fmt.Println("Better luck next time!")
}
