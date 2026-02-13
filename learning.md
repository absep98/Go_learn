Based on my comprehensive analysis of the codebase, I'll now provide you with a detailed breakdown of this repository.

---

# Comprehensive Codebase Analysis: wss-tools

## HIGH-LEVEL OVERVIEW

### 1. What Problem This Service Solves

**wss-tools** is a **performance and load testing toolkit** for Citrix's **Secure Private Access (SPA)** solution. It enables testing of VPN tunneling, authentication, and secure application access at scale.

**Key problems it solves:**

- **Scale Validation**: Tests whether NetScaler Gateway can handle 5,000, 10,000, or 20,000 concurrent users
- **End-to-End Performance Testing**: Measures latency and throughput for TCP/UDP tunnels, HTTP access, and VPN connections
- **Sizing Guidance**: Helps establish infrastructure requirements for enterprise deployments
- **Automation**: Provides E2E test automation for SPA configuration, policies, and client behavior

### 2. System Type: **Mixed System**

This is a **multi-tool repository** that includes:

| Component | Type | Technology |
|-----------|------|------------|
| `xk6-wss` | **Custom k6 Extension (Go)** | Go module for K6 load testing |
| `system-testing` | **Load Test Runner** | TypeScript + K6 |
| `automation` | **E2E Test Framework** | Python + Playwright + Pytest |
| `setup` | **Infrastructure Setup** | TypeScript + K6 for bulk provisioning |
| `utils` | **Utility Servers** | Go echo servers (TCP/UDP) |

### 3. Request/Job Flow (End-to-End)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  K6 Test Script (TypeScript) → Compiled to JS → K6 Runtime                 │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  xk6-wss Go Extension                                                        │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐         │
│  │   Login()       │───▶│   NewSession()  │───▶│  AccessTcpApp() │         │
│  │   (auth/athena) │    │   (VPN Session) │    │  AccessUdpApp() │         │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘         │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  TLS Tunnel to NetScaler Gateway (port 443)                                  │
│  - CsTCP: Client-Server TCP tunneling                                       │
│  - CsUDP: Multiplexed UDP over TLS                                          │
│  - SecureBrowse: HTTP proxy tunneling                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Backend Echo Servers (TCP/UDP) │  InfluxDB (Metrics) │  Redis (Session)   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## ARCHITECTURE

### Entry Point (Main Packages)

**There is no traditional Go `main()` entry point** for the core system. This is because:

1. **xk6-wss** is a **K6 extension module** that gets compiled into K6
2. The "main" is effectively **K6 itself** loading this extension

However, there are utility entry points:

| File | Purpose |
|------|---------|
| `utils/echo-server.go` | Simple TCP echo server for testing |
| `utils/udp-echo-server.go` | UDP echo server for testing |
| `xk6-wss/wss.go` (init function) | K6 module registration |

```go name=xk6-wss/wss.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/wss.go#L35-L40
// init is called by the Go runtime at application startup.
func init() {
    modules.Register("k6/x/wss", New())
}
```

### How Routing/Protocol Handling Works

**This is NOT an HTTP API service** - it's a VPN client library. Instead of HTTP routing, there's **protocol handling**:

```go name=xk6-wss/internal/pkg/ngs-vpn-lib/vpn/vpn.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/internal/pkg/ngs-vpn-lib/vpn/vpn.go#L426-L452
func (s *Session) TunnelHTTP(req *http.Request, tunnelType int) (*http.Response, error) {
    switch tunnelType {
    case TunnelTypeCSTCP:
        return s.HTTP(req)
    case TunnelTypeSecureBrowse:
        return s.SecureBrowse(req)
    default:
        panic(fmt.Errorf("Invalid tunnel type: %d", tunnelType))
    }
}
```

**Tunnel Types:**

- `TunnelTypeCSTCP (0)`: Direct TCP through VPN
- `TunnelTypeSecureBrowse (1)`: HTTP proxy-style tunneling

### Configuration Loading

Configuration is loaded via:

1. **Environment variables** for deployment settings
2. **K6 test script configuration** (TypeScript → JSON)
3. **Runtime config maps** passed to Go extension

```typescript name=system-testing/src/load-profiles/benchmark-10k-profile.ts url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/system-testing/src/load-profiles/benchmark-10k-profile.ts#L11-L24
export const benchmark_options = {
    insecureSkipTLSVerify: true,
    scenarios: {
        login_scenario: {
            executor: 'constant-arrival-rate',
            rate: 500,
            timeUnit: '1m',
            duration: '10m',
            preAllocatedVUs: 10,
            gracefulStop: '1m',
            maxVUs: 15,
            exec: 'executeLogin',
        },
        // ...
    }
};
```

### Dependency Wiring

| Dependency | How it's Wired | Purpose |
|------------|----------------|---------|
| Redis | Direct client creation | Session state storage |
| InfluxDB | K6 output plugin | Metrics persistence |
| Grafana | External service | Visualization |
| NetScaler Gateway | TLS dial | VPN tunnel endpoint |

```go name=xk6-wss/wss.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/wss.go#L605-L616
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

key := strconv.Itoa(index)
authCookieStr, err := redisClient.Get(ctx, key).Result()
```

**Design Decision**: Redis is hardcoded to `localhost:6379`. This is intentional since each K6 load generator VM has its own local Redis.

---

## CORE COMPONENTS

### 1. VPN Session Management (vpn/vpn.go)

The `Session` struct is the central abstraction:

```go name=xk6-wss/internal/pkg/ngs-vpn-lib/vpn/vpn.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/internal/pkg/ngs-vpn-lib/vpn/vpn.go#L29-L40
type Session struct {
    aaacCookie    *http.Cookie
    vpnFQDN       string
    udpMux        *tls.Conn      // Shared UDP multiplexer
    muxMap        *util.SafeMap  // Thread-safe map for UDP responses
    rand          *rand.Rand
    srcIP         net.IP
    dnsIP         net.IP
    transport     *http.Transport
    timeout       time.Duration
    CfgResponseMs int32
}
```

**Key Methods:**

- `NewSession()`: Establishes VPN session with authentication cookie
- `CsTCP()`: Creates TCP tunnel through VPN
- `CsUDP()`: Establishes multiplexed UDP channel
- `MuxUDP()`: Sends/receives UDP through multiplexer

### 2. Authentication (auth/athena.go)

Handles Citrix-specific authentication flow:

```go name=xk6-wss/internal/pkg/ngs-vpn-lib/auth/athena.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/internal/pkg/ngs-vpn-lib/auth/athena.go#L390-L410
func (a *Athena) domainProtocols(protocolsEndpoint, authDomain, tokenEndpoint string) (explicitFormsStartEndpoint string, err error) {
    token := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
    <requesttoken xmlns="http://citrix.com/delivery-services/1-0/auth/requesttoken">
        <for-service>1.-i.%s.token</for-service>
        <for-service-url>%s</for-service-url>
        // ...
    </requesttoken>`, authDomain, tokenEndpoint)
    // ...
}
```

### 3. K6 Extension API (wss.go)

Exports JavaScript-callable functions:

```go name=xk6-wss/wss.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/wss.go#L258-L290
// Login function to invoke DoClassicAuth and NewSession
func (n *NSVpn) Login(config map[string]interface{}, userIndex int) error {
    // ...
}

func (n *NSVpn) AccessUdpApp(config map[string]interface{}, index int) error {
    // ...
}

// AccessTcpApp function to access a TCP application
func (n *NSVpn) AccessTcpApp(config map[string]interface{}, index int) error {
    // ...
}

func (n *NSVpn) Logout(fqdn string, userIndex int) error {
    // ...
}
```

### 4. Thread-Safe UDP Multiplexing (SafeMap)

```go name=xk6-wss/internal/pkg/ngs-vpn-lib/util/safemap.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/internal/pkg/ngs-vpn-lib/util/safemap.go#L18-L26
// SafeMap is a thread-safe map for holding expirable values
type SafeMap struct {
    data           map[string]interface{}
    lock           *sync.RWMutex
    timeout        time.Duration
    cleanupFn      func(interface{})
    refreshTimeout chan *cleanupAction
}
```

**Why This Exists**: UDP responses come asynchronously on a shared TLS connection. The `SafeMap` correlates request IDs to response channels.

### 5. Metrics Collection

```go name=xk6-wss/wss.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/wss.go#L236-L258
func (n *NSVpn) pushMetrics(metricsValues map[*metrics.Metric]float64) {
    // Random nanosecond offset to avoid InfluxDB timestamp collisions
    now := time.Now().Add(time.Duration(rand.Intn(1000)) * time.Nanosecond)
    ctx := n.vu.Context()
    tags := n.vu.State().Tags.GetCurrentValues().Tags

    var samples []metrics.Sample
    for metric, value := range metricsValues {
        samples = append(samples, metrics.Sample{
            TimeSeries: metrics.TimeSeries{
                Metric: metric,
                Tags:   tags,
            },
            Value: value,
            Time:  now,
        })
    }

    metrics.PushIfNotDone(ctx, n.vu.State().Samples, metrics.ConnectedSamples{
        Samples: samples,
    })
}
```

**Important Design Note**: The random nanosecond offset is a workaround for InfluxDB dropping duplicate timestamps from concurrent VUs.

---

## CONCURRENCY & PERFORMANCE

### Where Goroutines Are Used

| Location | Purpose | Pattern |
|----------|---------|---------|
| `echo-server.go` | Handle client connections | `go handleConnection(conn)` |
| `SafeMap.CleanupLoop()` | Background key expiration | `go sm.CleanupLoop()` |
| `vpn.go:udpMuxLoop()` | UDP response demultiplexing | Background reader goroutine |
| `SendDNSRequest()` | Async DNS queries | Rate-limited loop with context |

```go name=utils/echo-server.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/utils/echo-server.go#L21-L32
for {
    conn, err := listener.Accept()
    if err != nil {
        fmt.Printf("Failed to accept connection: %v\n", err)
        continue
    }
    go handleConnection(conn)  // Classic goroutine-per-connection
}
```

### How Shared State Is Protected

```go name=xk6-wss/internal/pkg/ngs-vpn-lib/util/safemap.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/internal/pkg/ngs-vpn-lib/util/safemap.go#L107-L117
func (m *SafeMap) Get(key string) (interface{}, bool) {
    m.lock.RLock()                    // Read lock for concurrent reads
    defer m.lock.RUnlock()

    value, ok := m.data[key]
    if ok {
        m.refreshTimeout <- &cleanupAction{action: CleanupActionRefresh, key: key}
    }
    return value, ok
}
```

**Pattern**: RWMutex for read-heavy workloads, channel for async timeout refresh.

### Context Usage

```go name=xk6-wss/wss.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/wss.go#L383-L410
func (n *NSVpn) SendDNSRequest(ctx context.Context, pacer pacer.Pacer, domains []string, s *vpn.Session) {
    for {
        select {
        case <-ctx.Done():  // Graceful cancellation
            return
        default:
        }
        // ... send DNS request
    }
}
```

**Context is used correctly** for cancellation propagation in long-running operations.

### Shutdown Handling

**⚠️ RISK**: There's no explicit graceful shutdown for:

- TCP/UDP tunnels
- Redis connections
- UDP multiplexer cleanup

The system relies on K6's graceful stop periods and OS cleanup.

---

## RELIABILITY & PRODUCTION CONCERNS

### Error Handling Strategy

The codebase uses **custom error types** with error codes:

```go
err = lerr.Errorf(lerr.CsTcpErr, "Unexpected Status Code %d", resp.StatusCode)
```

**Pattern**: Errors are wrapped with context but don't use Go 1.13+ error wrapping (`%w`).

### Logging

**Go side**: Uses `fmt.Printf` - plain text logging (no structured logging).

**Python side**: Uses custom Logger class with file + console output:

```python name=automation/utils/logger.py url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/automation/utils/logger.py#L19-L43
formatter = logging.Formatter(
    '%(asctime)s.%(name)s.%(levelname)s.%(filename)s: %(message)s'
)
```

**⚠️ RISK**: No structured logging (JSON) makes log aggregation/parsing difficult in production.

### Metrics/Monitoring

**Excellent metrics coverage** with custom K6 metrics:

```go name=xk6-wss/wss.go url=https://github.com/csg-citrix-secure-private-access/wss-tools/blob/8af54b5bbae4539bd95719d27676aeee96afcc3f/xk6-wss/wss.go#L94-L170
m.CfgReqDuration, err = registry.NewMetric("cfg_req_duration", metrics.Trend, metrics.Time)
m.TcpRtt, err = registry.NewMetric("tcp_rtt_milisec", metrics.Trend, metrics.Time)
m.MuxudpBytesSent, err = registry.NewMetric("muxudp_bytes_sent", metrics.Counter, metrics.Default)
m.MuxudpRxTimeouts, err = registry.NewMetric("muxudp_rx_timeouts", metrics.Counter, metrics.Default)
```

**Metrics are batched** to reduce InfluxDB load:

```go
// Push metrics every 100 iterations for reduced metric data
if iterationCount >= 100 {
    n.pushMetrics(batchMetrics)
}
```

### Health Checks

**Not applicable** - this is a client-side load testing tool, not a server.

### Rate Limiting / Retries

Built-in through K6 executor configurations:

```typescript
executor: 'constant-arrival-rate',
rate: 500,           // 500 requests per minute
timeUnit: '1m',
```

Plus custom pacer implementation for DNS:

```go
delay, _, stop := pacer.Pace(time.Since(startTime))
if stop { return }
time.Sleep(delay)
```

---

## SYSTEM DESIGN TRADEOFFS

### Why This Architecture Was Chosen

1. **K6 Extension Pattern**: Enables high-performance load testing with Go's efficiency while keeping test scripts in TypeScript
2. **Redis for Session State**: Allows distributed test execution across multiple VMs
3. **UDP Multiplexing**: Single TLS connection handles all UDP packets, reducing connection overhead
4. **Goroutine-per-connection for Echo Server**: Simple, sufficient for test servers

### What Would Break at Scale

| Issue | Threshold | Symptom |
|-------|-----------|---------|
| Redis single-instance | ~50K concurrent sessions | Latency spikes, memory exhaustion |
| Hardcoded `localhost:6379` | Multiple machines | No shared state |
| `fmt.Printf` logging | High VU counts | Performance degradation, no log aggregation |
| Single UDP mux connection | Network saturation | Packet loss, increased RTT |
| No connection pooling | TCP tunnel creation rate | Port exhaustion |

### Improvements for 10x Traffic

1. **Redis Cluster or Sentinel**: Distribute session state
2. **Connection pooling**: Reuse TLS connections for TCP tunnels
3. **Structured logging with sampling**: JSON logs, sample high-frequency events
4. **Multiple UDP mux connections**: Shard by user ID
5. **gRPC instead of Redis**: Lower latency for hot paths
6. **Resource limits**: Add explicit limits on goroutines, connections

---

## 5 INTERVIEW QUESTIONS & MODEL ANSWERS

### Q1: "How does the UDP multiplexing work in this codebase, and why was it designed this way?"

**Model Answer**:
> "The system maintains a single TLS connection per VPN session for all UDP traffic (`udpMux` in the Session struct). Each UDP packet is wrapped with a unique source port that acts as a correlation ID. The `SafeMap` data structure maps these IDs to response channels.
>
> When a packet is sent, a channel is created and stored. A background goroutine (`udpMuxLoop`) reads responses and routes them to the correct channel based on the correlation ID. This design was chosen because:
>
> 1. It reduces TLS handshake overhead (one connection vs many)
> 2. It's how actual VPN clients work - they multiplex over a single tunnel
> 3. Memory is bounded by the SafeMap's expiration mechanism"

### Q2: "What's the thread-safety strategy in SafeMap, and could you improve it?"

**Model Answer**:
> "SafeMap uses an `RWMutex` for data access - read locks for concurrent reads, write locks for modifications. It also uses a channel to communicate timeout refreshes to a background cleanup goroutine.
>
> Improvements I'd consider:
>
> 1. Use `sync.Map` for better performance with high-contention workloads
> 2. Shard the map to reduce lock contention (e.g., 16 shards, hash by key)
> 3. Use a time.Timer per key instead of periodic cleanup scanning
>
> However, the current design is appropriate for the workload - thousands of keys with moderate access patterns."

### Q3: "How would you debug a situation where metrics show high UDP timeouts?"

**Model Answer**:
> "I'd approach this systematically:
>
> 1. **Check muxudp_rx_timeouts vs muxudp_bytes_sent** - Is it packet loss or slow responses?
> 2. **Verify the UDP mux connection** - Is `udpMux` nil or closed?
> 3. **Check SafeMap expiration** - The 60-second timeout might be too aggressive
> 4. **Network layer** - Use `tcpdump` to verify packets reach the gateway
> 5. **Gateway-side logs** - Check if requests are being processed
>
> The metrics batching (every 20 iterations) might mask intermittent issues, so I'd temporarily reduce that threshold for debugging."

### Q4: "Why does pushMetrics add a random nanosecond offset to timestamps?"

**Model Answer**:
> "This is a pragmatic workaround for InfluxDB's behavior. When multiple VUs (virtual users) push metrics at the same instant, they can have identical timestamps. InfluxDB treats points with the same measurement, tags, and timestamp as duplicates and drops all but one.
>
> By adding `rand.Intn(1000)` nanoseconds, we create unique timestamps while maintaining sub-microsecond accuracy. The trade-off is slightly less precise timing, but for millisecond-level RTT measurements, this is negligible.
>
> Alternative solutions would be: (1) adding a VU ID tag, (2) using a counter-based field, or (3) switching to a time-series DB that handles collisions better."

### Q5: "How would you scale this system to test 100,000 concurrent users?"

**Model Answer**:
> "The current architecture supports up to 20,000 users via orchestration across 4 VMs. For 100,000, I'd:
>
> 1. **Distributed Redis**: Use Redis Cluster or migrate to something like DragonflyDB for higher throughput
> 2. **Remove localhost hardcoding**: Make Redis address configurable via environment variables
> 3. **Horizontal scaling**: 20 VMs with 5,000 users each, coordinated via the orchestrator pattern already in `orchestrator.sh`
> 4. **Connection pooling**: The TCP tunnel creation (`CsTCP`) should use a pool to avoid port exhaustion
> 5. **Metrics downsampling**: Increase the batch size from 100 to 1000 iterations for TCP metrics
> 6. **Pre-warm connections**: Establish TLS connections before the test starts
>
> The bottleneck will likely be the NetScaler Gateway, not this client - which is exactly what we want to test."

---

**Note**: Search results were limited to 10 results per query. For a more complete picture, you can [view more results in GitHub's code search](https://github.com/search?q=repo%3Acsg-citrix-secure-private-access%2Fwss-tools&type=code).
