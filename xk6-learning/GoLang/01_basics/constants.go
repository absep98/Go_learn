package main

import "fmt"

const Pi = 3.14159
const MaxUsers = 100

func main() {
	// Constants
	const greeting = "Hello"
	const daysInWeek = 7
	
	// Multiple constants
	const (
		StatusOK = 200
		StatusNotFound = 404
		StatusError = 500
	)
	
	fmt.Println("Pi:", Pi)
	fmt.Println("Max Users:", MaxUsers)
	fmt.Println("Greeting:", greeting)
	fmt.Println("Days in week:", daysInWeek)
	fmt.Println("HTTP Status OK:", StatusOK)
}