package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var balance int64 = 100
	var wg sync.WaitGroup

	// 5 goroutines each try to withdraw 50
	// Only 2 should succeed (100 / 50 = 2)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// BUG: check and modify are separate operations
			// Gap between Load and Add allows multiple goroutines
			// to pass the check before any of them subtract
			if atomic.LoadInt64(&balance) >= 50 {
				atomic.AddInt64(&balance, -50)
				fmt.Printf("goroutine %d: withdrew 50\n", id)
			} else {
				fmt.Printf("goroutine %d: insufficient funds\n", id)
			}
		}(i)
	}
	wg.Wait()

	fmt.Println("final balance:", balance) // can go NEGATIVE!
}

// This is a TOCTOU bug (Time Of Check To Time Of Use).
// Each atomic call is safe individually, but the GAP
// between them lets multiple goroutines slip through.
//
// Fix: use a sync.Mutex to make check+subtract one unit,
// or use atomic.CompareAndSwap for lock-free approach.
