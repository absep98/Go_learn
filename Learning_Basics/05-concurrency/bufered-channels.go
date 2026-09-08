/*
Buffered Channels (The Mailbox)
A buffered channel is created by passing a capacity number as the second argument: 
make(chan Type, capacity).

Think of a buffered channel as a physical mailbox that can hold a specific number of letters.

If the mailbox can hold 3 letters, a Goroutine can drop 3 messages inside it, one after another,
 and instantly keep moving without waiting for anyone to read them.

The sender only freezes (blocks) if the mailbox becomes completely full.

The receiver only freezes if the mailbox becomes completely empty.
*/

package main

import (
	"fmt"
)

func main() {
	//  a buffered channel that can hold 2 items
	mailbox := make(chan string, 2)

	// we can send 2 items without any background workers catching them !
	mailbox <- "Message A"
	mailbox <- "Message B"

	// it tried adding another 3rd itme wihout any background workers catching them
	// it would crash because the mailbox would overflow

	// read them back
	fmt.Println(<-mailbox)
	fmt.Println(<-mailbox)
}