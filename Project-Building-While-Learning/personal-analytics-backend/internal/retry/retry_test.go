package retry

/*
=== RETRY PACKAGE TESTS ===

WHAT WE ARE TESTING:
Do(maxAttempts, initialDelay, operation) retries a function that returns an error.
We test 4 scenarios: success immediately, success after retries, all fail, zero attempts.

KEY CONCEPTS USED:

1. SAME-PACKAGE TESTS
   File is `package retry` (not `package retry_test`).
   This means we call Do() directly without any import — no retry.Do().
   Why? We are inside the package, so all exported and unexported symbols are visible.

2. CLOSURE AS MOCK
   We use a func() error closure to simulate the operation being retried.
   The closure captures `attempts` from the outer scope:
     attempts := 0
     operation := func() error {
         attempts++      ← modifies outer variable
         return nil
     }
   After Do() runs, we check `attempts` to verify how many times it was called.
   This is how we test behavior without any real network or database.

3. DELAY = 0 IN TESTS
   Do(3, 0, operation) — second argument is initialDelay.
   We pass 0 so tests run instantly. Real production code uses 100ms+.

4. ASSERTION PATTERN
   if err != nil  → we expected success, fail if there is an error
   if err == nil  → we expected failure, fail if there is no error
   if attempts != N → verify exactly how many times operation was called

Q: Why not just call Do() with a real HTTP request?
A: Tests must be deterministic — same result every time, no network needed.
   A closure gives full control: we decide what it returns and when.

Q: Why does delay=0 work?
A: Do() uses time.Sleep(delay). With delay=0, Sleep is instant.
   Exponential backoff doubles 0 each time — still 0. Tests stay fast.
*/

import (
	"fmt"
	"testing"
)

func TestDo_SuccessOnFirstTry(t *testing.T) {

	attempts := 0
	operation := func() error {
		attempts++
		return nil
	}

	err := Do(3, 0, operation)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}

}

func TestDo_SuccessOnThirdTry(t *testing.T) {

	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("not ready yet") // fail on attempts 1 and 2
		}
		return nil // succeed on attempt 3
	}

	err := Do(3, 0, operation)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempt, got %d", attempts)
	}

}

func TestDo_AllAttemptsFail(t *testing.T) {

	attempts := 0
	operation := func() error {
		attempts++
		return fmt.Errorf("not ready yet")
	}

	err := Do(3, 0, operation)

	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempt, got %d", attempts)
	}

}

func TestDo_ZeroAttempts(t *testing.T) {

	attempts := 0
	operation := func() error {
		attempts++
		return nil
	}

	err := Do(0, 0, operation)

	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if attempts != 0 {
		t.Errorf("expected 0 attempt, got %d", attempts)
	}
}
