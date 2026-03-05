package retry

/*
=== RETRY WITH EXPONENTIAL BACKOFF ===

Problem: External services (Redis, DB, APIs) sometimes fail temporarily.
A single failure shouldn't crash the whole request. We should try again.

But HOW we retry matters:

=== WHY NOT JUST RETRY IMMEDIATELY? ===

Immediate retry = hammering a service that's already struggling.
If Redis is overloaded:
  Attempt 1: fail (Redis busy)
  Attempt 2: fail (immediately, Redis still busy!)
  Attempt 3: fail (still hammering!)
  → Made the problem WORSE by adding more load

=== EXPONENTIAL BACKOFF ===

Wait longer between each retry:
  Attempt 1: fail → wait 100ms
  Attempt 2: fail → wait 200ms  (doubled!)
  Attempt 3: fail → wait 400ms  (doubled again!)
  Attempt 4: fail → wait 800ms
  Attempt 5: give up

This gives the struggling service TIME TO RECOVER.

=== HIGHER-ORDER FUNCTIONS ===

Do() takes a FUNCTION as an argument:

  retry.Do(3, 100*time.Millisecond, func() error {
      return redis.Client.Set(ctx, key, value, ttl).Err()
  })

The func() error is a "closure" — it captures variables from the
surrounding scope (ctx, key, value, ttl). Do() doesn't know or care
WHAT operation you're retrying. It just calls your function and
checks if it returns an error.

This is the power of higher-order functions: write retry logic ONCE,
use it for ANY operation.
*/

import (
	"fmt"
	"log/slog"
	"time"
)

/*
=== INTERVIEW ANSWER: RETRY WITH EXPONENTIAL BACKOFF ===

WHAT:
Retry handles temporary failures — network blip, service briefly overloaded,
container still starting. Instead of failing immediately, try again a few times
before giving up. Key word: TEMPORARY. Never retry permanent errors (wrong password,
invalid key) — you'll just waste time getting the same error repeatedly.

WHY EXPONENTIAL BACKOFF (not immediate retry):
Immediate retry hammers a struggling service and makes it worse.
Exponential backoff gives the service breathing room to recover:
  Attempt 1 fails → wait 100ms
  Attempt 2 fails → wait 200ms
  Attempt 3 fails → wait 400ms → give up
Each wait doubles, so the service gets progressively more time to recover.

THUNDERING HERD PROBLEM:
If many clients retry at the same exact time (e.g. all retry after exactly 1s),
they all hit the service simultaneously when it recovers and knock it back down.
Exponential backoff with jitter (random offset added to delay) is the production
solution. Our implementation uses pure exponential — simple and good enough for
startup retries.

WHERE WE USE IT:
Only in InitRedis() (startup), NOT in cache.Get/Set/Delete (per-request).
- Startup: Redis might still be booting. A few retries over seconds is fine,
  no user is waiting.
- Per-request: Retrying in cache.Get() means every user request to a down Redis
  waits 500ms+1s+2s = 3.5s. Terrible UX.
- Runtime Redis failures are handled by the circuit breaker instead — fast rejection.

HIGHER-ORDER FUNCTION:
Do() takes func() error so it's generic. The same retry logic works for Redis ping,
DB connection, HTTP call — anything that can fail and succeed on retry.
*/

// Do retries the given operation with exponential backoff
//
// Parameters:
//   - maxAttempts: how many times to try (e.g., 3 means: try, retry, retry)
//   - initialDelay: wait time after first failure (doubles each retry)
//   - operation: the function to retry (must return error or nil)
//
// Returns:
//   - nil if operation succeeds on any attempt
//   - error from the last attempt if all attempts fail
//
// Example usage:
//
//	err := retry.Do(3, 100*time.Millisecond, func() error {
//	    return redis.Client.Ping(ctx).Err()
//	})
func Do(maxAttempts int, initialDelay time.Duration, operation func() error) error {
	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Try the operation
		lastErr = operation()
		if lastErr == nil {
			// Success! If this wasn't the first attempt, log the recovery
			if attempt > 1 {
				slog.Info("Operation succeeded after retry",
					"attempt", attempt,
					"total_attempts", maxAttempts,
				)
			}
			return nil // Success!
		}

		// Last attempt failed — don't wait, just return error
		if attempt == maxAttempts {
			break
		}

		// Log the retry
		slog.Warn("Operation failed, retrying",
			"attempt", attempt,
			"max_attempts", maxAttempts,
			"next_delay", delay.String(),
			"error", lastErr,
		)

		// Wait before next attempt
		time.Sleep(delay)

		// Double the delay for next retry (exponential backoff)
		delay *= 2
	}

	// All attempts failed
	return fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
}
