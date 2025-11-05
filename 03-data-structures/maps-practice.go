// ### Day 5
// - **Goal:** Maps + basic error handling.
// - **Tasks:**
//   - Build a word counter:
//     - Take a string.
//     - Split into words.
//     - Store counts in `map[string]int`.
//     - Print each word and its count.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Print("Enter a sentence: ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() { // Fixed: should be !scanner.Scan() for error
		fmt.Println("Error in reading input.")
		return
	}
	text := scanner.Text() // Fixed: added := to declare text

	// Split the text into words
	words := strings.Fields(text)
	// Create a map to count word frequencies
	wordCount := make(map[string]int)

	// Count each word
	for _, word := range words {
		word = strings.ToLower(word)
		wordCount[word]++
	}

	// Print results
	fmt.Println("\nWord frequencies:")
	for word, count := range wordCount {
		fmt.Printf("%s: %d\n", word, count)
	}

	fmt.Printf("\nTotal unique words: %d\n", len(wordCount))
	fmt.Printf("Total words: %d\n", len(words))
}
