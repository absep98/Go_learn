package worker

import (
	"log"
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

	log.Printf("🔧 Started %d background workers", numWorkers)
}

// worker is a single worker that processes jobs from the queue
// It runs forever, waiting for jobs
func worker(id int) {
	// This loop runs FOREVER (until program exits)
	for job := range JobQueue {
		// "range JobQueue" = wait for next job, then process it
		// When JobQueue is closed, this loop exits

		log.Printf("👷 Worker %d: Processing job type=%s", id, job.Type)

		// Process the job based on its type
		processJob(job)

		log.Printf("✅ Worker %d: Completed job type=%s", id, job.Type)
	}
}

// processJob handles different job types
func processJob(job Job) {
	switch job.Type {
	case "entry_created":
		// Simulate sending notification / updating analytics
		// In real app: send email, update stats, notify webhooks, etc.
		log.Printf("   📝 Processing entry creation: %v", job.Payload)
		time.Sleep(2 * time.Second) // Simulate slow task

	case "entry_deleted":
		log.Printf("   🗑️ Processing entry deletion: %v", job.Payload)
		time.Sleep(1 * time.Second)

	default:
		log.Printf("   ❓ Unknown job type: %s", job.Type)
	}
}

// ========================================
// HELPER TO ADD JOBS
// ========================================

// AddJob adds a new job to the queue
// This is what handlers call to schedule background work
func AddJob(jobType string, payload interface{}) {
	// Non-blocking send (if queue is full, log warning)
	select {
	case JobQueue <- Job{Type: jobType, Payload: payload}:
		log.Printf("📬 Job added to queue: %s", jobType)
	default:
		log.Printf("⚠️ Job queue full! Dropping job: %s", jobType)
	}
}
