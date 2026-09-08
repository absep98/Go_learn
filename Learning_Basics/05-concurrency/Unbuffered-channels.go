package main

import "fmt"
/*
An unbuffered channel is the default channel you get when you write make(chan Type) without giving 
it a size number.

Think of an unbuffered channel as a strict game of catch or a face-to-face handshake.

The Goroutine sending data cannot just throw the ball and walk away. It must block and freeze 
until another Goroutine is actively standing there to catch it.

The Goroutine receiving data must freeze and wait until a sender actually drops a ball in.

Go
package main

func main() {
    ch := make(chan int) // Unbuffered channel created

    ch <- 42 // ❌ CRASH: fatal error: all goroutines are asleep - deadlock!
}
Why did it crash? Because main tried to send 42 into the pipeline, but there was no other 
background Goroutine running to catch it. main froze forever waiting for a catcher, and
 Go's runtime realized your program was permanently stuck.
*/

func fetchDatabaseID(pipe chan int) {
    id := 9923 // Imagine this took 2 seconds to fetch from a DB
    
    // Send the ID into the channel pipeline
    pipe <- id 
}

func main() {
    // 1. Create the pipeline
    idChannel := make(chan int)

    // 2. Spin up the background worker
    go fetchDatabaseID(idChannel)

    // 3. This line BLOCKS and waits until the worker drops the value in!
    receivedID := <-idChannel 

    fmt.Println("Successfully received ID from background worker:", receivedID)
}

/*


The Arrow Syntax (<-):
Go uses a beautiful arrow syntax to show data moving into and out of channels:

Send data INTO a channel: channel <- value (Arrow points into the channel)

Receive data OUT of a channel: variable := <-channel (Data flows out of the channel into your variable)

*/