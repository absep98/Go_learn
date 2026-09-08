package main

import (
	"fmt"
	"bufio"
	"os"
	"time"
	"strconv"
)

func printGo(n int) {
	for i := 1 ; i <= n ; i++ {
		fmt.Println(i)
		time.Sleep(100*time.Millisecond)
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	input := scanner.Text()
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Errorf("failed to parse '%s' as integer: %v", input, err)
	}


	fmt.Println("Printing numbers started in background")
	go printGo(num)
	fmt.Println("Continued the execution further so")
	time.Sleep(500*time.Millisecond)
	fmt.Println("We put a tiny pause here just so the program doesn't exit before the background job prints")
}