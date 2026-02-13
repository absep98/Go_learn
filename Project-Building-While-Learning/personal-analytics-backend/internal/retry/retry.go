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
