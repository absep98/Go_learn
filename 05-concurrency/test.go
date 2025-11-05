package main

import (
	"fmt"
	"sync"
)

func printInit(n, value int, wg *sync.WaitGroup) {
	for i := 1; i <= n; i++ {
		fmt.Printf("%d ", value)
	}
	fmt.Println()
	defer wg.Done()
}

func main() {
	fmt.Println("=== Go Goroutines ===")
	var wg sync.WaitGroup
	wg.Add(5)

	go printInit(5, 1, &wg)
	go printInit(5, 2, &wg)
	go printInit(5, 3, &wg)
	go printInit(5, 4, &wg)
	go printInit(5, 5, &wg)

	wg.Wait()
}
