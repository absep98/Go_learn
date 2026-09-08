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
			unsafeCounter++                    // BROKEN: read-modify-write is 3 steps
			atomic.AddInt64(&atomicCounter, 1) // SAFE: single CPU instruction
		}()
	}
	wg.Wait()

	fmt.Println("unsafe:", unsafeCounter) // usually < 10000
	fmt.Println("atomic:", atomicCounter) // always 10000
}

// KEY CONCEPTS:
//
// counter++ is THREE cpu steps: READ, ADD, WRITE
// Two goroutines can READ the same value, both ADD 1,
// both WRITE — losing one increment.
//
// atomic.AddInt64 does all three as ONE indivisible
// CPU instruction (LOCK XADD on x86). Hardware guarantees
// no other core can interfere.
//
// Why int64 not int? sync/atomic needs exact size to pick
// the right CPU instruction. Plain int varies by platform.
//
// RUN WITH: go run -race race_demo.go
// The -race flag will catch the unsafe counter's data race.
