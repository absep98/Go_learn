package main
import (
	"fmt"
	"sync"
)

func worker( id int, wg *sync.WaitGroup, pipeline chan int, results chan int) {
	
	defer wg.Done()

	for job := range pipeline {
		squared := job * job
		results <- squared
	}
}

// HOW THIS WORKER POOL WORKS:
//
// 1. BUFFERED CHANNELS: jobs and results are buffered (size 10), meaning we can
//    push up to 10 items without blocking. This lets us load all jobs upfront
//    before workers even start.
//
// 2. WHY WAITGROUP: We spawn 3 goroutines. Main has no way to know when they
//    finish unless we explicitly track them. WaitGroup is a counter:
//    - wg.Add(1) increments for each worker launched
//    - wg.Done() decrements when a worker's function returns (all its jobs done)
//    - wg.Wait() blocks until counter hits 0 (all workers finished)
//
// 3. WHY CLOSE(RESULTS): The "for result := range results" loop in main runs
//    forever until the channel is closed. Without close(results), main hangs
//    waiting for more data that will never come. We close it AFTER wg.Wait()
//    guarantees all workers are done sending.
//
// 4. WHY CLOSE(JOBS): Workers use "for job := range pipeline" which loops until
//    the channel is closed. Without close(jobs), workers block forever waiting
//    for more jobs. Closing signals "no more work coming."
//
// FLOW: load jobs → close jobs → workers process → wg.Wait() → close results → print
func main() {

	jobs := make(chan int, 10)

	results := make(chan int, 10)
	for i:=0 ; i < 10 ; i++ {
		jobs <- i
	}

	close(jobs)
	var wg sync.WaitGroup

	for i := 0 ; i < 3 ; i++ {
		wg.Add(1)

		go worker(i, &wg, jobs, results)
	}

	wg.Wait()
	close(results)

	for result := range results {
		fmt.Println( result)
	}
}
