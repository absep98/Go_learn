# Second Real Contribution — `-json-stats` flag on echo-server.go

File: `utils/echo-server.go` (same real repo file)

## The feature
Alongside the existing human-readable stats line, add a `-json-stats` flag
that outputs machine-readable JSON instead — useful for feeding stats into
monitoring tools instead of grepping log text.

## The pieces (all built step by step, each verified before moving on)

1. **New struct with JSON tags:**
   ```go
   type Stats struct {
       ActiveConns int64 `json:"active_conns"`
       TotalConns  int64 `json:"total_conns"`
       BytesIn     int64 `json:"bytes_in"`
       BytesOut    int64 `json:"bytes_out"`
   }
   ```
   First attempt used `string` fields — wrong. Fields must be `int64` to
   match what `atomic.LoadInt64` returns; JSON has native number support,
   no need to stringify.

2. **New flag type — `flag.Bool`** (hadn't used this one yet, only Int/Duration):
   ```go
   jsonStats := flag.Bool("json-stats", false, "Output stats as JSON instead of plain text")
   ```

3. **Threaded through** `main → statsLogger`, same pattern as `timeout` before.

## Three real mistakes caught during review (all fixed before running)

**Mistake 1 — wrong package.** First attempt defined `Stats` and a
`NewStats()` constructor inside `vpn.go` — a completely different Go
package (`package vpn`) that has no access to `echo-server.go`'s
`activeConns`/`totalConns` package-level variables. Moved everything into
`echo-server.go` (`package main`) where it belongs.

**Mistake 2 — unnecessary constructor.** Used the `New<Type>()` pattern
learned from `NewClassicAuth`/`NewVPN`, but that pattern earns its keep
when a struct is built in multiple places or needs setup logic. Here
`Stats` is built in exactly one place with zero logic — a plain struct
literal `Stats{Field: value, ...}` is simpler and more idiomatic.
**Lesson: recognize a pattern, but don't apply it where it adds no value.**

**Mistake 3 — two bugs in the same 5 lines:**
```go
if err != nil {
    log.Printf("Failed to marshal struct to json")
    return          // BUG: exits the ENTIRE statsLogger goroutine forever,
}                    // not just this tick. One marshal failure = stats logging
                     // silently dies for the rest of the program's life.
log.Printf(string(data))   // BUG: Printf's first arg is a FORMAT STRING.
                            // If `data` ever contains a literal `%`, Printf
                            // tries to interpret it as a format verb and
                            // prints garbled output (e.g. %!s(MISSING)).
```
Fixed: removed `return` (let the loop naturally continue to the next tick),
changed to `log.Printf("%s", string(data))` — explicitly passing `data` as
an argument to `%s` instead of as the format string itself.

**Rule to keep permanently: never pass a dynamic/variable string directly
as `Printf`'s first argument.** Use `Println(x)` or `Printf("%s", x)`.

## Verified for real
```bash
go build -o echo-server-test.exe echo-server.go   # compiles clean
./echo-server-test.exe -start 19091 -end 19091 -stats 2s -json-stats
```
Sent 39 bytes from a Python client, observed the live log:
```
{"active_conns":0,"total_conns":0,"bytes_in":0,"bytes_out":0}   # before connect
{"active_conns":1,"total_conns":1,"bytes_in":39,"bytes_out":39} # during connection
{"active_conns":0,"total_conns":1,"bytes_in":39,"bytes_out":39} # after disconnect
                                                                  # (total_conns correctly
                                                                  #  stays 1 — cumulative)
```
Exact match to real traffic. Valid JSON, correct field names, accurate numbers.

## Pattern reinforced from vpn.go reading
`json.Marshal` (struct → JSON bytes) is the exact mirror of `json.Unmarshal`
(JSON bytes → struct) seen in `vpn.go`'s `Cfg()` method one session earlier.
Reading real code first, then building the opposite operation immediately
after, made this land fast — no new concept, just the reverse direction.
