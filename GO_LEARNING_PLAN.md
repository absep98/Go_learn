# Go Backend Learning Plan (6–8 Weeks)

Time available: **30–60 minutes per day**  
Goal: Become comfortable with Go for backend work and build a strong portfolio project.

---

## Overview

- **Weeks 1–2** – Core Go refresh (syntax, structs, interfaces, basic concurrency).
- **Weeks 3–4** – Three small mini-projects:
  - Mini-project A: Worker / job processing system.
  - Mini-project B: Simple REST API.
  - Mini-project C: Monitoring / metrics agent.
- **Weeks 5–8** – Flagship backend project that combines A+B+C ideas.

You do **not** have to match calendar weeks exactly. Just move to the next "week" when you finish the tasks.

---

## Week 1 – Core language refresh

### Day 1
- **Goal:** Warm up basic syntax.
- **Tasks:**
  - Write a small program that:
    - Reads a line from stdin.
    - Prints it back with a prefix (e.g., "You said:").
  - Practice:
    - `if`, `for`, `switch`.

### Day 2
- **Goal:** Practice slices and loops.
- **Tasks:**
  - Create a slice of integers and:
    - Append values.
    - Loop with `for` and `range`.
    - Calculate sum and average.

### Day 3
- **Goal:** Structs + basic methods.
- **Tasks:**
  - Define a `User` struct with fields like `FirstName`, `LastName`, `Age`.
  - Write a method `FullName()` that returns a string.
  - Create a few users in `main` and print their full names.

### Day 4
- **Goal:** Work with slices of structs.
- **Tasks:**
  - Create a slice of `User`.
  - Write a function that filters users by minimum age.
  - Print the filtered slice.

### Day 5
- **Goal:** Maps + basic error handling.
- **Tasks:**
  - Build a word counter:
    - Take a string.
    - Split into words.
    - Store counts in `map[string]int`.
    - Print each word and its count.

### Day 6
- **Goal:** Functions, return values, and errors.
- **Tasks:**
  - Write a function that parses an int from string, returns `(int, error)`.
  - Call it with valid and invalid inputs and handle errors with `if err != nil {}`.

### Day 7
- **Goal:** Review and clean up.
- **Tasks:**
  - Pick one small program from this week and refactor it:
    - Better function names.
    - Separate logic into functions.
  - Create a `notes.md` (if you don’t have one) and write:
    - What feels easy now.
    - What still feels confusing.

---

## Week 2 – Interfaces and basic concurrency

### Day 8
- **Goal:** Understand interfaces.
- **Tasks:**
  - Define an interface `Notifier` with method `Notify(msg string)`.
  - Implement `ConsoleNotifier` that prints to stdout.
  - Write a function that takes a `Notifier` interface and calls `Notify`.

### Day 9
- **Goal:** Multiple implementations of interfaces.
- **Tasks:**
  - Add another implementation, e.g. `FileNotifier` (or just simulate writing to file).
  - Show polymorphism: pass different implementations into the same function.

### Day 10
- **Goal:** Intro to goroutines.
- **Tasks:**
  - Write two functions that print messages in a loop.
  - Run them in separate goroutines.
  - Use `time.Sleep` in `main` so you can see outputs (temporary, just for learning).

### Day 11
- **Goal:** Wait for goroutines using `WaitGroup`.
- **Tasks:**
  - Replace `time.Sleep` with `sync.WaitGroup`.
  - Start N goroutines that each print something and then `Done()`.

### Day 12
- **Goal:** Intro to channels.
- **Tasks:**
  - Create a channel of `int`.
  - Start a goroutine that sends numbers 1–5 into the channel.
  - In `main`, receive them and print.

### Day 13
- **Goal:** Combining `WaitGroup` + channels.
- **Tasks:**
  - Start multiple goroutines that compute something (e.g., square of a number) and send result to a channel.
  - In `main`, read from the channel and print results.

### Day 14
- **Goal:** Review concurrency basics.
- **Tasks:**
  - Clean up your examples.
  - Add notes on:
    - When you’d use goroutines.
    - What channels are good for.

---

## Week 3 – Mini-project A: Worker / job processing system

### Goal
Build a simple **worker pool** that processes jobs concurrently.

### Day 15
- Design:
  - Define a `Job` struct (e.g., `ID int`, `Payload string`).
  - Decide how many workers you want (e.g., 3–5).

### Day 16
- Implement a `worker` function that:
  - Takes an ID and a `jobs` channel.
  - Loops over jobs from the channel and "processes" them (e.g., print and `time.Sleep`).

### Day 17
- In `main`:
  - Create the `jobs` channel.
  - Start N workers as goroutines.
  - Send some jobs into the channel.

### Day 18
- Add a `WaitGroup` so that:
  - `main` waits until all jobs are processed.
  - Then closes the `jobs` channel and exits.

### Day 19
- Improve logging:
  - Print when a worker starts/finishes a job.
  - Maybe simulate occasional errors and log them.

### Day 20–21
- Review and polish:
  - Clean up code.
  - Add comments and notes on how the worker pool works.

---

## Week 4 – Mini-project B: Simple REST API

### Goal
Build a small in-memory REST API (e.g., for TODO items or notes).

### Day 22
- Design:
  - Choose entity: `Todo` or `Note` with `ID`, `Title`, `Done`/`Content`.
  - Decide endpoints:
    - `POST /items` – create.
    - `GET /items` – list.

### Day 23
- Implement basic HTTP server using `net/http`.
  - Add a `GET /health` endpoint that returns `{"status":"ok"}`.

### Day 24
- Implement `POST /items`:
  - Read JSON body.
  - Decode into struct.
  - Store in slice or map.

### Day 25
- Implement `GET /items`:
  - Return all items as JSON.

### Day 26
- Add `GET /items/{id}` and/or `DELETE /items/{id}` (using a simple router or manual parsing).

### Day 27–28
- Review:
  - Refactor handlers into separate functions or files.
  - Add basic error handling and logging.

---

## Week 5 – Mini-project C: Monitoring / metrics agent

### Goal
Build a small agent that periodically collects (fake) metrics and exposes them.

### Day 29
- Design:
  - Decide what metrics to track (e.g., `CPUUsage`, `MemoryUsage`).

### Day 30
- Implement a function that generates fake metrics (use random values for now).

### Day 31
- Use `time.Ticker` to collect metrics every few seconds in a goroutine.

### Day 32
- Store recent metrics in memory (slice or struct).

### Day 33
- Expose `GET /metrics` endpoint that returns the latest metrics as JSON.

### Day 34–35
- Add a way to stop the ticker and goroutine cleanly (e.g., using `context` or a `done` channel).

---

## Weeks 6–8 – Flagship backend project

### Idea
**Job Processing API with Monitoring**:
- REST API that accepts jobs.
- Background workers process jobs.
- Monitoring endpoint shows processed jobs and errors.

### High-level milestones

#### Milestone 1 – Basic API + in-memory storage
- `POST /jobs` – create job.
- `GET /jobs` – list jobs.
- Store jobs in memory.

#### Milestone 2 – Integrate worker pool
- Reuse ideas/code from Mini-project A.
- When a job is created, send it to the worker pool.
- Update job status (`pending`, `running`, `done`, `failed`).

#### Milestone 3 – Monitoring endpoint
- Add `/metrics` endpoint that returns:
  - total jobs processed.
  - total failed jobs.
  - maybe per-worker stats.

#### Milestone 4 – Cleanup and documentation
- Refactor into packages (e.g., `api`, `worker`, `store`).
- Write a clear `README` explaining:
  - what the project does.
  - how to run it.
  - which Go concepts it demonstrates.

---

## How to use this file

- Each day, open this file and find your **Day N**.
- Spend 30–60 minutes doing exactly those tasks.
- If a day takes longer, spread it across two real days.
- Update or extend this plan as you progress and learn more.
