package worker

import (
	"log/slog"
	"personal-analytics-backend/internal/circuitbreaker"
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

// This is for waitgroup of workers
var wg sync.WaitGroup

// StartWorkerPool starts N workers that listen for jobs
// Each worker is like a "chef" waiting for orders
func StartWorkerPool(numWorkers int, url string, s webhook.Sender) {
	// Create the job queue (buffered channel with capacity 100)
	// Buffer = how many jobs can wait in line before blocking
	JobQueue = make(chan Job, 100)

	webhookURL = url

	sender = s
	// Start the workers (each runs in its own goroutine)
	for i := 1; i <= numWorkers; i++ {
		go worker(i) // "go" = run in background
	}

	slog.Info("Worker pool started", "num_workers", numWorkers, "queue_capacity", 100)
}

// worker is a single worker that processes jobs from the queue
// It runs forever, waiting for jobs
func worker(id int) {
	// This loop runs FOREVER (until program exits)
	for job := range JobQueue {
		// "range JobQueue" = wait for next job, then process it
		// When JobQueue is closed, this loop exits

		// increment: job starting
		wg.Add(1)
		slog.Info("Worker processing job", "worker_id", id, "job_type", job.Type)

		// Process the job based on its type
		processJob(job)

		slog.Info("Worker completed job", "worker_id", id, "job_type", job.Type)

		// Worker is done with job
		wg.Done()
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
			slog.Error(err.Error())
		}

	case "entry_deleted":
		slog.Debug("Processing entry deletion", "payload", job.Payload)
		err := WebhookBreaker.Execute(func() error {
			return retry.Do(3, 500*time.Millisecond, func() error {
				return sender.Send(webhookURL, job.Payload)
			})
		})

		if err != nil {
			slog.Error(err.Error())
		}

	default:
		slog.Warn("Unknown job type", "job_type", job.Type)
	}
}

// Drain workers after shutdown is initiated
func DrainWorkers() {
	close(JobQueue) // unblocks idle workers
	wg.Wait()       // waits for in-flight jobs
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
*/

// AddJob adds a new job to the queue
// This is what handlers call to schedule background work
func AddJob(jobType string, payload interface{}) {
	// Non-blocking send (if queue is full, log warning)
	select {
	case JobQueue <- Job{Type: jobType, Payload: payload}:
		slog.Info("Job added to queue", "job_type", jobType)
	default:
		slog.Warn("Job queue full, dropping job", "job_type", jobType)
	}
}
