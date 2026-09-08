# pkg/configuration/configuration.go — type-safe config helpers

Real code: `xk6-wss/pkg/configuration/configuration.go` (145 lines)
The `config.GetInt(...)`, `config.GetDuration(...)`, `config.GetString(...)`
calls sprinkled throughout the repo (constant.go, etc.) all live here.

## Nothing new to learn — that's the milestone
Every function is the SAME shape you've already seen many times:
1. Look up a key in `map[string]interface{}` (comma-ok)
2. Type-assert the value to the expected type (comma-ok again)
3. Return a clean typed value OR a wrapped error

Read the whole 145-line file end-to-end without help = real progress marker.

## The two things worth naming

### `GetInt` uses a TYPE SWITCH (not a plain type assertion)
```go
switch n := If.(type) {
case int64:   return n, nil
case float64: return int64(n), nil     // JSON parser sometimes gives float64
case int:     return int64(n), nil
default:      return 0, fmt.Errorf("...must be a number")
}
```
JSON decoders in Go can hand you a number as `int`, `int64`, or `float64`
depending on where the config came from. Instead of demanding one specific
type and failing, GetInt accepts all three and normalizes to `int64`.
Pragmatic — the caller doesn't need to know the JSON library's quirks.

### `GetSliceOfStrings` composes smaller helpers
Calls `GetSlice` first (returns `[]interface{}`), then type-asserts each
element to `string`. Same "small helpers building bigger helpers" pattern
as `NewSession` → `NewVPN` + `SetClient` + `Cfg` in vpn.go.

## Why this file exists at all
Instead of every caller repeating:
```go
v, ok := cfg["port"]
if !ok { return err }
s, ok := v.(string)
if !ok { return err }
p, err := strconv.ParseUint(s, 10, 16)
...
```
You get one line: `port, err := config.GetPort(cfg, "port")`.
Classic DRY (don't repeat yourself) refactor into a shared helper package.
