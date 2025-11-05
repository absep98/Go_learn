package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	// "unused/package" - This is unused and will be removed by go mod tidy
)

func main() {
	fmt.Println("Go Modules Explanation")

	// Using encoding/json
	data := map[string]string{"name": "Go"}
	jsonData, _ := json.Marshal(data)
	fmt.Println(string(jsonData))

	// Using os
	fmt.Println("Working directory:", os.Getenv("PWD"))

	// Using log
	log.Println("This is a log message")
}
