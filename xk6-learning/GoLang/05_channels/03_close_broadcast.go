package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, stop <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-stop: // unblocks when close(stop) is called
			fmt.Printf("worker %d: received stop signal\n", id)
			return
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func main() {
	stop := make(chan struct{}) // empty struct = zero memory, signal only
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go worker(i, stop, &wg)
	}

	time.Sleep(1 * time.Second)
	fmt.Println("broadcasting stop...")
	close(stop) // wakes up ALL workers simultaneously
	wg.Wait()
	fmt.Println("all stopped")
}

// close(ch) vs ch <- true:
//   ch <- true   wakes up ONE receiver
//   close(ch)    wakes up ALL receivers
//
// <-chan struct{} in function signature = receive-only channel
//   The function can read from it but can't accidentally write.
//
// struct{} is the zero-size type — no data, no memory.
//   Used when you only care about the signal, not any value.
