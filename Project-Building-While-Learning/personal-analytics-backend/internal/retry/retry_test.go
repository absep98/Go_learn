package retry

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
