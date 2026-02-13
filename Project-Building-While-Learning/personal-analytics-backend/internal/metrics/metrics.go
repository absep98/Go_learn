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
