package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done() // ALWAYS first line — guaranteed to run
	fmt.Printf("worker %d starting\n", id)
	time.Sleep(1 * time.Second)
	fmt.Printf("worker %d done\n", id)
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)           // BEFORE go — in parent goroutine
		go worker(i, &wg)   // &wg = pointer, so all share same WaitGroup
	}

	fmt.Println("all workers spawned, waiting...")
	wg.Wait() // blocks until counter hits 0
	fmt.Println("all workers finished")
}

// WaitGroup contract:
//   Parent:  wg.Add(1)      — BEFORE spawning
//   Child:   defer wg.Done() — FIRST line in goroutine
//   Parent:  wg.Wait()      — blocks until counter = 0
//
// Three bugs to avoid:
//   1. wg.Add inside goroutine → race with wg.Wait
//   2. Forget wg.Add → Wait returns immediately, goroutines killed
//   3. Forget wg.Done → Wait blocks forever (deadlock/hang)
//
// Why &wg not wg?
//   Passing wg by value copies it. The goroutine's Done()
//   would decrement a COPY, not the original. Wait() never
//   sees it and hangs forever.
//
// TRY: change "go worker(i, &wg)" to "worker(i, &wg)"
//   (remove "go") — now it takes 5 seconds instead of 1.
