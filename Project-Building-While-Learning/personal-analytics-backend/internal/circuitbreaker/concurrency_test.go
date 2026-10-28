package circuitbreaker

/*
=== CIRCUIT BREAKER CONCURRENCY TESTS ===

The first version held cb.mu for the whole of Execute — including the call to
operation(). For the webhook breaker that operation is retry.Do (sleeps of
500ms + 1s) plus an HTTP POST, and all three workers share one breaker. So the
"worker pool of 3" could only ever send ONE webhook at a time.

1. CLOSED: independent calls must be able to run concurrently.
2. HALF-OPEN: exactly one trial call goes through; concurrent callers are
   rejected immediately instead of queueing behind the trial.
*/

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClosed_OperationsRunConcurrently(t *testing.T) {
	cb := NewCircuitBreaker(5, 30*time.Second)
	const callers = 3

	var inside sync.WaitGroup
	inside.Add(callers)
	allInside := make(chan struct{})
	go func() { inside.Wait(); close(allInside) }()

	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.Execute(func() error { //nolint:errcheck
				inside.Done()
				select {
				case <-allInside:
					return nil
				case <-time.After(2 * time.Second):
					return errors.New("gave up waiting for the other callers")
				}
			})
		}()
	}

	select {
	case <-allInside:
	case <-time.After(time.Second):
		t.Errorf("only one operation could run at a time: Execute serializes callers")
	}
	wg.Wait()
}

func TestHalfOpen_AdmitsOneTrialAndRejectsConcurrentCallers(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	cb.Execute(func() error { return errors.New("down") }) //nolint:errcheck // trips to OPEN
	time.Sleep(20 * time.Millisecond)                      // cooldown elapsed → next call is the trial

	trialStarted := make(chan struct{})
	releaseTrial := make(chan struct{})
	trialDone := make(chan error, 1)
	go func() {
		trialDone <- cb.Execute(func() error {
			close(trialStarted)
			<-releaseTrial
			return nil
		})
	}()
	<-trialStarted

	var secondRan atomic.Bool
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- cb.Execute(func() error { secondRan.Store(true); return nil })
	}()

	select {
	case err := <-secondDone:
		if err == nil || secondRan.Load() {
			t.Errorf("concurrent caller during half-open trial: want rejection, got err=%v ran=%v", err, secondRan.Load())
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("concurrent caller blocked behind the half-open trial instead of being rejected")
	}

	close(releaseTrial)
	if err := <-trialDone; err != nil {
		t.Fatalf("trial call: %v", err)
	}
	if err := cb.Execute(func() error { return nil }); err != nil {
		t.Errorf("after a successful trial the breaker should be CLOSED, got %v", err)
	}
}
