package circuitbreaker

import (
	"fmt"
	"testing"
	"time"
)

func TestClosedToOpen(t *testing.T) {

	cb := NewCircuitBreaker(5, 30*time.Second)

	// call Execute 5 times with a failing operation
	for i := 0; i < 5; i++ {
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
		cb.Execute(func() error {
			return fmt.Errorf("service down")
		})
	}
	// 1 success

	cb.Execute(func() error {
		return nil
	})
	// 4 more failures
	for i := 0; i < 4; i++ {
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
