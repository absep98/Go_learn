package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

type wordCount struct {
	Word  string
	Count int
}

func main() {

	filename := flag.String("file", "empty", "Enter the file path whichs need to get done word count")
	topN := flag.Int("topN", 0, "how many top words to print")

	flag.Parse()

	if *filename == "empty" {
		log.Fatalf("File path is required in order to process for word count")
	}

	fileData, err := os.ReadFile(*filename)
	if err != nil {
		log.Fatalf("error while reading the file")
	}

	fileWords := strings.Fields(string(fileData))
	counts := make(map[string]int)
	for _, word := range fileWords {
		counts[word]++
	}

	var ranked []wordCount

	for w, n := range counts {
		ranked = append(ranked, wordCount{Word: w, Count: n})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Count > ranked[j].Count
	})

	log.Println(*topN)
	for i := 0; i < *topN && i < len(ranked); i++ {
		fmt.Printf("%d. %s (%d)\n", i+1, ranked[i].Word, ranked[i].Count)
	}
}
