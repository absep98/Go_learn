package main

import (
	"fmt"
	"time"
)

// Function that sends data through a channel
func sendMessages(ch chan string, name string) {
	messages := []string{"Hello", "How are you?", "Goodbye"}

	for _, msg := range messages {
		fullMessage := fmt.Sprintf("%s says: %s", name, msg)
		ch <- fullMessage // Send message to channel
		fmt.Printf("📤 Sent: %s\n", fullMessage)
		time.Sleep(500 * time.Millisecond)
	}

	close(ch) // Close channel when done
}

// Function that processes numbers and sends results
func processNumbers(numbers chan int, results chan int) {
	for num := range numbers { // Receive until channel is closed
		result := num * num // Square the number
		fmt.Printf("🔢 Processing %d -> %d\n", num, result)
		results <- result
		time.Sleep(200 * time.Millisecond)
	}
	close(results)
}

func main() {
	fmt.Println("=== Go Channels ===")

	// EXAMPLE 1: Basic channel communication
	fmt.Println("\n--- Basic Channel Communication ---")

	// Create a channel for strings
	messageChannel := make(chan string)

	// Start goroutine that sends messages
	go sendMessages(messageChannel, "Alice")

	// Receive messages in main goroutine
	for message := range messageChannel { // Loop until channel is closed
		fmt.Printf("📥 Received: %s\n", message)
		time.Sleep(200 * time.Millisecond)
	}

	// EXAMPLE 2: Multiple goroutines with channels
	fmt.Println("\n--- Multiple Goroutines with Channels ---")

	numbers := make(chan int)
	results := make(chan int)

	// Start processor goroutine
	go processNumbers(numbers, results)

	// Send numbers to process
	go func() {
		for i := 1; i <= 5; i++ {
			fmt.Printf("📨 Sending number: %d\n", i)
			numbers <- i
			time.Sleep(100 * time.Millisecond)
		}
		close(numbers)
	}()

	// Collect results
	fmt.Println("Results:")
	for result := range results {
		fmt.Printf("✅ Got result: %d\n", result)
	}

	// EXAMPLE 3: Channel with buffer
	fmt.Println("\n--- Buffered Channel ---")

	bufferedCh := make(chan string, 3) // Buffer size 3

	// Send multiple messages without blocking
	bufferedCh <- "Message 1"
	bufferedCh <- "Message 2"
	bufferedCh <- "Message 3"
	fmt.Println("📦 Sent 3 messages to buffered channel")

	// Receive them
	for i := 0; i < 3; i++ {
		msg := <-bufferedCh
		fmt.Printf("📮 Received from buffer: %s\n", msg)
	}

	fmt.Println("\n🎉 Channel communication complete!")
}
