package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

func parseNumbers(path string) ([]int, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var listNos []int
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		no, err := strconv.Atoi(line)

		if err != nil {
			return nil, fmt.Errorf("could not parse line %d: %v", lineCount, err)
		}
		listNos = append(listNos, no)
	}

	return listNos, nil
}

func main() {

	filepath := flag.String("file", "empty", "Enter the file paths")
	flag.Parse()

	if *filepath == "empty" {
		log.Fatalf("File path is required in order to process or perfrom operation over it.")
	}

	listNos, err := parseNumbers(*filepath)

	if err != nil {
		log.Fatalf("Failed after parsing numbers: %v", err)
	}
	computedSum := 0

	for _, val := range listNos {
		computedSum += val
	}

	if len(listNos) == 0 {
		log.Fatalf("no numbers in file")
	}
	computedAvg := float64(computedSum) / float64(len(listNos))
	fmt.Printf("Sum: %d, average: %.2f\n", computedSum, computedAvg)

}
