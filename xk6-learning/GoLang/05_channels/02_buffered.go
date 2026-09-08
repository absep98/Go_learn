package main

import "fmt"

func main() {
	ch := make(chan string, 2) // buffered — 2 slots

	// Can send WITHOUT a receiver (drops into buffer)
	ch <- "hello"
	ch <- "world"

	// Reads are FIFO — first in, first out
	fmt.Println(<-ch) // hello
	fmt.Println(<-ch) // world

	// Works in same goroutine because buffer holds the value.
	// Unbuffered would deadlock here.
}

// Buffered = mailbox with N slots
//   Send blocks only when buffer is FULL
//   Receive blocks only when buffer is EMPTY
//
// Unbuffered = hand delivery
//   Send blocks until someone receives
//   Receive blocks until someone sends
