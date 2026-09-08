# Third Real Contribution — `-max-conns` flag on echo-server.go

File: `utils/echo-server.go`

## The feature
Cap the number of simultaneous open connections. Above the cap, immediately
close new connections instead of handling them. Prevents file-descriptor
or memory exhaustion at high load (real-world "backpressure").

Flag: `-max-conns N` (0 = unlimited, default).

## The reject-when-full gate
```go
if maxConns > 0 && atomic.LoadInt64(&activeConns) >= int64(maxConns) {
    conn.Close()
    continue                 // back to the top of the for loop — NOT return
}
```
Placed right after `if err != nil { return }`, before the counter bumps.

## Three real Go footguns caught during review

**1. `&maxConns` vs `*maxConns` at the call site.**
`flag.Int` already returns a `*int`. Writing `&maxConns` takes the address of
that pointer → `**int` (pointer to pointer). Must use `*maxConns` to
dereference and pass the value. Same lesson as `-timeout` day.

**2. Argument ordering — `wg` stays last.**
Idiom in this codebase: feature args (`bufSize`, `timeout`, `maxConns`) go
before the plumbing arg (`wg`). Keeps signatures readable & consistent.

**3. Go doesn't auto-convert numeric types — even when they'd fit.**
Comparing `atomic.LoadInt64(&activeConns)` (returns `int64`) with `maxConns`
(a plain `int`) fails to compile. Must write `int64(maxConns)` explicitly.
Same strictness as `time.Duration` vs `int`.

## The chain of forced choices (worth internalizing)
1. `activeConns` is shared across many goroutines → must be `int64`
   (only fixed-width types support atomic operations).
2. `atomic.LoadInt64` returns `int64` (contract).
3. `maxConns` is `int` (contract of `flag.Int`).
4. Go refuses to compare `int64` vs `int` → explicit `int64(maxConns)` needed.
Every downstream conversion is FORCED by step 1. "Once counter is shared,
everything else is inevitable."

## `continue` vs `return` — critical distinction
- `continue` → skip the rest of this loop iteration, go back to the top.
  Used here so the acceptLoop keeps taking new connections.
- `return` → exits the ENTIRE function (and triggers `defer wg.Done()`).
  Used only on real errors (like `ln.Close()` from shutdown).
Getting these confused = whole acceptLoop goroutine dies on first full-server
event, exactly the `return`-kills-goroutine bug pattern from `-json-stats` day.

## Verified live
```bash
go build -o echo-server-test.exe echo-server.go   # compiles clean
./echo-server-test.exe -start 19092 -end 19092 -max-conns 2
```
Python test opened 2 conns and HELD them (fills the limit), then tried a 3rd:
```
opened & holding conns 1 and 2
conn 3: error [WinError 10054] An existing connection was forcibly closed by the remote host
```
Exactly the intended behavior — server's gate triggered `conn.Close()` on the
3rd connection, the OS reported the reset back to the client. Feature working
as designed with real socket-level evidence.
