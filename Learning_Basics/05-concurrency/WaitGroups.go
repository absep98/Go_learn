/*
WaitGroups (sync.WaitGroup)
While channels are amazing for passing data around, sometimes you don't care about passing data—you just want to say: "Hey, I am spinning up 3 background download workers. I just want main to wait until all 3 are 100% finished before we print 'Success'."

For this, Go provides a counter utility called a WaitGroup inside the standard sync package.

Think of a WaitGroup like a Security Guard's Clicker Counter:

Every time you spin up a worker, you click the counter UP (wg.Add(1)).

When a worker finishes its task, it clicks the counter DOWN (wg.Done()).

The main function stands at a gate and refuses to move until the counter hits exactly 0 (wg.Wait())
*/

package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Worker %d starting job...\n", id)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Worker %d finished!\n", id)
}

func main() {
	var wg sync.WaitGroup
	for i := 1 ; i <= 3 ; i++ {
		wg.Add(1)   // increment the counter before spawning the goroutine

		go worker(i, &wg)

	}

	wg.Wait()
	fmt.Println("All workers finised. Main exiting safely!")
}