Great questions! Let me explain real-world approaches.

## Real-World Issue 1: Thread Safety ✅

**Yes, this is a REAL production issue!**

**What happens without protection:**
- Race conditions cause data corruption
- Server can crash with "concurrent map write"
- Metrics become unreliable/random

**How companies solve it:**

| Approach | Used By | Pros | Cons |
|----------|---------|------|------|
| `sync.Mutex` | Simple apps | Easy to understand | Blocks all access |
| `sync.RWMutex` | Production apps | Multiple readers OK | Slightly complex |
| `sync/atomic` | High-performance | Fastest for counters | Only for simple types |
| Prometheus library | Most companies | Handles everything | External dependency |

**Our approach:** `sync.RWMutex` - production-grade, no external dependencies

---

## Real-World Issue 2: Memory Explosion ✅

**Yes, this is a REAL production issue!**

**What happens storing all latencies:**
- Server runs out of memory after millions of requests
- GC (garbage collection) pauses increase
- Server becomes slow, then crashes

**How companies solve it:**

| Approach | What It Does | Trade-off |
|----------|--------------|-----------|
| Running sum/count | Keep total, calculate avg | Loses percentiles (p99) |
| Histogram buckets | Count requests in ranges (0-10ms, 10-50ms, etc.) | Fixed bucket boundaries |
| HDR Histogram | Advanced math for percentiles | Complex |
| Prometheus Histogram | Industry standard | External library |

**Our approach:** Running sum + min/max - simple, no memory issues, good enough for learning

---

## Real-World Structure: Prometheus-Style Metrics

Companies use **metric types** defined by Prometheus:

```
Counter   → Only increases (request count, error count)
Gauge     → Can go up/down (in-flight requests, memory usage)
Histogram → Distribution (latency percentiles)
```

**For us, simplified version:**

```go
type Metrics struct {
    mu sync.RWMutex

    // Counters (only increase)
    requestCount map[string]int64   // per endpoint
    errorCount   map[string]int64   // per endpoint

    // Gauges (go up and down)
    inFlight     map[string]int64   // per endpoint

    // For latency stats (instead of histogram)
    totalLatency map[string]float64 // sum for average
    minLatency   map[string]float64 // fastest
    maxLatency   map[string]float64 // slowest
}
```

---

## Your Task: Create the Structure

Now create `internal/metrics/metrics.go` with:

**Part 1: The struct** (as shown above)

**Part 2: Constructor function**
```go
func NewMetrics() *Metrics {
    // Initialize struct with all maps created using make()
    // Return pointer to the struct
}
```

**Remember:**
- Maps need `make()` to initialize
- Return pointer (`*Metrics`) so all code shares same instance

Go ahead and write the code! Start with just the struct and constructor.
