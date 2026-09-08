# Mistakes Log — patterns that bit me in blank-page writing

Reviewed BEFORE starting each new writing exercise. Each entry:
- what I wrote wrong
- what the compiler / runtime said
- the rule
- where I've seen this same rule before in the codebase

---

## Session 1 — `wordcount.go` (2026-09-01)

### M1 — Compared `*string` (from flag) directly to a string literal
```go
filename := flag.String("file", "empty", "...")
if filename == "empty" { ... }   // ❌ mismatched types *string and string
```
**Compiler:** `invalid operation: filename == "empty" (mismatched types *string and untyped string)`

**Rule:** `flag.String`/`flag.Int`/`flag.Duration` all return **pointers**. To compare or use the VALUE, dereference with `*`:
```go
if *filename == "empty" { ... }
```

**Seen before in real code:** every `main()` in echo-server.go — `*bufSize`, `*timeout`, `*maxConns` all dereferenced before use.

---

### M2 — Wrote function-SIGNATURE syntax at the CALL site
```go
fileData, err := os.ReadFile(*filename)([]byte, error)   // ❌ garbage syntax
```
The `([]byte, error)` is what appears in `func` DECLARATIONS to describe returns. At the call site, you just capture with `:=`:
```go
fileData, err := os.ReadFile(*filename)   // ✅
```
**Rule:** Return types are declared once (in the function's signature). At every call site, you just receive them into variables. Never annotate at the call site.

---

### M3 — Declared a variable, then never used it
```go
fileData, err := os.ReadFile(*filename)   // fileData used nowhere below
```
**Compiler:** `declared and not used: fileData`

**Rule:** Go treats **unused local variables as a HARD compile error** (unlike Python/JS which allow it). If you truly need to ignore a value, use the blank identifier `_`:
```go
_, err := os.ReadFile(*filename)   // "I don't need the data, just the error"
```
**Seen before:** same rule caught `data, err := json.Marshal(...)` when `data` was unused in the `-json-stats` work.

---

### M4 — Forgot `flag.Parse()`
Without `flag.Parse()`, ALL flags stay at their default values — command-line args are never read. Silent bug (compiles fine, just always uses defaults).

**Rule — the flag lifecycle:**
```go
myFlag := flag.X("name", default, "help")   // 1. Declare (returns pointer)
flag.Parse()                                 // 2. Parse command line into all declared flags
value := *myFlag                             // 3. Now safe to read
```
Missing step 2 = flag never gets user's value.

---

### M5 — Passed `[]byte` where `string` was required
```go
fileData, _ := os.ReadFile(...)          // returns []byte
strings.Fields(fileData)                  // ❌ Fields wants string
```
**Compiler:** `cannot use fileData (variable of type []byte) as string value`

**Rule:** Go **never auto-converts between related types** — even ones that share the same memory layout. Explicit wrap:
```go
strings.Fields(string(fileData))   // ✅
```
**Same rule seen before:** `uint16(port)` in tcp/configuration.go, `int64(maxConns)` in echo-server, `int64(n)` inside `GetInt`. **The rule is universal.**

---

### M6 — Runtime bug: `%d` with a slice (compiler doesn't catch this one)
```go
fileWords := strings.Fields(...)               // fileWords is []string
log.Printf("total word count %d", fileWords)   // prints garbage
```
`%d` expects a number. A slice is not a number. Compiles fine, but at runtime prints something like `%!d([]string=[hello world ...])`.

**Rule:** `%d` = integer, `%s` = string, `%v` = anything (safe default). For the COUNT of a slice, use `len(slice)`:
```go
log.Printf("total word count %d", len(fileWords))
```

**Lesson from this specific bug:** the compiler catches TYPE mismatches at function boundaries (`strings.Fields([]byte)` — caught). But it does NOT catch format-string mismatches (`Printf("%d", slice)` — silent). Format specifiers are checked at runtime, not compile time. So `Printf`-family calls need extra care.

---

## Meta-lesson from Session 1
All 6 bugs mapped to concepts I ALREADY KNEW. The problem wasn't knowledge — it was **fluency of application** when writing from a blank page. Recognition memory ≠ production memory. Only repeated blank-page writing closes this gap.

---

## Session 2 — `topwords.go` (2026-09-02)

### M7 — Used the TYPE name where the VARIABLE name belonged
```go
type wordCount struct { Word string; Count int }   // wordCount = the BLUEPRINT
var ranked []wordCount                              // ranked = the DATA

sort.Slice(wordCount, func(i, j int) bool {         // ❌ passed the blueprint
    return wordCount[i].Count > wordCount[j].Count  // ❌ indexed the blueprint
})
```
**Compiler:** `wordCount (type) is not an expression` / `wordCount is not a generic type`

**Rule:** A **type** is a blueprint (like a cookie cutter). A **variable** is the actual data (like a real cookie). `sort.Slice` operates on data, not on blueprints:
```go
sort.Slice(ranked, func(i, j int) bool {
    return ranked[i].Count > ranked[j].Count
})
```

**Trigger this mistake avoids:** if you named your type and variable the same (e.g. `type Ranked struct{...}; var ranked Ranked`), the confusion vanishes. Naming the type distinctly from the variable (`type wordCount` + `var ranked`) is idiomatic; just don't lose track of which is which.

---

### M8 — Iterated the WRONG container (compiler couldn't catch this)
```go
sort.Slice(ranked, ...)          // sorted `ranked`
for word, n := range counts {    // ...then threw it away by iterating the unsorted map
    ...
}
```
**Rule:** After sorting, always iterate the SORTED thing (the slice), not the source (the map). Map iteration order in Go is RANDOM by design — you'd get top-N in random order, wasting the sort work.

**Also compounded:** used a `break` inside the map iteration — meaningless because map order is random, so you'd break at some arbitrary word, not after N words.

**Rule of thumb:** if you `sort.Slice(X, ...)`, the very next loop should be `for i := 0; i < N; i++ { ... X[i] ... }` or `for _, x := range X`. Same X.

---

### M9 — Confused "count of a word" with "count of words to print"
```go
if n >= *topN { fmt.Println(word, n) }
```
`n` = "how many times THIS word appeared" (e.g. 7). `*topN` = "print top N different words" (e.g. 5). Comparing them (`7 >= 5`) is meaningless — different concepts. `n=7` doesn't make "hello" a top-5 word.

**Rule:** name variables so their meaning is unmistakable — `wordFreq >= wordsToShow` would have made this bug jump out. Rushed 1-letter variable names hide semantic mistakes.

---

### M10 — Didn't follow spec-specified defaults
Spec said `-top` default 5, I wrote default 0. Program appeared to work only because I always passed `-topN` explicitly at the CLI. Without the flag, it would silently print nothing.

**Rule:** re-read the spec before declaring the flag — spec defaults are part of the contract. Compiler can't catch spec-violation.

---

### M11 — Left `log.Println(*topN)` as leftover debug
Harmless, but pollutes output. Always sweep debug prints before saying "done."

---

## Meta-lesson from Session 2
Same theme: all 5 mistakes are things I could have caught by (a) reading each line critically, and (b) actually running the program with different inputs (missing `-topN`, etc.) before saying "done." **Testing the code isn't optional — it's the final debugging pass.**

---

## Session 3 — `grep-lite.go` (2026-09-02)

Big win: **compiled clean on first try.** All bugs were output-quality, not compile-blocking. The mistakes-log habit is paying off — I'm skipping the compile-error class of bugs now.

### M12 — Used `log.Printf` for USER-FACING output (should be `fmt.Printf`)
`log.Printf` prints a timestamp prefix and writes to stderr. Fine for diagnostics, wrong for user output that people might pipe (`| wc -l`, `| grep`).

**Rule:**
- `log.Printf` / `log.Println` / `log.Fatalf` = LOG events (errors, diagnostics, background info). Adds timestamp. Goes to stderr.
- `fmt.Printf` / `fmt.Println` = REGULAR user output. No timestamp. Goes to stdout.
- Keep `log.Fatalf` for error paths — errors ARE log-worthy. Use `fmt` for the actual results.

### M13 — Classic off-by-one on line numbers
```go
lineCount := 0
for scanner.Scan() {
    line := scanner.Text()
    // print... uses lineCount (still 0 for the first line!)
    lineCount++            // ← incrementing here made line 1 print as 0
}
```
**Fix:** increment at the TOP of the loop body:
```go
lineCount := 0
for scanner.Scan() {
    lineCount++            // ← first line becomes 1, matching every real tool
    line := scanner.Text()
    // print...
}
```
Real tools (`grep -n`, editors, error messages) all use 1-based line numbers. Convention.

### M14 — Missed a literal character in the format string
Spec said `LINENUM: LINETEXT`; I wrote `"%d %s\n"` — dropped the colon. Compiler can't catch missing literal chars in format strings; only spec-reading catches this. **Rule:** re-read the spec's exact output format before writing the Printf.

---

## Meta-lesson from Session 3
Zero compile errors on first try — recognition memory of the mistakes log is starting to work as a mental checklist. Remaining bugs were: (a) wrong stdlib PACKAGE for the job (log vs fmt), (b) off-by-one counter placement, (c) missing spec detail. All caught by actually RUNNING the program and comparing output to spec, not by staring at code.

---

## Session 4 — `pgrep.go` (2026-09-02)

Highlight: **self-recovered from 2 compile errors WITHOUT asking for help** — first time doing the read-compiler-error-and-fix cycle solo.

### M15 — Passed value instead of pointer to a shared coordinator
```go
go grepFile(path, *pattern, wg)     // ❌ passes wg BY VALUE (copy)
go grepFile(path, *pattern, &wg)    // ✅ passes wg BY POINTER (shared)
```
**Compiler:** `cannot use wg (variable of type sync.WaitGroup) as *sync.WaitGroup value`

**Rule:** Any type meant for coordination between goroutines — `sync.WaitGroup`, `sync.Mutex`, `sync.RWMutex`, channels held as struct fields — MUST be passed by pointer (`&x`), and the parameter type must be `*T`. Copying these silently breaks them: each goroutine mutates its own copy, and the "shared" state is never actually shared.

**The one check that catches this:** if a callee will call any state-mutating method (`Done`, `Lock`, etc.), the caller must pass with `&`.

### M7 recurrence — type-vs-variable AGAIN
Second occurrence in real code (first was `sort.Slice(wordCount, ...)` in Session 2). This time: passed `sync.WaitGroup` (the TYPE) as an argument where I meant `wg` (the VARIABLE). Fixed after compiler error.

### M11 recurrence — leftover debug prints AGAIN
`fmt.Println("all workers spawned...")` and `fmt.Println("all workers finished")` — copy-pasted from waitgroup.go tutorial code, forgot to strip them. Third session in a row this bug tried to slip through.

### M14 recurrence — missing spec detail in Printf format AGAIN
Spec said `FILENAME:LINENUM: LINETEXT`; my Printf was `"%d: %s\n"` (no filename). Second time.

---

## Meta-lesson from Session 4
The compile-error class of bugs is now self-fixable (M15 caught by compiler → I fixed it without help). Where I still lose: (a) **cleanup sweep** (M11 debug prints) and (b) **spec fidelity** (M14 missing filename). Both fixable by adopting a "cleanup sweep before saying done" ritual:
1. Re-read spec's exact output format
2. `Ctrl+F` for `Println` / `Printf` / debug comments — sweep any that aren't part of real output
3. Search for unused imports, `_ = x`, TODO markers

**Rule I need to burn in:** compiling ≠ done. Running once ≠ done. Matching spec output character-by-character AND cleaning debug garbage = done.

---

## Recurring-mistake heatmap (bugs that keep coming back)
| # | Bug | Sessions hit | Status |
|---|---|---|---|
| M7 | Type vs variable | 2 (Session 2, 4) | Still recurring — one more retrieval and it should stick |
| M11 | Leftover debug prints | 3 (Sessions 1, 2, 4) | Hardest to catch — need cleanup ritual |
| M14 | Missing spec detail in Printf | 3 (Sessions 2, 3, 4) | Hardest to catch — need spec re-read ritual |
