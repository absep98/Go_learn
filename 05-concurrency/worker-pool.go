package main

import (
	"fmt"
	"sync"
	"time"
)

// Job represents work to be done
type Job struct {
	ID   int
	Task string
}

// Result represents the output of completed work
type Result struct {
	JobID   int
	Output  string
	Worker  int
}

// Worker function - processes jobs from jobs channel
func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for job := range jobs {
		fmt.Printf("🔧 Worker %d started job %d: %s\n", id, job.ID, job.Task)
		
		// Simulate work being done
		time.Sleep(time.Duration(500+job.ID*100) * time.Millisecond)
		
		// Send result
		result := Result{
			JobID:  job.ID,
			Output: fmt.Sprintf("Completed: %s", job.Task),
			Worker: id,
		}
		
		results <- result
		fmt.Printf("✅ Worker %d finished job %d\n", id, job.ID)
	}
}

func main() {
	fmt.Println("=== Worker Pool Pattern ===")
	
	// Create channels
	jobs := make(chan Job, 10)     // Jobs to be done
	results := make(chan Result, 10) // Results from workers
	
	// Number of workers
	numWorkers := 3
	
	// WaitGroup to wait for all workers to finish
	var wg sync.WaitGroup
	
	// Start workers
	fmt.Printf("🚀 Starting %d workers...\n", numWorkers)
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}
	
	// Send jobs
	jobList := []Job{
		{1, "Process data file A"},
		{2, "Generate report B"},
		{3, "Send email notifications"},
		{4, "Backup database"},
		{5, "Clean temporary files"},
		{6, "Update user profiles"},
		{7, "Generate thumbnails"},
	}
	
	fmt.Printf("📨 Sending %d jobs to workers...\n", len(jobList))
	go func() {
		for _, job := range jobList {
			jobs <- job
		}
		close(jobs) // No more jobs
	}()
	
	// Collect results
	go func() {
		wg.Wait()      // Wait for all workers to finish
		close(results) // Close results channel
	}()
	
	// Print results as they come in
	fmt.Println("\n📊 Results:")
	for result := range results {
		fmt.Printf("📝 Job %d (Worker %d): %s\n", 
			result.JobID, result.Worker, result.Output)
	}
	
	fmt.Println("\n🎉 All jobs completed!")
	fmt.Println("Notice how jobs were distributed among workers automatically!")
}