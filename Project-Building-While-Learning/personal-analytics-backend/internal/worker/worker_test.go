package worker

/*
=== WORKER POOL SHUTDOWN TESTS ===

Graceful shutdown has to guarantee two things:

1. Every job that AddJob accepted is processed before DrainWorkers returns.
   (Original: each worker called wg.Add(1) only AFTER it received a job. When
   Wait() ran while the counter was momentarily 0 — before a worker picked up
   the next buffered job — DrainWorkers returned with jobs still queued, and
   main() went on to close the DB those jobs write their failures to.)

2. AddJob never panics, even if it races with shutdown.
   (Original: DrainWorkers closed JobQueue; a handler still running — e.g. one
   that outlived its request timeout — then hit "send on closed channel".)

Run with: go test -race ./internal/worker/
*/

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countingSender struct {
	delay time.Duration
	sent  atomic.Int32
}

func (s *countingSender) Send(url string, payload interface{}) error {
	time.Sleep(s.delay)
	s.sent.Add(1)
	return nil
}

func TestDrainWorkers_ProcessesEveryAcceptedJob(t *testing.T) {
	sender := &countingSender{delay: 10 * time.Millisecond}
	StartWorkerPool(1, "http://example.invalid/hook", sender)

	const jobs = 10
	for i := 0; i < jobs; i++ {
		AddJob("entry_created", map[string]interface{}{"entry_id": i})
	}
	DrainWorkers()

	if got := sender.sent.Load(); got != jobs {
		t.Fatalf("DrainWorkers returned after %d of %d accepted jobs were processed", got, jobs)
	}
}

func TestAddJob_AfterDrain_DropsInsteadOfPanicking(t *testing.T) {
	StartWorkerPool(1, "http://example.invalid/hook", &countingSender{})
	DrainWorkers()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AddJob after shutdown panicked: %v", r)
		}
	}()
	AddJob("entry_created", map[string]interface{}{"entry_id": 1})
}

func TestAddJob_ConcurrentWithDrain_NeitherPanicsNorRaces(t *testing.T) {
	StartWorkerPool(2, "http://example.invalid/hook", &countingSender{delay: time.Millisecond})

	var wg sync.WaitGroup
	var panics atomic.Int32
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() {
				if recover() != nil {
					panics.Add(1)
				}
			}()
			AddJob("entry_created", map[string]interface{}{"entry_id": i})
		}(i)
	}
	DrainWorkers()
	wg.Wait()

	if n := panics.Load(); n != 0 {
		t.Fatalf("%d AddJob calls panicked while the pool was draining", n)
	}
}
