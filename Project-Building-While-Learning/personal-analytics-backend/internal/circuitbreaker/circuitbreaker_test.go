package circuitbreaker

/*
=== CIRCUIT BREAKER PACKAGE TESTS ===

WHAT WE ARE TESTING:
The circuit breaker has 3 states: closed (normal), open (blocking), half-open (testing).
We test 3 state transition scenarios.

KEY CONCEPTS USED:

1. SAME-PACKAGE TESTS
   `package circuitbreaker` — we can access private fields like cb.state directly.
   This is useful to verify internal state without adding exported getters.

2. ARRANGE → ACT → ASSERT PATTERN
   Every test follows this structure:
   - Arrange: create a breaker, put it in the desired state
   - Act: call Execute() and capture the result
   - Assert: check result matches expectation

3. TRIPPING THE BREAKER
   Tests that need an open breaker first cause 5 consecutive failures:
     for i := 0; i < 5; i++ {
         cb.Execute(func() error { return fmt.Errorf("down") })
     }
   After this, the next Execute() is rejected without calling the operation.

4. ATTEMPTS COUNTER TO PROVE OPERATION WAS NOT CALLED
   attempts := 0
   cb.Execute(func() error { attempts++; return nil })
   if attempts != 0 → proves operation was never touched (breaker rejected immediately)
   Same closure pattern as retry tests.

5. CHECKING SPECIFIC ERROR MESSAGE
   err.Error() != "circuit breaker is open"
   Not just checking err != nil — verifying the rejection came from the BREAKER,
   not from some other error in the operation.

Q: Why test that attempts==0 when open?
A: err!=nil alone proves something failed, but doesn't prove the operation
   was skipped. attempts==0 proves the breaker rejected WITHOUT calling the service.
   That fast-rejection is the whole point of a circuit breaker.

Q: Why 4 failures + success + 4 failures in TestClosedResetsOnSuccess?
A: We want to prove the counter resets on success, not just accumulates.
   4+4=8 total failures, but only 4 consecutive — breaker should stay closed.
   If reset wasn't working, it would trip on the 5th total failure.
*/

import (
	"fmt"
	"testing"
	"time"
)

func TestClosedToOpen(t *testing.T) {

	cb := NewCircuitBreaker(5, 30*time.Second)

	// call Execute 5 times with a failing operation
	for i := 0; i < 5; i++ {
		//nolint:errcheck // intentionally ignoring errors to trip the breaker
		cb.Execute(func() error {
			return fmt.Errorf("service down")
		})
	}

	err := cb.Execute(func() error {
		return nil
	})

	if err == nil {
		t.Errorf("expected circuit breaker error, got nil")
	}

	if err.Error() != "circuit breaker is open" {
		t.Errorf("expected 'circuit breaker is open', got %v", err)
	}
}

func TestOpenRejectsImmediately(t *testing.T) {

	cb := NewCircuitBreaker(5, 30*time.Second)

	for i := 0; i < 5; i++ {
		//nolint:errcheck // intentionally ignoring errors to trip the breaker
		cb.Execute(func() error {
			return fmt.Errorf("service down")
		})
	}

	// now call with a counter - operation should never be called
	attempts := 0
	err := cb.Execute(func() error {
		attempts++
		return nil
	})

	if err == nil {
		t.Errorf("expected circuit breaker error, got nil")
	}

	if attempts != 0 {
		t.Errorf("expected operation not to be called, got %d calls", attempts)
	}
}

func TestClosedResetsOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(5, 30*time.Second)

	// 4 failures

	for i := 0; i < 4; i++ {
		//nolint:errcheck // intentionally ignoring errors to trip the breaker
		cb.Execute(func() error {
			return fmt.Errorf("service down")
		})
	}
	// 1 success

	//nolint:errcheck // intentionally ignoring success result
	cb.Execute(func() error {
		return nil
	})
	// 4 more failures
	for i := 0; i < 4; i++ {
		//nolint:errcheck // intentionally ignoring errors to trip the breaker
		cb.Execute(func() error {
			return fmt.Errorf("service down")
		})
	}
	// final call — breaker should still be closed, operation runs
	err := cb.Execute(func() error {
		return nil
	})

	// assert: err is nil (breaker did NOT trip)
	if err != nil {
		t.Errorf("expected breaker to be closed, got: %v", err)
	}
}
