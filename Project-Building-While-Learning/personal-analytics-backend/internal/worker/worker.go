package worker

import (
	"log/slog"
	"personal-analytics-backend/internal/circuitbreaker"
	"personal-analytics-backend/internal/retry"
	"personal-analytics-backend/internal/webhook"
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

var WebhookBreaker = circuitbreaker.NewCircuitBreaker(5, 3*time.Second)

// StartWorkerPool starts N workers that listen for jobs
// Each worker is like a "chef" waiting for orders
func StartWorkerPool(numWorkers int) {
	// Create the job queue (buffered channel with capacity 100)
	// Buffer = how many jobs can wait in line before blocking
	JobQueue = make(chan Job, 100)

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
				return webhook.Send("https://webhook.site/11eccba7-a84d-4b58-b86e-68d56a5d7021", job)
			})
		})

		if err != nil {
			slog.Error(err.Error())
		}

	case "entry_deleted":
		slog.Debug("Processing entry deletion", "payload", job.Payload)
		time.Sleep(1 * time.Second)

	default:
		slog.Warn("Unknown job type", "job_type", job.Type)
	}
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

TRADE-OFFS:
- Jobs dropped when queue full — acceptable for non-critical background work
- No retry on failed jobs — production would use a persistent queue (Redis, RabbitMQ)
- No graceful drain on shutdown — production would drain queue before exiting
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
