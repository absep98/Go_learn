# Go Learning Journey

## Files at the root
- `README.md` — this index
- `QUICK_REVIEW.md` — cold-recall question bank (35+ Qs, hidden answers) — run before any new session
- `MISTAKES_LOG.md` — bugs I hit while writing Go from a blank page + the rule for each — review BEFORE every blank-page exercise
- `REPO_COVERAGE_MAP.md` — which files in wss-tools I've read and understood, tracked

## How to Run
```bash
go run 02_pointers/pointers.go
go run -race 03_atomic_race/race_demo.go   # -race flag catches data races
```

## Learning Path

### 01_basics/ — Variables, constants, control flow
- `variables.go` — var, :=, types, zero values
- `constants.go` — const blocks
- `control_flow.go` — if/else, for, switch

### 02_pointers/ — Memory addresses and indirection
- `pointers.go` — &x (address of), *p (dereference), value copy vs alias

### 03_atomic_race/ — Concurrency safety
- `race_demo.go` — why counter++ fails, why atomic.AddInt64 works (run with -race)
- `toctou_bug.go` — check-then-act is NOT safe even with atomic (TOCTOU)

### 04_goroutines_waitgroup/ — Spawning concurrent work
- `waitgroup.go` — go keyword, WaitGroup contract (Add before go, defer Done)

### 05_channels/ — Communication between goroutines
- `01_unbuffered.go` — send/receive must be different goroutines
- `02_buffered.go` — buffer lets send/receive in same goroutine
- `03_close_broadcast.go` — close() wakes ALL receivers, <-chan (receive-only type)

### 06_signals_shutdown/ — Graceful server shutdown
- `graceful_shutdown.go` — signal.Notify, cooperative shutdown chain

### 07_interfaces/ — Contracts, type assertion, pointer receivers
- `interfaces.go` — Shape/Circle example: interface satisfaction, type switch, comma-ok assertion
- `NOTES.md` — mapped to real code in xk6-wss/auth/auth.go + classic.go; code-reading strategy for unfamiliar files

### 08_struct_embedding/ — Composition over inheritance
- `embedding.go` — Base/Athena example: field & method promotion vs named fields
- `NOTES.md` — mapped to real athena.go (520 lines); how to break down a big file via its ONE interface entry point; Classic vs Athena comparison

### 09_first_contribution/ — Real change, verified end-to-end
- `NOTES.md` — `-timeout` flag: type bug caught (int vs time.Duration), Go SDK installed, built + tested live, measured 3.0s drop matching the flag
- `02_json_stats.md` — `-json-stats` flag: json.Marshal (mirror of Unmarshal seen in vpn.go), wrong-package mistake caught, unnecessary-constructor caught, return-kills-goroutine bug caught, Printf format-string trap caught — verified live with real byte counts
- `03_max_conns.md` — `-max-conns` flag: reject-when-full backpressure gate, `&flag` vs `*flag` bug caught, Go's no-auto-numeric-conversion rule, continue-vs-return distinction — verified live with 3 sockets (2 held, 3rd rejected with WinError 10054)

### 11_util_package/ — Shared helpers used everywhere
- `NOTES.md` — http.go (SendHTTPRequest, hidden CheckRedirect, DumpHttp) + safemap.go (RWMutex, cleanupFn callback, unsafeRemove, empty vs named interface{})

### 12_error_package/ — Custom errors + lerr.Errorf
- `NOTES.md` — Go's `error` is just `interface { Error() string }`; custom Error type with RootCause tag; Errorf variadic (`...interface{}`) + type switch; type alias (`=`) vs named type

### 13_design_patterns/ — Named reusable designs (grows over time)
- `NOTES.md` — catalog; `01_factory/factory.go` — Factory pattern (runnable), from auth/configuration.go's FromConfiguration

### 14_pacer_package/ — Load-test rate pacer
- `NOTES.md` — Pacer interface (3 returns: delay, expectedDuration, stop), Factory again, defer-placement-as-logic, self-correcting sleep=0 catch-up

### 15_configuration_helpers/ — Typed config getters
- `NOTES.md` — MILESTONE: read all 145 lines without new concepts. GetInt type-switch for int/int64/float64. Compose-smaller-helpers pattern.

### 16_protocol_dns/ — DNS query builder + response parser
- `NOTES.md` — closes `protocol/` folder; nested struct literals demystified; `[4]byte` array vs `[]byte` slice + `r.A[:]` conversion; filter-then-extract loop with continue

### 17_mux/ — Custom binary protocol (serialize/deserialize)
- `NOTES.md` — MuxMessage header layout; **key rule: slice RANGE assignment doesn't exist — use `copy(dst, src)`**; big-endian for network order; length-prefixed read pattern

### practice_problems/ — Exercises
- `problem1_calculator.go` — Basic functions and operations
- `problem2_loops.go` — Loops and mathematical calculations

## Key Rules Learned

| Rule | Why |
|------|-----|
| `atomic` for single counters, `mutex` for compound operations | Atomic = 1 CPU instruction. Two atomics with a gap = TOCTOU bug |
| `wg.Add(1)` BEFORE `go`, never inside | Race with `wg.Wait()` — goroutine might not start in time |
| `defer wg.Done()` as FIRST line | Guarantees cleanup even on early return or panic |
| Never put code after `for { }` | Unreachable — use `defer` instead |
| `c := a` copies value, `c := &a` aliases | Only pointers see live updates |
| Unbuffered channels need 2 goroutines | Same-goroutine send+receive = deadlock |
| `close(ch)` to broadcast, `ch <- val` for one receiver | close wakes ALL, send wakes ONE |
| Don't kill goroutines — remove what they wait on | ln.Close() → Accept() errors → goroutine exits naturally |
| Pointer receiver → only pointer type satisfies interface | `(c *Circle) Area()` means `var s Shape = Circle{}` fails to compile, needs `&Circle{}` |
| `.(type)` discovers, `.(ConcreteType)` extracts | Type switch asks "what is it?"; type assertion unwraps it to use |
| Hunt structs/interfaces/methods FIRST, logic SECOND | Reading unfamiliar Go top-to-bottom is overwhelming; find the nouns and verbs first |
| `type X struct { Base }` (no name) promotes fields/methods | `a.vpnFQDN` works directly even though it's declared inside `Base` — Go's alternative to inheritance |
| A big file ≠ a hard file | Find the ONE method the interface requires (e.g. `DoAuth`), trace its call order, treat the rest as fillable detail |
| A named type != its underlying type for FUNCTION PARAMETERS too | `time.Duration` is `int64` underneath, but a param declared `int` rejects it — same rule as `http.Header` |
| Recognize a pattern, but don't force it where it adds no value | `New<Type>()` constructor is great for multi-use/setup-needing structs; a single-use struct literal is simpler |
| `return` inside a `for`+`select` goroutine exits the WHOLE function | One error inside a ticker case can silently kill a background logger forever — don't `return`, just let the case end |
| Never pass a dynamic string as `Printf`'s first argument | It's parsed as a format string — stray `%` breaks output. Use `Println(x)` or `Printf("%s", x)` |

## Status (as of 2026-08-14)
Core Go language fundamentals are DONE: pointers, atomics/race, goroutines/WaitGroup,
channels/signals, interfaces/type-assertion, pointer receivers, struct embedding.
Two real contributions shipped and verified live in `utils/echo-server.go`
(-timeout, -json-stats). See `REPO_COVERAGE_MAP.md` for full-repo tracking —
goal is BOTH Go mastery (largely done) AND full depth on this specific codebase
(ongoing, sequenced file by file). Remaining work is repo-specific reading
practice, not new language theory: `vpn.go` (13 unread methods), `protocol/{tcp,udp,dns}`,
`util/`, `pacer/`, `error/`, `auth/configuration.go`, and `wss.go` (CepGwAccess section).

**First real contribution shipped:** added `-timeout` flag to `utils/echo-server.go`
(threaded time.Duration through acceptLoop → handleConnection, applied via
conn.SetDeadline before each Read). Go SDK installed on this machine same day
(was missing — only gopls existed). Built with `go build`, tested live with a
silent Python TCP client — measured server dropped the connection at exactly
3.0s for `-timeout 3s`. Full loop closed: idea → code → compile → run → verify.

Next: continue reading `vpn.go`, or attempt another small real change.
