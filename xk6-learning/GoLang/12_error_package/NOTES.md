# error/ package — custom error type + the lerr.Errorf pattern

Real code: `xk6-wss/internal/pkg/ngs-vpn-lib/error/`
Decodes the `lerr.Errorf(...)` seen in EVERY auth/vpn function.

## The big payoff: Go's `error` is just an interface
```go
type error interface {
    Error() string
}
```
That's the ENTIRE definition of Go's built-in `error`. Anything with an
`Error() string` method IS an error — same interface rule as `Auth { DoAuth() }`.

This custom type satisfies it:
```go
type Error struct {
    RootCause RootCauseT
    ErrorText string
}
func (err *Error) Error() string {
    return fmt.Sprintf("[Root cause: %s] %s", err.RootCause, err.ErrorText)
}
```
That single method is why `return nil, lerr.Errorf(...)` works anywhere Go
expects an `error`. Not magic — the interface rule applied to Go's most
fundamental interface. Bonus: it carries a `RootCause` tag so every error
says WHY it happened (CgiLoginError, StoreDiscoveryError, ...).

## Errorf — variadic + type switch
```go
func Errorf(rc RootCauseT, t interface{}, args ...interface{}) *Error {
    switch t := t.(type) {
    case error:  return NewError(rc, fmt.Sprintf("%s", t))   // wrap an existing error
    case string: return NewError(rc, fmt.Sprintf(t, args...)) // format a message
    default:     panic("Invalid argument to Errorf")
    }
}
```
- `args ...interface{}` = VARIADIC — "zero or more args." Lets you call
  `Errorf(rc, err)` OR `Errorf(rc, "Status %d", 404)`. Inside, `args` is a slice.
- Concrete: `Errorf(rc, "Unexpected Status %d", 404)` → `t` = the string,
  `args` = `[404]` → `fmt.Sprintf(t, args...)` → "Unexpected Status 404".
- `args...` at a CALL site (with dots) = "spread this slice back into individual
  args" — the mirror of `...` in the parameter that gathered them up.

## The other two files
`auth.go` + `vpn.go` in this folder are JUST lists of RootCauseT string
constants (StoreDiscoveryError, CgiLoginError, etc.). Nothing new.

## Small extra: `type RootCauseT = string`
The `=` makes it a type ALIAS — RootCauseT is literally just another name for
string (fully interchangeable). Different from `type Duration int64` (a distinct
NAMED type that is NOT interchangeable with int64). Alias = nickname; named type = new type.
