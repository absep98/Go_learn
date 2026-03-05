package metrics

import (
	"math"
	"sync"
)

type Metrics struct {
	mu sync.RWMutex

	// Counters (only increases)
	requestCount map[string]int64 // per endpoint
	errorCount   map[string]int64 // per endpoint

	// Gauges (go up and down)
	inFlight map[string]int64 // per endpoint

	// For latency stats (instead of histogram)
	totalLatency map[string]float64 // sum for average
	minLatency   map[string]float64 // fastest
	maxLatency   map[string]float64 // slowest
}

func NewMetrics() *Metrics {
	// Initialize struct with all maps created using make()
	// Return pointer to the struct
	return &Metrics{
		requestCount: make(map[string]int64),
		errorCount:   make(map[string]int64),
		inFlight:     make(map[string]int64),
		totalLatency: make(map[string]float64),
		minLatency:   make(map[string]float64),
		maxLatency:   make(map[string]float64),
	}
}

func (m *Metrics) RequestStarted(path string) {
	m.mu.Lock() // Lock for reading
	m.inFlight[path]++
	defer m.mu.Unlock() // Unlock when function ends
}

func (m *Metrics) RequestCompleted(path string, duration float64, statusCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inFlight[path]--
	m.requestCount[path]++

	if statusCode >= 400 {
		m.errorCount[path]++
	}

	// Update total latency (for calculating average)
	m.totalLatency[path] += duration

	//  Update min latency
	if _, exists := m.minLatency[path]; !exists {
		m.minLatency[path] = math.MaxFloat64 // first time, set to max
	}

	if duration < m.minLatency[path] {
		m.minLatency[path] = duration
	}

	// Update max latency
	if duration > m.maxLatency[path] {
		m.maxLatency[path] = duration
	}
}

func (m *Metrics) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]interface{})

	// Looping through all endpoints that received requests
	for path, count := range m.requestCount {
		// Calculate average latency
		avgLatency := 0.0
		if count > 0 {
			avgLatency = m.totalLatency[path] / float64(count)
		}

		// Build stats for this endpoint
		result[path] = map[string]interface{}{
			"total_requests": count,
			"errors":         m.errorCount[path],
			"in_flight":      m.inFlight[path],
			"avg_latency_ms": avgLatency,
			"min_latency_ms": m.minLatency[path],
			"max_latency_ms": m.maxLatency[path],
		}
	}
	return result
}

/*
=== INTERVIEW ANSWER: METRICS SYSTEM ===

WHAT:
Per-endpoint observability: request count, error count, inflight requests,
and latency stats (avg/min/max). Exposed via /metrics endpoint as a snapshot.

WHY:
Without metrics you're flying blind. You can't know which endpoints are slow,
which are erroring, or how many requests are currently in-flight. Metrics let
you detect problems before users complain.

HOW:
Metrics struct holds maps keyed by endpoint path:
- requestCount: total completed requests per endpoint
- errorCount: responses with status >= 400
- inFlight: requests currently being processed (gauge: goes up and down)
- totalLatency: sum of all durations (used to calculate average)
- minLatency / maxLatency: fastest and slowest request seen

MIDDLEWARE FLOW:
1. RequestStarted(path): Lock() → inFlight[path]++ → Unlock()
2. Record start time: time.Now()
3. Wrap ResponseWriter to intercept WriteHeader() and save status code
   (http.ResponseWriter cannot be read back after writing — wrapper is required)
4. Call next handler with wrapped writer
5. RequestCompleted(path, duration, statusCode):
   Lock() → inFlight-- → requestCount++ → if status>=400: errorCount++
   → update totalLatency, minLatency, maxLatency → Unlock()

WHY RESPONSEWRITER WRAPPER:
Once a handler calls w.WriteHeader(201), that status goes to the client and
cannot be retrieved from the original ResponseWriter. The wrapper struct embeds
http.ResponseWriter, overrides WriteHeader() to capture the code in a field,
then passes through to the real writer. After next() returns, wrapped.statusCode
has the actual status written by the handler.

RWMUTEX — READ vs WRITE LOCK:
- Lock() (write): exclusive. One goroutine at a time. Blocks readers AND writers.
  Used by RequestStarted and RequestCompleted — they modify maps.
- RLock() (read): shared. Multiple goroutines can hold simultaneously.
  Used by GetSnapshot — only reads, never modifies.
  Practical benefit: 10 concurrent /metrics calls don't block each other or
  block incoming requests from updating counters.

RULE: modifying data → Lock(). reading only → RLock().

TRADE-OFFS:
- In-memory only — reset on server restart
- No percentiles (p95/p99) — production uses Prometheus histograms
- Single global AppMetrics var — fine for one server, won't aggregate across fleet
*/
