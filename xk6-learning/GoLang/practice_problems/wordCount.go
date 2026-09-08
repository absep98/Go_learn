package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

func main() {
	filename := flag.String("file", "empty", "Enter the file path which needs to get word count for that")

	flag.Parse()

	if *filename == "empty" {
		log.Fatalf("can't find the word count without proper path")
	}

	fileData, err := os.ReadFile(*filename)
	if err != nil {
		log.Fatalf("error while reading the file")
	}

	fileWords := strings.Fields(string(fileData))
	log.Printf("total word count %d", len(fileWords))
}
