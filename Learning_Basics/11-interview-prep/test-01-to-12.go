package main

import "fmt"

type Notifier interface {
	Send(msg string) error
}

type Email struct {
	Address string
}

func (e *Email) Send(msg string) error {
	fmt.Printf("Sending %s to %s\n", msg, e.Address)
	return nil
}

func NotifyUser(n Notifier, message string) {
	n.Send(message)
}

func main() {
	var e *Email = &Email{Address: "aks@example.com"}
	NotifyUser(e, "Hello")

	// Initialize the map before writing to avoid a nil map panic
	var scores map[string]int
	scores = make(map[string]int)
	scores["math"] = 100
}
