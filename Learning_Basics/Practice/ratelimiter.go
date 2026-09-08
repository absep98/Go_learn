package main

import (
	"fmt"
	"time"
)

type RateLimiter struct {
	requests	map[string][]time.Time
}

func (rl *RateLimiter) Allow(userID string) bool {
	
	now := time.Now()

	var recent []time.Time

	for _, t := range rl.requests[userID] {
		if time.Since(t) < 10*time.Second {
			recent = append(recent, t)
		}
	}

	rl.requests[userID] = recent

	if len(recent) < 3 {
		rl.requests[userID] = append(rl.requests[userID], now)
		return true
	}

	return false
}


func main() {

	rl := &RateLimiter{requests: make(map[string][]time.Time)}

	for i := 0 ; i < 5 ; i++ {
		fmt.Printf("Request %d: %v\n", i+1, rl.Allow("alice"))
	}

}