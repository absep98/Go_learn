package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {

	filepath := flag.String("file", "empty", "Enter the file path")
	pattern := flag.String("pattern", "", "Enter the string to find in a file")
	flag.Parse()

	if *filepath == "empty" {
		log.Fatalf("File path is required in order to process or perfrom operation over it.")
	}

	if *pattern == "" {
		log.Fatalf("enter the pattern need to search in file")
	}

	file, err := os.Open(*filepath)
	if err != nil {
		log.Fatalf("Error while reading the file")
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		if strings.Contains(line, *pattern) {
			fmt.Printf("%d: %s\n", lineCount, line)
		}
	}

}
