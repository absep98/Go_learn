package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var unsafeCounter int64
	var atomicCounter int64
	var wg sync.WaitGroup

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsafeCounter++                     // BROKEN
			atomic.AddInt64(&atomicCounter, 1)  // SAFE
		}()
	}
	wg.Wait()

	fmt.Println("unsafe:", unsafeCounter)   // usually < 10000
	fmt.Println("atomic:", atomicCounter)   // always 10000
}