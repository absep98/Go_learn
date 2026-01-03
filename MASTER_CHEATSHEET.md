# Go Master Cheat Sheet (Repo-Driven)

This is a single revision sheet built from the examples in this repo.

## How to use this sheet

- For each topic, skim the **Patterns** and **Gotchas**, then run the linked example file.
- Run any file from repo root:

```bash
go run ./01-basics/hello.go
```

---

## 1) Basics (types, vars, zero values)

**Key examples**
- [01-basics/hello.go](01-basics/hello.go)
- [01-basics/types.go](01-basics/types.go)

**Patterns**
- Declaration:
  - `var x int = 10`
  - `x := 10` (short declaration inside functions)
  - `const pi = 3.14`
- Zero values (when not initialized):
  - `int` → `0`, `float64` → `0.0`, `bool` → `false`, `string` → `""`
- Type conversions are explicit:
  - `float64(i)`
  - `fmt.Sprintf("%d", i)`

**Gotchas**
- Integer division: `9/5 == 1` (use `9.0/5.0` when you want float math).

---

## 2) Functions (returns, errors, variadic, first-class)

**Key examples**
- [01-basics/functions.go](01-basics/functions.go)
- [08-errors/day6-functions-errors.go](08-errors/day6-functions-errors.go)

**Patterns**
- Multiple returns: `(value, err)` is the Go standard.
- Early return on error:

```go
res, err := doThing()
if err != nil { return err }
```

- Named returns exist, but use them carefully (readability > cleverness).
- Variadic params: `func sum(nums ...int) int`
- Functions as values:

```go
func apply(a, b int, op func(int, int) int) int { return op(a, b) }
```

**Error handling patterns**
- Create: `errors.New("msg")`
- Format: `fmt.Errorf("bad input: %v", err)`

---

## 3) Control flow (if/for/switch/type switch)

**Key examples**
- [02-control-flow/loops.go](02-control-flow/loops.go)
- [02-control-flow/switch.go](02-control-flow/switch.go)

**Patterns**
- `if` can include an init statement:

```go
if score := 85; score >= 80 { /* ... */ }
```

- Go has one loop keyword: `for` (classic / while-style / infinite).
- `range` over slices: `for i, v := range s {}`
- Switch without expression:

```go
switch {
case x > 10:
case x > 5:
}
```

- Type switch:

```go
switch v := any.(type) {
case string:
case int:
case nil:
}
```

**Gotchas**
- `fallthrough` ignores the next case condition (rarely needed).

---

## 4) Arrays & Slices (len/cap/append/copy)

**Key examples**
- [03-data-structures/arrays-slices.go](03-data-structures/arrays-slices.go)
- [11-interview-prep/interview-challenge.go](11-interview-prep/interview-challenge.go)

**Patterns**
- Arrays: fixed size `[5]int`.
- Slices: dynamic `[]int`.
- `make([]T, len, cap)` controls length/capacity.
- Copy to avoid sharing underlying array:

```go
dst := make([]int, len(src))
copy(dst, src)
```

**Gotchas (interview favorites)**
- `append` returns the new slice:

```go
append(s, 1) // does NOTHING useful unless you assign
s = append(s, 1)
```

- Slice sharing: `s2 := s1[1:3]` shares the same backing array; mutation can affect both.

---

## 5) Maps (make, ok idiom, nil map, zero values)

**Key examples**
- [03-data-structures/maps.go](03-data-structures/maps.go)
- [03-data-structures/nil-map-panic.go](03-data-structures/nil-map-panic.go)
- [03-data-structures/map-zero-values.go](03-data-structures/map-zero-values.go)

**Patterns**
- Create:

```go
m := make(map[string]int)
// or
m := map[string]int{"a": 1}
```

- Read missing key → zero value of the map’s value type.
- Existence check (`ok` idiom):

```go
v, ok := m["k"]
```

- Delete: `delete(m, "k")`

**Gotchas**
- `var m map[string]int` is a **nil map**:
  - Reading is OK (returns zero value)
  - Writing panics: `panic: assignment to entry in nil map`

---

## 6) Structs & Methods (value vs pointer receiver)

**Key examples**
- [03-data-structures/structs.go](03-data-structures/structs.go)
- [03-data-structures/struct-with-pointers.go](03-data-structures/struct-with-pointers.go)

**Patterns**
- Struct literal (named fields preferred):

```go
t := Task{ID: 1, Description: "x"}
```

- Pointer receivers when the method needs to mutate state or avoid copying.

**Gotchas**
- Embedded structs: field name collisions require qualifying (see `Employee.Person.City` vs `Employee.Address.City`).

---

## 7) Interfaces (polymorphism + nil-interface trap)

**Key examples**
- [03-data-structures/interfaces.go](03-data-structures/interfaces.go)
- [03-data-structures/interfaces-practical-usecases.go](03-data-structures/interfaces-practical-usecases.go)
- [07-Pointers/nil-pointers-clean.go](07-Pointers/nil-pointers-clean.go)

**Patterns**
- Interfaces are satisfied implicitly (no `implements` keyword).
- Prefer accepting interfaces, returning concrete types.

**Gotchas (very common in interviews)**
- **Nil-interface trap**:
  - A nil *pointer* stored in an interface is **not** a nil interface.
  - Interface value is a pair: `(dynamicType, value)`.

---

## 8) Pointers & nil safety

**Key examples**
- [07-Pointers/nil-pointers-clean.go](07-Pointers/nil-pointers-clean.go)

**Patterns**
- Always nil-check before dereference.
- It’s OK to write methods that handle nil receivers when that’s useful.

---

## 9) Errors (wrap, propagate, validate)

**Key examples**
- [08-errors/errors.go](08-errors/errors.go)
- [08-errors/day6-functions-errors.go](08-errors/day6-functions-errors.go)

**Patterns**
- Wrap errors with context (keeps original error):

```go
return fmt.Errorf("load config failed: %w", err)
```

- Validate inputs early; return descriptive errors.

---

## 10) Packages & Modules

**Key examples**
- [09-packages/go-modules-guide.txt](09-packages/go-modules-guide.txt)
- [09-packages/myapp/](09-packages/myapp/)

**Patterns**
- `go mod init <module>` to start a module.
- `go mod tidy` to sync dependencies.

---

## 11) JSON Marshalling / Unmarshalling

**Key examples**
- [12-marshalling/marshalling-intro.go](12-marshalling/marshalling-intro.go)

**Patterns**
- Struct tags define JSON field names:

```go
type Config struct {
  AppName string `json:"app_name"`
}
```

- Unmarshal into a struct pointer: `json.Unmarshal(data, &cfg)`

**Gotchas**
- JSON only fills **exported** fields (must start with a capital letter).

---

## 12) Concurrency (goroutines, channels, worker pool)

**Key examples**
- [05-concurrency/goroutines-basic.go](05-concurrency/goroutines-basic.go)
- [05-concurrency/channels-basic.go](05-concurrency/channels-basic.go)
- [05-concurrency/worker-pool.go](05-concurrency/worker-pool.go)

**Patterns**
- Goroutine: `go fn()` runs concurrently.
- Channels communicate between goroutines:
  - Send: `ch <- x`
  - Receive: `x := <-ch`
  - Range receive until close: `for v := range ch {}`
- Worker pool pattern:
  - `jobs` channel in, `results` channel out
  - `sync.WaitGroup` to wait
  - Close channels to signal completion

**Gotchas**
- Avoid `time.Sleep` as synchronization in real code; prefer channels/WaitGroup.

---

## 13) Small projects (practice integrating everything)

**Key examples**
- [06-projects/todo-list.go](06-projects/todo-list.go)
- [06-projects/number-guessing-game.go](06-projects/number-guessing-game.go)

**What these teach**
- Struct + methods + slice manipulation
- Input parsing (`Scanln` / `Scanf`) and branching
- State updates + simple UX loop

---

## 14) Interview rapid-fire checklist (what to say out loud)

- Slices share backing arrays; `append` may reallocate.
- Maps: missing key returns zero value; use `value, ok` to check existence.
- Nil map: reads OK, writes panic.
- Interface nil trap: `(type, value)` pair.
- Pointer receiver methods affect interface satisfaction.
- Errors: return them, wrap with `%w`, don’t ignore them.
- Concurrency: avoid data races; coordinate with channels/WaitGroup.

---

## Suggested 30-minute revision route

1) Basics + functions: run [01-basics/types.go](01-basics/types.go) and [01-basics/functions.go](01-basics/functions.go)
2) Slices + maps gotchas: run [03-data-structures/arrays-slices.go](03-data-structures/arrays-slices.go) and [03-data-structures/nil-map-panic.go](03-data-structures/nil-map-panic.go)
3) Interfaces/pointers: run [03-data-structures/interfaces.go](03-data-structures/interfaces.go) and [07-Pointers/nil-pointers-clean.go](07-Pointers/nil-pointers-clean.go)
4) Concurrency: run [05-concurrency/worker-pool.go](05-concurrency/worker-pool.go)
5) Final interview drill: run [11-interview-prep/interview-challenge.go](11-interview-prep/interview-challenge.go)
