package main

import "fmt"

type Logger interface {
	Log(message string)
}

type ConsoleLooger struct{}

func (c ConsoleLooger) Log(message string) {
	fmt.Println("LOG:", message)
}

type FileLogger struct {
	Filename string
}

func (f FileLogger) Log(message string) {
	fmt.Println("Writing to a file : ", f.Filename, message)
}

func ProcessOrder(logger Logger) {
	logger.Log("Order processed")
}

func main() {
	console := ConsoleLooger{}
	file := FileLogger{Filename: "app.txt"}

	ProcessOrder(console)
	ProcessOrder(file)
}
