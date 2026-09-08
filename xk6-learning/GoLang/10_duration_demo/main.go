package main

import (
	"flag"
	"fmt"
)

func main() {
	timeout := flag.Duration("timeout", 0, "test duration")
	flag.Parse()
	fmt.Println("You entered:", *timeout)
	fmt.Println("As seconds:", timeout.Seconds())
}

/*
PS C:\learning\GoLang\10_duration_demo>   go run main.go -timeout 60s
You entered: 1m0s
As seconds: 60
*/
