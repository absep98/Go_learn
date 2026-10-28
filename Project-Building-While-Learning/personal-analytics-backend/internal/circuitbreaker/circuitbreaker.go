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
	"errors"
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
	generation      uint64        // bumped on every state change; results from an older generation are ignored
}

// ErrOpen is returned without calling the operation while the breaker is open
// (or while a half-open trial call is already in flight).
var ErrOpen = errors.New("circuit breaker is open")

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
// The circuit breaker is generic — it works for any operation that can succeed or fail. The caller decides what the operation is.
//
// FIX: the first version held the mutex for the whole call, INCLUDING operation().
// For webhooks that is retry.Do (500ms + 1s of backoff sleeps) plus an HTTP POST,
// and all workers share one breaker — so the worker pool could only send one
// webhook at a time. Now the lock covers only the state check before the call
// and the state update after it; operation() runs unlocked and concurrently.
// See concurrency_test.go.
func (cb *CircuitBreaker) Execute(operation func() error) error {
	gen, err := cb.beforeCall()
	if err != nil {
		return err
	}

	err = operation() // NOT holding cb.mu — independent calls run in parallel

	cb.afterCall(gen, err)
	return err
}

// beforeCall decides, under the lock, whether this call may go through.
func (cb *CircuitBreaker) beforeCall() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case "open":
		// We don't want to stay OPEN forever — the service might have recovered.
		// We wait for `cooldown` after the LAST failure, then let ONE call test it.
		if time.Since(cb.lastFailureTime) <= cb.cooldown {
			return 0, ErrOpen // cooldown not yet passed — reject without calling the service
		}
		cb.setState("half-open") // this caller is the single trial
		return cb.generation, nil

	case "half-open":
		// A trial is already in flight. Reject instead of queueing behind it.
		return 0, ErrOpen
	}
	return cb.generation, nil // closed
}

// afterCall records the outcome, under the lock.
func (cb *CircuitBreaker) afterCall(gen uint64, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// The state changed while this call was in flight (e.g. other calls tripped
	// the breaker). A stale result must not re-close or re-open it.
	if gen != cb.generation {
		return
	}

	if err != nil {
		cb.lastFailureTime = time.Now()
		if cb.state == "half-open" {
			cb.setState("open") // trial failed — still down, start another cooldown
			return
		}
		cb.failureCount++
		if cb.failureCount >= cb.threshold {
			cb.setState("open") // too many consecutive failures — trip the breaker
		}
		return
	}

	// Success — reset counter (we care about CONSECUTIVE failures, not total)
	if cb.state == "half-open" {
		cb.setState("closed") // trial succeeded — recovered
	}
	cb.failureCount = 0
}

func (cb *CircuitBreaker) setState(s string) {
	cb.state = s
	cb.generation++
}

/*
=== INTERVIEW ANSWER: CIRCUIT BREAKER ===

WHAT:
A circuit breaker tracks the health of an external service and stops sending
requests when that service is down.

WHY:
Redis is our cache layer. Without a circuit breaker, if Redis goes down, every
request still tries to connect, waits for a timeout (~5s), and ties up a goroutine.
With 100 concurrent requests, that's 100 goroutines stuck waiting — server crawls.
After 5 failures the circuit opens. Requests 6-100 get an instant rejection — no
waiting, no wasted goroutines.

HOW (3 states):
- Closed:    normal operation, all requests go through, failures are counted
- Open:      threshold hit, all requests rejected immediately, cooldown starts
- Half-open: cooldown passed, ONE test request allowed through
             success → closed | failure → back to open

IMPLEMENTATION DECISIONS:
- Execute() takes func() error (higher-order function) so the breaker is generic —
  it doesn't know about Redis, HTTP, or anything. Works for any external call.
- cache.Get() needs the Redis result string out, but Execute() only returns error.
  Solution: closure — declare result in outer scope, inner func writes to it directly.
- sync.Mutex protects internal state (failureCount, state) from race conditions
  when multiple goroutines call Execute() simultaneously.

TRADE-OFFS (what production would do differently):
- Use separate mutexes for read vs write (sync.RWMutex) for better performance
- Expose state via metrics endpoint so you can observe when circuit is open
- Per-service breakers with different thresholds (DB vs Redis vs external API)
- Use a library like sony/gobreaker for battle-tested edge case handling
*/
