# First Real Contribution — `-timeout` flag on echo-server.go

File: `utils/echo-server.go` (real repo file, not a practice copy)

## The problem
Before this change, if a client connected and went silent (no data, no
disconnect), `handleConnection`'s `conn.Read(buf)` blocked FOREVER — leaking
a goroutine + socket permanently. At scale, dead connections would slowly
starve the server.

## The fix — 4 small steps
1. **Declared the flag** (same shape as existing `-stats` flag):
   ```go
   timeout := flag.Duration("timeout", 0, "Per-connection read/write timeout (0 to disable)")
   ```
2. **Threaded it through** main → acceptLoop → handleConnection, same pattern
   as the existing `bufSize` parameter.
3. **Type bug caught during threading:** first attempt declared the parameter
   as `timeout int`, but `flag.Duration` returns `*time.Duration` — a DISTINCT
   named type from `int`, even though it's `int64` underneath. Same lesson as
   `http.Header = map[string][]string` — named types are not interchangeable
   with their underlying type. Fixed: `timeout time.Duration` in both signatures.
4. **Applied the deadline** inside the read loop:
   ```go
   if timeout > 0 {
       conn.SetDeadline(time.Now().Add(timeout))
   }
   ```
   Used `SetDeadline` (not `SetReadDeadline`) — covers BOTH read and write,
   matching the flag's own description ("read/write timeout").
   `time.Now().Add(timeout)` — SetDeadline wants an ABSOLUTE point in time,
   not a duration, so you compute "now + timeout" every iteration.

## Verified for real — not just "looks right"
Go wasn't even installed on this machine before this session (only `gopls`,
the editor's language server, existed — that's why VS Code could flag syntax
errors but nothing could actually build). Installed via `winget install GoLang.Go`.

Then:
```bash
go build -o echo-server-test.exe echo-server.go   # compiles clean
./echo-server-test.exe -start 19090 -end 19090 -timeout 3s   # run it
```
Wrote a tiny Python client that connects and stays silent:
```python
s = socket.socket(...)
s.connect(('localhost', 19090))
data = s.recv(1024)   # blocks until server sends data OR closes
# data == b'' means server closed the connection
```
**Result: connection dropped at exactly 3.0 seconds** — matching `-timeout 3s` exactly.

## Duration flag format (the thing that was confusing)
`flag.Duration` does NOT accept a bare number. Requires a unit suffix:
```
-timeout 60s     ✅ 60 seconds
-timeout 5m      ✅ 5 minutes
-timeout 60      ❌ error: "missing unit in duration"
```

## The full loop this proved
idea → write code → compiler catches a real type bug → fix →
build succeeds → run the real binary → hit it with a real client →
MEASURE the actual behavior → confirms the feature works.

This is the difference between "I read the code and it looks right" and
"I proved the code does what I claimed."
