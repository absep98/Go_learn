# pacer/ package — constant-rate load pacer

Real code: `xk6-wss/internal/pkg/phobos-lib/pacer/`
Only 100 lines total — the smallest package in the repo, but ubiquitous:
every load-test loop in wss.go calls `p.Pace(elapsed)` to time its next request.

## The interface (pacer.go, 21 lines)
```go
type Pacer interface {
    Pace(elapsed time.Duration) (time.Duration, time.Duration, bool)
}
```
Called BEFORE each request. Returns 3 things:
1. **delay** — how long to sleep before firing the next request
2. **expectedDuration** — when this request was "supposed" to fire (for metrics)
3. **stop bool** — "we're done, exit the loop"

Typical caller pattern:
```go
delay, _, stop := p.Pace(elapsed)
if stop { return }
time.Sleep(delay)
// fire request
```

## The factory (same file)
```go
func FromConfiguration(cfg map[string]interface{}) (Pacer, error) {
    return constantFromConfiguration(cfg)
}
```
Same Factory pattern as auth/configuration.go — one entry point returns
whichever Pacer implementation the config asks for. Only "constant" today;
shape is ready for more types without touching callers.

## The constant implementation (constant.go, 79 lines)

**Idea in plain English:** "fire `count` requests spread evenly across `in`."
Example: count=1000, in=10s → 100 req/sec, spaced 10ms apart, for 10 seconds.

**Struct = plan + progress:**
- `duration` (total window), `target` (total req count), `completed` (fired so far)
- `rate` (req/sec = target/duration), `interval` (nanoseconds between = 1s/rate)

## Two clever design choices worth learning from

### 1. `defer cp.completed++` placed mid-function (not at top)
There are 5 `return` statements. The `defer` sits between them:
- Returns BEFORE the defer line (time's up, rate=0, all done): counter NOT bumped
- Returns AFTER the defer line (pacing an actual request): counter IS bumped

The **placement of `defer` is the logic**. "Only count when we're actually pacing
a real request; don't count sentinel/done returns." Elegant use of defer's
"only runs if we reach the defer statement" rule.

### 2. Sleep=0 branch = self-correcting when behind schedule
```go
if cp.completed < expectedCount {
    return 0, expectedDuration, false   // sleep 0 → fire immediately
}
return expectedDuration - elapsedDuration, expectedDuration, false  // sleep the gap
```
- `expectedCount = rate × elapsed` — how many SHOULD have fired by now
- If actual `completed` < `expectedCount`, we're **behind schedule** (maybe a
  slow request stalled us). Return delay=0 → fire immediately → burn through the
  backlog → catch up.
- If we're on/ahead of schedule, sleep the gap to slow down.
- Self-correcting: the pacer keeps you aligned with the target rate even when
  previous requests took longer than expected.

## Concept recap
- Interface + factory pattern (again — 3rd time in this codebase after auth/, err/, pacer/)
- `defer` position as logic, not just cleanup
- Self-correcting rate control via "expected vs actual" math
