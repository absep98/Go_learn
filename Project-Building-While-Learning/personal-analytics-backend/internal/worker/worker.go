package worker

import (
	"log/slog"
	"personal-analytics-backend/internal/circuitbreaker"
	"personal-analytics-backend/internal/db"
	"personal-analytics-backend/internal/retry"
	"personal-analytics-backend/internal/webhook"
	"sync"
	"time"
)

// ========================================
// JOB DEFINITION
// ========================================

// Job represents a background task to be processed
// Think of it as an "order ticket" in a pizza shop
type Job struct {
	Type    string      // What kind of job? "entry_created", "email", etc.
	Payload interface{} // The data for this job (entry ID, user ID, etc.)
}

// ========================================
// WORKER POOL
// ========================================

// JobQueue is the channel where jobs are sent
// Think of it as the "order counter" where tickets pile up
var JobQueue chan Job

var WebhookBreaker = circuitbreaker.NewCircuitBreaker(5, 30*time.Second)

// package level variable (near JobQueue and WebhookBreaker):
var webhookURL string

var sender webhook.Sender // package-level, holds whatever Sender is injected

// wg counts running WORKERS (not jobs). DrainWorkers waits for all of them to exit.
var wg sync.WaitGroup

// queueMu guards JobQueue against being closed while AddJob sends on it.
// AddJob holds the read lock (many handlers can enqueue at once);
// DrainWorkers takes the write lock to flip queueClosed and close the channel.
var (
	queueMu     sync.RWMutex
	queueClosed bool
)

// StartWorkerPool starts N workers that listen for jobs
// Each worker is like a "chef" waiting for orders
func StartWorkerPool(numWorkers int, url string, s webhook.Sender) {
	queueMu.Lock()
	defer queueMu.Unlock()

	// Create the job queue (buffered channel with capacity 100)
	// Buffer = how many jobs can wait in line before blocking
	JobQueue = make(chan Job, 100)
	queueClosed = false

	webhookURL = url

	sender = s

	// FIX: Add BEFORE starting the goroutines. The first version called
	// wg.Add(1) inside the worker after it received a job, so Wait() could
	// observe a zero counter while jobs were still buffered and return early
	// (and Add-during-Wait is a WaitGroup misuse the runtime panics on).
	wg.Add(numWorkers)
	for i := 1; i <= numWorkers; i++ {
		go worker(i) // "go" = run in background
	}

	slog.Info("Worker pool started", "num_workers", numWorkers, "queue_capacity", 100)
}

// worker is a single worker that processes jobs from the queue.
// It exits only after JobQueue is closed AND every buffered job is drained.
func worker(id int) {
	defer wg.Done()

	// "range JobQueue" = wait for next job, then process it.
	// When JobQueue is closed, the loop still delivers every buffered job, then exits.
	for job := range JobQueue {
		slog.Info("Worker processing job", "worker_id", id, "job_type", job.Type)

		// Process the job based on its type
		processJob(job)

		slog.Info("Worker completed job", "worker_id", id, "job_type", job.Type)
	}
}

// processJob handles different job types
func processJob(job Job) {
	switch job.Type {
	case "entry_created":
		// Simulate sending notification / updating analytics
		// In real app: send email, update stats, notify webhooks, etc.
		slog.Debug("Processing entry creation", "payload", job.Payload)
		err := WebhookBreaker.Execute(func() error {
			return retry.Do(3, 500*time.Millisecond, func() error {
				return sender.Send(webhookURL, job.Payload)
			})
		})

		if err != nil {
			deadLetter(job, err)
		}

	case "entry_deleted":
		slog.Debug("Processing entry deletion", "payload", job.Payload)
		err := WebhookBreaker.Execute(func() error {
			return retry.Do(3, 500*time.Millisecond, func() error {
				return sender.Send(webhookURL, job.Payload)
			})
		})

		if err != nil {
			deadLetter(job, err)
		}

	default:
		slog.Warn("Unknown job type", "job_type", job.Type)
	}
}

// deadLetter persists a job that exhausted its retries so an operator can
// inspect or replay it via /failed-jobs. If even that insert fails, the job is
// lost — log it loudly with enough context to reconstruct it by hand.
func deadLetter(job Job, jobErr error) {
	slog.Error("job failed, saving to dead letter queue", "job_type", job.Type, "error", jobErr)
	if err := db.InsertFailedJob(job.Type, job.Payload, jobErr.Error()); err != nil {
		slog.Error("dead-letter insert failed, job lost",
			"job_type", job.Type, "payload", job.Payload, "job_error", jobErr, "error", err)
	}
}

// DrainWorkers stops accepting jobs, then blocks until every job already
// accepted has been processed. Safe to call more than once.
func DrainWorkers() {
	queueMu.Lock()
	if queueClosed {
		queueMu.Unlock()
		wg.Wait()
		return
	}
	queueClosed = true
	close(JobQueue) // no AddJob can be mid-send: they all hold the read lock
	queueMu.Unlock()

	wg.Wait() // workers exit once the buffer is empty
}

// ========================================
// HELPER TO ADD JOBS
// ========================================

/*
=== INTERVIEW ANSWER: WORKER POOL ===

WHAT:
A worker pool is a fixed number of goroutines that process background tasks
from a shared queue. Slow work (emails, webhooks, analytics) runs in background
so the HTTP handler can return immediately.

WHY:
Go's HTTP server already spins up one goroutine per request — concurrency for
handling requests is automatic. The worker pool solves a different problem:
some tasks are slow and the user shouldn't wait for them.

Without worker pool:
  POST /entries → save to DB (5ms) → send email (2000ms) → return 201
  User waits 2 seconds. Bad UX.

With worker pool:
  POST /entries → save to DB (5ms) → AddJob (instant) → return 201
  Worker picks up job in background → sends email
  User gets response in 5ms. Email sends behind the scenes.

HOW:
- Job struct: ticket with Type (what to do) and Payload (data needed)
- JobQueue: buffered channel (size 100) — the handoff point between HTTP goroutines and workers
- StartWorkerPool(n): spins up n goroutines, each running `for job := range JobQueue`
- range on a channel BLOCKS when empty — worker sleeps with zero CPU until a job arrives
- AddJob: non-blocking send using select+default — if queue full, drop job and log warning

WHY CAP THE WORKERS:
HTTP goroutines are short-lived (milliseconds) — Go handles thousands fine.
Worker goroutines doing outbound calls hold connections open for seconds.
10,000 simultaneous email sends = 10,000 connections to email server → crashes it.
The cap protects downstream services, not Go itself.

WHY NON-BLOCKING AddJob:
If send blocked on full queue, the HTTP handler would hang waiting for space.
That defeats the purpose. Better to drop a background job than block a user request.

CACHE INVALIDATION ON WRITE:
CreateEntry deletes the cached GET /entries result after insert.
Reason: cached list is now stale (missing the new entry).
Next GET will miss cache, fetch fresh from DB, re-cache updated list.
Pattern: every write must invalidate related cache keys.

LOAD SHEDDING (what happens when queue is full):
AddJob uses select+default — non-blocking channel send:

  select {
  case JobQueue <- job:   // space available → enqueue
  default:                // full → drop immediately, log warning
  }

Why drop instead of block?
AddJob is called INSIDE the HTTP handler. If it blocked, the user's HTTP response
would also block — defeating the entire purpose of background processing.
The HTTP handler must return fast. The webhook is a best-effort side effect:
the entry IS saved to DB (that's the core operation), webhook delivery is optional.

Load shedding options compared:
  Drop (current):       log warning, discard job     — webhook missed, user unaffected
  Block:                wait for queue space         — HTTP handler hangs, bad UX
  Grow queue unbounded: always accept                — OOM crash under sustained load
  Persist to DB:        save failed jobs for retry   — no data loss, production choice

GRACEFUL SHUTDOWN ORDER:
  1. OS sends SIGINT (Ctrl+C) → received on quit channel
  2. server.Shutdown(ctx)     → stop accepting NEW HTTP requests
                                wait for in-flight HTTP handlers to finish (5s timeout)
  3. worker.DrainWorkers()    → close(JobQueue) stops new jobs being picked up
                                wg.Wait() blocks until all running workers finish

Why server.Shutdown() BEFORE DrainWorkers()?
  HTTP handlers call AddJob(). If we drained workers first, an HTTP handler
  could still be running and call AddJob() on a closed channel → panic.
  We must stop the HTTP layer (the job producer) before stopping the workers
  (the job consumers). Producer must stop before consumer stops.

Why close(JobQueue) in DrainWorkers()?
  Workers sit in `for job := range JobQueue` — they block waiting for jobs.
  Closing the channel unblocks all idle workers and signals them to exit.
  wg.Wait() then waits for currently-running workers to finish their job.

TRADE-OFFS:
- Jobs dropped when queue full — acceptable for non-critical background work
- No retry on failed jobs — production would use a persistent queue (Redis, RabbitMQ)

UPDATE — "producer stops before consumer" was not enough (found with go test -race):
server.Shutdown() only waits for handlers that are still attached to a
connection. A handler that outlived its request timeout keeps running in the
background and can call AddJob AFTER close(JobQueue) → "send on closed channel"
panic. AddJob now checks queueClosed under a read lock and drops the job
instead; DrainWorkers takes the write lock to close. See worker_test.go.
*/

// AddJob adds a new job to the queue
// This is what handlers call to schedule background work
func AddJob(jobType string, payload interface{}) {
	queueMu.RLock()
	defer queueMu.RUnlock()

	if queueClosed || JobQueue == nil {
		slog.Warn("Worker pool not accepting jobs (shutting down), dropping job", "job_type", jobType)
		return
	}

	// Non-blocking send (if queue is full, log warning)
	select {
	case JobQueue <- Job{Type: jobType, Payload: payload}:
		slog.Info("Job added to queue", "job_type", jobType)
	default:
		slog.Warn("Job queue full, dropping job", "job_type", jobType)
	}
}
