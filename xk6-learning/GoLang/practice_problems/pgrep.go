package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

func grepFile(path string, pattern string, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error while reading the file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			fmt.Printf("%s:%d: %s\n", path, lineCount, line)
		}
	}
}

func main() {

	filepaths := flag.String("file", "empty", "Enter the file paths")
	pattern := flag.String("pattern", "", "Enter the string to find in a file")
	flag.Parse()

	if *filepaths == "empty" {
		log.Fatalf("File path is required in order to process or perfrom operation over it.")
	}

	if *pattern == "" {
		log.Fatalf("enter the pattern need to search in file")
	}

	var wg sync.WaitGroup

	filePaths := strings.Split(*filepaths, ",")

	for _, path := range filePaths {
		wg.Add(1)
		go grepFile(path, *pattern, &wg)
	}

	wg.Wait()
}
