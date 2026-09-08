# Quick Review — Cold-Recall Question Bank

Purpose: fight "drift" by RETRIEVING concepts from cold memory (retrieval cements
far better than re-reading). Run through this in ~10 min whenever you feel rusty,
or at the start of a session after a gap. Cover the answers, say your answer out
loud or type it, THEN check. Getting one wrong is good — that's the one to redo.

How to use: don't read the answers first. Struggling to recall IS the exercise.

---

## Pointers (02)
Q1. Difference between `c := a` and `c := &a`?
Q2. `*p` vs `&x` — which is "address of" and which is "value at address"?
Q3. Why does `flag.Int(...)` return `*int` instead of `int`?

<details><summary>answers</summary>
A1. `c := a` copies the value (independent); `c := &a` makes c point at a (sees live changes).
A2. `&x` = address of x. `*p` = value at the address p holds (dereference).
A3. flag.Int runs BEFORE flag.Parse(); it returns a pointer so Parse() can write the
    parsed value into that memory later. A plain int couldn't be updated afterward.
</details>

## Atomics & Race (03)
Q4. Why does `counter++` lose counts across goroutines but `atomic.AddInt64` doesn't?
Q5. Is `if atomic.Load(&x) >= 50 { atomic.Add(&x,-50) }` safe with many goroutines?
Q6. Why must an atomic counter be `int64`, not plain `int`?

<details><summary>answers</summary>
A4. counter++ is 3 steps (read,add,write) — two goroutines can read the same value and
    one increment is lost. atomic does all 3 as ONE indivisible CPU instruction.
A5. NO — TOCTOU bug. Each atomic call is safe alone, but the GAP between check and modify
    lets multiple goroutines pass the check. Need a mutex or CompareAndSwap.
A6. sync/atomic needs the exact size to pick the CPU instruction; plain int varies by platform.
</details>

## Goroutines & WaitGroup (04)
Q7. Where does `wg.Add(1)` go — before or after `go`? Why?
Q8. Where does `wg.Done()` go and with what keyword?
Q9. Forget `wg.Done()` entirely — what happens?

<details><summary>answers</summary>
A7. BEFORE `go`, in the parent. If inside the goroutine, wg.Wait() may run while counter
    is still 0 (goroutine hasn't started) → program exits early.
A8. `defer wg.Done()` as the FIRST line — defer guarantees it runs on any return path.
A9. Counter never hits 0 → wg.Wait() blocks forever (hang / deadlock).
</details>

## Channels (05)
Q10. Does this deadlock? `ch := make(chan int); ch <- 42; fmt.Println(<-ch)`
Q11. Same but `make(chan int, 1)` — deadlock or works?
Q12. `close(ch)` vs `ch <- true` to signal — what's the difference?
Q13. What does `<-chan struct{}` mean in a function parameter?

<details><summary>answers</summary>
A10. DEADLOCKS — unbuffered send blocks with no receiver in another goroutine.
A11. WORKS — buffer of 1 holds the value, no second goroutine needed.
A12. close() wakes ALL receivers at once (broadcast); ch<-true wakes only ONE.
A13. A RECEIVE-ONLY channel — the function can read from it but not send (compiler-enforced).
</details>

## Signals & Shutdown (06)
Q14. What makes each acceptLoop goroutine exit when you press Ctrl+C?
Q15. Why is the signal channel buffered (size 1)?

<details><summary>answers</summary>
A14. ln.Close() → the blocked ln.Accept() returns an error → `if err != nil { return }`
    → defer wg.Done(). You don't kill goroutines; you remove what they wait on.
A15. So a signal arriving before the code reaches <-sigCh isn't dropped — it waits in the buffer.
</details>

## Interfaces & Receivers (07, 08)
Q16. What makes a type "satisfy" an interface in Go?
Q17. `.(type)` vs `.(ConcreteType)` — which discovers, which extracts?
Q18. Method has `func (c *Circle) Area()`. Does `var s Shape = Circle{}` compile?
Q19. `type Athena struct { Base }` (no field name) — what does that do?

<details><summary>answers</summary>
A16. Just having the required method(s) with matching signatures. No "implements" keyword.
A17. `.(type)` discovers ("what is it?", only in a switch); `.(ConcreteType)` extracts/unwraps.
A18. NO — pointer receiver means only *Circle satisfies Shape. Needs `&Circle{}`.
A19. Embedding — Base's fields AND methods are promoted onto Athena (Go's alt to inheritance).
</details>

## util package (11)
Q20. `interface{}` (empty) vs `Auth interface { DoAuth() }` — meaning of each?
Q21. Why does SafeMap take a `cleanupFn func(interface{})` instead of hardcoding cleanup?
Q22. What does the `unsafe` prefix in `unsafeRemove` signal?
Q23. Why RWMutex instead of plain Mutex for SafeMap?

<details><summary>answers</summary>
A20. Empty interface{} = ANY value. Named interface with methods = a CONTRACT.
A21. SafeMap holds interface{} (anything) so can't know how to clean a value (conn→Close,
    string→nothing). Caller supplies the "what"; SafeMap handles the "when". A callback.
A22. "This method does NOT lock — the caller must already hold the lock." Avoids double-lock deadlock.
A23. Map is read constantly, written rarely. RWMutex lets many reads run in parallel;
    only writes are exclusive. Plain Mutex would serialize even simultaneous reads.
</details>

## mux (17)
Q34. `b[4] = 42` vs `b[4:8] = someSlice` — which works, which doesn't, and what do you use instead?
Q35. Why does `Deserialize` need TWO separate `reader.Read` calls, not one big read?

<details><summary>answers</summary>
A34. Single-index assignment `b[4] = 42` works fine. Range assignment `b[4:8] = ...`
     is NOT valid Go — the compiler rejects it. To write into a range, use
     `copy(dst, src)` — a built-in that fills the destination window with source bytes.
A35. Length-prefixed protocol. First read gets the fixed 20-byte header. Header
     contains a `Length` field. Second read grabs exactly that many bytes for the
     payload. You can't read the payload without first knowing its size.
</details>

## protocol/dns (16)
Q32. `r.A` is `[4]byte` and `net.IP()` wants `[]byte`. What does `r.A[:]` do and why?
Q33. Inside a for-loop over DNS answers, why does non-A records trigger `continue` not `return`?

<details><summary>answers</summary>
A32. Slices the whole fixed-size array into a dynamic slice. `[4]byte` (array) and
     `[]byte` (slice) are DIFFERENT types — Go won't auto-convert. `[:]` bridges them.
A33. DNS responses can carry multiple record types (A, CNAME, AAAA, MX). `continue`
     skips the ones we don't want and keeps scanning; `return` would exit the whole
     function on the first non-A record and miss real A records further down the list.
</details>

## configuration helpers (15)
Q30. Why does `GetInt` use a type-switch (int/int64/float64) instead of `v.(int64)`?
Q31. `port := int64(8080); DestinationPort uint16` — what does the code have to write, and why?

<details><summary>answers</summary>
A30. JSON parsers hand numbers back inconsistently — sometimes int64, sometimes float64,
     sometimes int. A plain assertion `v.(int64)` would fail for the other two. The
     switch accepts all three and normalizes to int64. Discover → convert → return one type.
A31. `uint16(port)` — explicit conversion. Go never auto-converts between numeric types
     even when no data is lost. Same rule as `int64(maxConns)` from the -max-conns day.
</details>

## pacer package (14)
Q27. `Pace(elapsed) (time.Duration, time.Duration, bool)` — what's each return?
Q28. Why does `defer cp.completed++` sit MID-function, not at the top?
Q29. When `Pace` returns delay=0, what real situation is it handling?

<details><summary>answers</summary>
A27. delay (how long to sleep), expectedDuration (when this req was "supposed" to
     fire — for metrics), stop bool (done? exit the loop).
A28. Position IS the logic: `defer` only runs if code reaches the `defer` statement.
     Returns ABOVE it (time's up, rate=0, all done) don't bump the counter; only
     returns BELOW it (actually pacing a real request) do. Elegant.
A29. We're BEHIND schedule (completed < expectedCount). Fire immediately → burn
     through the backlog → catch up. Self-correcting rate control.
</details>

## error package (12)
Q24. What ONE method must any type have to count as a Go `error`?
Q25. What does `args ...interface{}` (variadic) let a function do?
Q26. `type RootCauseT = string` vs `type Duration int64` — difference?

<details><summary>answers</summary>
A24. `Error() string` — takes nothing, returns a string. That's the whole `error` interface.
A25. Accept zero-or-more args of any type; inside, `args` is a slice. Enables both
     `Errorf(rc, err)` and `Errorf(rc, "Status %d", 404)`. At a call site, `args...` spreads a slice back out.
A26. `=` is a type ALIAS (RootCauseT IS string, interchangeable). No `=` is a distinct NAMED
     type (Duration is NOT interchangeable with int64, e.g. won't fit an int64 param).
</details>

---

## Gotcha rules (the ones that bite in real code)
- Named type != underlying type for params: `time.Duration` won't fit an `int` param.
- Never `Printf(dynamicString)` — it's a format string; use `Println(x)` or `Printf("%s", x)`.
- `return` inside a for+select goroutine exits the WHOLE function, not just that tick.
- Code after an infinite `for {}` is unreachable — use `defer`.
- `c := a` copies (even for structs); only pointers alias.
- Go doesn't auto-convert numeric types — `int64(x)` explicitly when comparing/passing.
- `flag.Int` returns `*int` — pass `*flagVar` (deref), not `&flagVar` (would give `**int`).
- `continue` vs `return` in a `for` loop: continue skips this iteration; return kills the whole function.
