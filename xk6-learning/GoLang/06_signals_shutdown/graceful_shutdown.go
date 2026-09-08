package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func acceptLoop(ln net.Listener, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("acceptLoop exiting:", err)
			return // ln.Close() caused this error → exit cleanly
		}
		fmt.Println("new connection from:", conn.RemoteAddr())
		conn.Close() // just accept and close for demo
	}
}

func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Println("listen failed:", err)
		return
	}
	fmt.Println("listening on :9090 — press Ctrl+C to stop")

	var wg sync.WaitGroup
	wg.Add(1)
	go acceptLoop(ln, &wg)

	// Intercept Ctrl+C
	sigCh := make(chan os.Signal, 1) // buffer=1 so signal isn't dropped
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh // BLOCKS here until Ctrl+C
	fmt.Println("\nreceived:", sig)

	// Cooperative shutdown chain:
	//   ln.Close() → Accept() returns error → acceptLoop returns → wg.Done()
	ln.Close()
	wg.Wait()
	fmt.Println("shutdown complete")
}

// Shutdown chain:
//   Ctrl+C → SIGINT → sigCh → <-sigCh unblocks
//   → ln.Close() → Accept() errors → acceptLoop returns
//   → defer wg.Done() → wg.Wait() unblocks → main exits
//
// You don't kill goroutines. You remove what they're waiting on.
//
// TRY: run this, then in another terminal:
//   echo "hi" | nc localhost 9090
//   Then press Ctrl+C in this terminal.
