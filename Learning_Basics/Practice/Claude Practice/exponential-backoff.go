package main

import (
	"fmt"
	"time"
)

// retry implements exponential backoff retry logic.
// LEARNING:
// - Use fmt.Errorf() for formatted error strings, NOT errors.New() (which only takes one argument)
// - Always specify time units explicitly: time.Second, time.Millisecond, etc.
// - Without units, time.Duration defaults to nanoseconds (1ns, 2ns, 4ns = invisible!)
// - Never create errors you don't use — every line must have a purpose
func retry(op func() error, max_attempt int) error {

	attempt_count := 0
	exponent_time := 1
	for i := 0; i < max_attempt; i++ {
		err := op()
		if err != nil {
			attempt_count++
			exponent_time *= 2
			// IMPORTANT: Always include * time.Second to sleep in seconds, not nanoseconds
			time.Sleep(time.Duration(exponent_time) * time.Second)
		} else {
			break
		}
	}

	if attempt_count == max_attempt {
		return fmt.Errorf("Failed after 3 retries.")
	}

	return nil

}

// alwaysFails is a test function that always returns an error.
func alwaysFails() error {
	return fmt.Errorf("boom")
}

func main() {

	err := retry(alwaysFails, 3)
	// LEARNING: Control program flow with if/else
	// - if err != nil: operation failed, print error
	// - else: operation succeeded, print success message
	// DON'T print "Success!" unconditionally — it will print even on failure!
	if err != nil {
		fmt.Printf("Failed at retrying %v\n", err)
	} else {
		fmt.Println("Success!")
	}
}
