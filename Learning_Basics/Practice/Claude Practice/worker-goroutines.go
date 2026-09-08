package main

import (
	"fmt"
	"sync"
)
/*
"Write a function that takes a slice of 20 integers, processes each one (square it)
using a pool of 3 worker goroutines, and returns all results. Don't worry about order."
*/


func square(id int, wg *sync.WaitGroup, jobs chan int, results chan int) {
	defer wg.Done()

	for job := range jobs {
		squared := job * job
		results <- squared
	}
}

func main() {

	jobs := make(chan int, 20)
	results := make(chan int, 20)
	var wg sync.WaitGroup

	for i := 0 ; i < 20 ; i++ {
		jobs <- i
	}

	// start 3 workers
	for w := 0 ; w < 3 ; w++ {
		wg.Add(1)
		go square(w, &wg, jobs, results)
	}

	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Println(result)
	}


}