---
title: I Needed to Send an HTTP Request Without Slowing Down My API. Goroutines Fixed That.
subtitle: Part 13 of Learning Go in Public
tags: go, goroutines, backend, buildinpublic
series: Learning Go in Public
---

In [Part 12](https://dev.to/abhishek_sharma_a9792aee8/my-backend-crashed-every-time-redis-went-down-three-patterns-fixed-that-43j2), I added resilience patterns — retry, timeout, circuit breaker. My backend could now survive Redis going down.

But there was still a problem. When someone created an entry, I needed to notify an external service. A webhook: fire an HTTP POST, tell the world something happened.

I added it inline in the handler:

```go
func CreateEntry(w http.ResponseWriter, r *http.Request) {
    // ... validate, save to DB ...
    webhook.Send(url, payload) // 🐌 blocks here
    respondJSON(w, http.StatusCreated, ...)
}
```

The user waited. Not for their entry to save — that was fast. They waited for *my* HTTP request to some external service to complete. If that service was slow, every create request was slow. If it was down, every create request failed.

The handler shouldn't care whether the webhook delivery succeeded. The entry was saved. That's the job. Everything else is secondary.

I needed fire-and-forget.

## The Wrong Mental Model

My first instinct: just run it in a goroutine.

```go
go webhook.Send(url, payload) // "fire and forget", right?
```

Technically works. But now I have an unbounded number of goroutines. 10,000 entries created? 10,000 goroutines, each holding an HTTP connection, all fighting over bandwidth. Goroutines are cheap but not free — and more importantly, there's no control. No backpressure. No way to say "slow down."

The correct solution is a **worker pool with a channel**.

## The Pizza Shop Mental Model

Imagine a pizza shop:

- **Channel** = the order counter. Tickets pile up here.
- **Workers** = the chefs. Fixed number. Each processes one order at a time.
- **Handler** = the cashier. Takes the order, puts it on the counter, immediately says "order received." Doesn't wait for the pizza.

The cashier's job is done in milliseconds. The chefs handle the rest in the background. If the shop gets slammed, orders queue up — but there's a limit. If the queue is full, you say "we're busy, try later" instead of spinning up infinite chefs.

## The Worker Pool

```go
type Job struct {
    Type    string
    Payload interface{}
}

var JobQueue chan Job

func StartWorkerPool(numWorkers int) {
    JobQueue = make(chan Job, 100) // buffered: 100 jobs can queue up

    for i := 1; i <= numWorkers; i++ {
        go worker(i)
    }

    slog.Info("Worker pool started", "workers", numWorkers)
}

func worker(id int) {
    for job := range JobQueue {
        processJob(job)
    }
}
```

Three things to understand:

**Buffered channel (`make(chan Job, 100)`)**: The queue can hold 100 jobs without blocking. When the handler sends a job, it returns immediately if there's space. If the queue is full, the `select` in `AddJob` drops it with a warning rather than blocking the handler.

**`for job := range JobQueue`**: This blocks until a job arrives, processes it, then loops back and blocks again. No busy-waiting. No polling. The goroutine sleeps until there's work. When `JobQueue` is closed, the loop exits cleanly — that's how graceful shutdown works.

**Fixed goroutine count**: Three workers (configurable). Not one goroutine per request. The pool handles any volume without spinning up new goroutines.

```go
func AddJob(jobType string, payload interface{}) {
    select {
    case JobQueue <- Job{Type: jobType, Payload: payload}:
        // Added successfully
    default:
        slog.Warn("Job queue full, dropping job", "type", jobType)
    }
}
```

Non-blocking send with `select`. If the queue is full, log a warning and move on. The API response is never delayed.

## The Handler Change

Before:
```go
// Handler blocks waiting for webhook
webhook.Send(url, payload)
respondJSON(w, http.StatusCreated, ...)
```

After:
```go
// Handler returns immediately, worker delivers webhook in background
worker.AddJob("entry_created", map[string]interface{}{
    "event":     "entry_created",
    "entry_id":  id,
    "user_id":   userID,
    "text":      req.Text,
    "mood":      req.Mood,
    "category":  req.Category,
    "timestamp": time.Now().UTC(),
})
respondJSON(w, http.StatusCreated, ...)
```

`AddJob` returns in nanoseconds. The user gets their 201 immediately. The worker picks up the job and delivers the webhook in the background.

## The Webhook Delivery

The webhook sender itself is simple:

```go
func Send(url string, payload interface{}) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshal failed: %w", err)
    }

    resp, err := http.Post(url, "application/json", bytes.NewReader(data))
    if err != nil {
        return fmt.Errorf("http post failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("webhook returned non-2xx: %d", resp.StatusCode)
    }
    return nil
}
```

But a plain HTTP call in the background isn't resilient. The external service might be down, or slow, or return a 500. So the worker wraps delivery with retry + circuit breaker — the patterns from Post 12:

```go
func processJob(job Job) {
    switch job.Type {
    case "entry_created", "entry_deleted":
        err := WebhookBreaker.Execute(func() error {
            return retry.Do(3, 500*time.Millisecond, func() error {
                return sender.Send(webhookURL, job.Payload)
            })
        })
        if err != nil {
            slog.Error("Webhook delivery failed", "error", err)
        }
    }
}
```

Three layers:
1. **Circuit breaker**: if the webhook service has 5 consecutive failures, stop trying for 30 seconds. Don't waste goroutine time on a dead service.
2. **Retry with backoff**: attempt delivery up to 3 times, waiting 500ms → 1s → 2s between attempts.
3. **Log and continue**: if all attempts fail, log it and move on. The job failing doesn't crash the worker or affect the API.

Note `job.Payload` — not `job`. The external service gets clean data:

```json
{
  "event": "entry_created",
  "entry_id": 42,
  "user_id": 3,
  "text": "productive day",
  "mood": 8,
  "category": "work",
  "timestamp": "2026-04-17T13:37:00Z"
}
```

Not internal Go struct fields with type names and Go-specific formatting. Just the data the receiver actually needs.

## The Tricky Part: Graceful Shutdown

Worker pools introduce a subtle problem. When the server gets a shutdown signal (Ctrl+C, SIGTERM), what happens to jobs that are queued but not yet processed?

Without graceful shutdown: server exits, 50 queued jobs dropped, 50 webhooks never sent.

The fix:

```go
var wg sync.WaitGroup

func worker(id int) {
    for job := range JobQueue {
        wg.Add(1)
        processJob(job)
        wg.Done()
    }
}

func StopWorkerPool() {
    close(JobQueue) // signals workers to stop accepting new jobs
    wg.Wait()       // blocks until all in-progress jobs complete
}
```

In `main.go`, the graceful shutdown sequence:

```go
// On SIGTERM/SIGINT:
server.Shutdown(ctx) // stop accepting new HTTP requests
worker.StopWorkerPool() // drain the job queue
```

HTTP stops first — no new jobs get created. Then the worker pool drains — all queued jobs run to completion. Then the process exits. No jobs dropped.

## Putting It Together

The complete flow for `POST /entries`:

```
1. Handler receives request (milliseconds)
2. Validates input, saves to Postgres
3. Calls worker.AddJob() — returns immediately
4. Handler sends 201 to client — user is done waiting

[background, asynchronously]
5. Worker picks up job from channel
6. Circuit breaker: is webhook service healthy?
7. retry.Do: attempt 1 → fail → wait 500ms → attempt 2 → success
8. HTTP POST to webhook.site with full entry payload
9. Worker loops back and waits for next job
```

The user never experiences steps 5-9. They got their response at step 4.

## What I Learned

**Channels are not just a communication mechanism — they're a coordination tool.** A buffered channel is a queue with backpressure built in. When it fills up, you know you need more workers or fewer producers. Goroutines alone give you none of that.

**`for job := range channel` is the cleanest worker pattern in Go.** It blocks, processes, loops — and exits cleanly when the channel is closed. No `select`, no `done` channel, no `context`. Just range over the channel.

**Fire-and-forget is not the same as "spawn a goroutine."** Fire-and-forget means the caller doesn't wait for the result. A worker pool is how you implement fire-and-forget responsibly — with a fixed number of workers, a bounded queue, and graceful shutdown.

**Close the channel to stop workers, use WaitGroup to wait for them.** This two-step pattern — `close(JobQueue)` then `wg.Wait()` — is the canonical way to drain a worker pool in Go. Close signals "no more work." WaitGroup confirms "current work is done."

---

Up next: the last piece of the puzzle — dependency injection refactor, per-user rate limiting, and token refresh. Everything that ties the backend together.

*This is Part 13 of "Learning Go in Public". [Part 1](https://dev.to/abhishek_sharma_a9792aee8/why-i-decided-to-learn-go-and-what-my-first-commit-looked-like-38ii) | [Part 2](https://dev.to/abhishek_sharma_a9792aee8/i-stopped-watching-tutorials-and-started-building-a-rest-api-in-go-4e84) | [Part 3](https://dev.to/abhishek_sharma_a9792aee8/building-authentication-from-scratch-in-go-no-libraries-no-magic-2c46) | [Part 4](https://dev.to/abhishek_sharma_a9792aee8/tests-update-delete-and-the-refactor-i-didnt-plan-5g2o) | [Part 5](https://dev.to/abhishek_sharma_a9792aee8/what-happens-when-your-api-has-10000-rows-i-added-pagination-and-caching-to-find-out-2jkk) | [Part 6](https://dev.to/abhishek_sharma_a9792aee8/i-built-rate-limiting-from-scratch-in-go-then-replaced-it-with-redis-82i) | [Part 7](https://dev.to/abhishek_sharma_a9792aee8/my-server-was-silently-dropping-requests-heres-the-one-line-fix-that-changed-everything-5b4g) | [Part 8](https://dev.to/abhishek_sharma_a9792aee8/i-built-my-backend-on-sqlite-then-i-deleted-it-15a) | [Part 9](https://dev.to/abhishek_sharma_a9792aee8/my-database-disappeared-every-time-i-restarted-docker-migrations-fixed-that-n93) | [Part 10](https://dev.to/abhishek_sharma_a9792aee8/every-request-looked-the-same-in-my-logs-then-i-added-one-line-19m7) | [Part 11](https://dev.to/abhishek_sharma_a9792aee8/i-had-no-idea-which-endpoint-was-slowest-so-i-built-my-own-metrics-5fjp) | [Part 12](https://dev.to/abhishek_sharma_a9792aee8/my-backend-crashed-every-time-redis-went-down-three-patterns-fixed-that-43j2)*