package circuitbreaker

/*
=== CIRCUIT BREAKER PATTERN ===

Named after the electrical MCB in your home's fuse box.
Just like an MCB cuts power when there's a short circuit to protect wiring,
a software circuit breaker cuts requests when a service fails to protect the server.

=== THE 3 STATES ===

CLOSED (normal operation):
  - All requests flow through to the external service (Redis, DB, etc.)
  - Failures are counted. If failures reach threshold → switch to OPEN
  - On success: reset failure count (consecutive failures matter, not total)
  - Like MCB when everything is fine: electricity flows normally

OPEN (tripped / blocking):
  - External service is assumed DOWN
  - ALL requests are rejected IMMEDIATELY without trying the service
  - Why? Because if Redis is down, hammering it with requests makes it worse
  - After cooldown period (e.g. 30s) → switch to HALF-OPEN to test
  - Like MCB after trip: power is cut, nothing flows

HALF-OPEN (testing / recovery):
  - Cooldown has passed, service MIGHT have recovered
  - Allows ONE test request through
  - If test succeeds → service is back → switch to CLOSED (normal)
  - If test fails → still down → switch back to OPEN (another cooldown)
  - Like manually flipping MCB back: test if problem is fixed

=== WHY THIS MATTERS ===

Without circuit breaker (Redis is down):
  Request 1: waits 5s timeout ❌
  Request 2: waits 5s timeout ❌
  Request 100: waits 5s timeout ❌
  → 100 goroutines stuck, server crawls, users wait 5s for every request

With circuit breaker (Redis is down):
  Request 1-5: waits 5s timeout ❌ (circuit opens on 5th failure)
  Request 6-100: instant fail <1ms ✅ (breaker rejects without trying)
  → Only 5 goroutines stuck, 95 free, users get fast error instead of long wait
*/

import (
	"fmt"
	"sync"
	"time"
)

// CircuitBreaker tracks the health of an external service and
// stops sending requests when the service appears to be down
type CircuitBreaker struct {
	state           string        // "closed", "open", or "half-open"
	failureCount    int           // consecutive failures in current window
	lastFailureTime time.Time     // when the last failure happened (used for cooldown)
	threshold       int           // how many consecutive failures before opening
	cooldown        time.Duration // how long to wait in OPEN state before testing
	mu              sync.Mutex    // protects all fields — multiple goroutines may call Execute()
}

// NewCircuitBreaker creates a new breaker.
// threshold: number of consecutive failures before tripping (e.g. 5)
// cooldown: how long to stay OPEN before testing recovery (e.g. 30s)
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     "closed", // Start in normal operation
		threshold: threshold,
		cooldown:  cooldown,
	}
}

// Execute runs the given operation through the circuit breaker.
// The operation is only called if the breaker allows it (not OPEN).
// Automatically tracks failures and manages state transitions.
//
// Why lock the mutex at the start?
// Multiple goroutines (one per request) call Execute() at the same time.
// Without a lock, two goroutines could both read state="closed" and both
// increment failureCount simultaneously — leading to wrong counts (race condition).
// The mutex ensures only ONE goroutine reads and updates state at a time.
func (cb *CircuitBreaker) Execute(operation func() error) error {

	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case "closed":
		err := operation()
		if err != nil {
			cb.failureCount++
			cb.lastFailureTime = time.Now()

			// Too many consecutive failures? Trip the breaker!
			if cb.failureCount >= cb.threshold {
				cb.state = "open"
			}
			return err
		}
		// Success — reset counter (we care about CONSECUTIVE failures, not total)
		cb.failureCount = 0
		return nil

	case "open":
		// Why check time.Since(lastFailureTime)?
		// We don't want to stay OPEN forever — the service might have recovered.
		// We wait for `cooldown` duration after the LAST failure, then test once.
		// time.Since() returns how much time has passed since lastFailureTime.
		// If that duration exceeds cooldown, enough time has passed to try again.
		if time.Since(cb.lastFailureTime) > cb.cooldown {
			cb.state = "half-open"
			// Test one request — same logic as half-open
			err := operation()
			if err != nil {
				// Still failing — reset cooldown timer and stay OPEN
				cb.state = "open"
				cb.lastFailureTime = time.Now()
				return err
			}
			// Recovered! Back to normal operation
			cb.state = "closed"
			cb.failureCount = 0
			return nil
		}
		// Cooldown not yet passed — reject immediately without calling service
		return fmt.Errorf("circuit breaker is open")

	case "half-open":
		// This case is reached only if Execute() is somehow called when state
		// is already half-open (shouldn't happen given our locking, but defensive)
		err := operation()
		if err != nil {
			cb.state = "open"
			cb.lastFailureTime = time.Now()
			return err
		}
		cb.state = "closed"
		cb.failureCount = 0
		return nil
	}
	return nil
}

