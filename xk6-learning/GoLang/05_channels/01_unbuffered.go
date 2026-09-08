package main

import "fmt"

func main() {
	ch := make(chan int) // unbuffered — send blocks until receive

	// Goroutine must exist BEFORE the blocking operation
	go func() {
		ch <- 42 // send — blocks until main does <-ch
	}()

	val := <-ch // receive — blocks until goroutine sends
	fmt.Println(val) // 42
}

// Unbuffered channel rules:
//   1. Sender and receiver must be DIFFERENT goroutines
//   2. Both goroutines must EXIST before either blocks
//   3. Order doesn't matter — whoever arrives first waits
//
// DEADLOCK example (don't run — it will hang):
//   ch := make(chan int)
//   ch <- 42           // blocks forever, no receiver exists
//   fmt.Println(<-ch)  // never reached
