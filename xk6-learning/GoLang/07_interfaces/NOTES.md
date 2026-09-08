# Interfaces, Type Assertion, Pointer Receivers

Real code studied: `xk6-wss/internal/pkg/ngs-vpn-lib/auth/auth.go` + `classic.go`

## 1. Interface = contract, no registration needed

```go
type Auth interface {
    DoAuth() (*http.Cookie, *vpn.Session, *vpn.CfgResponse, error)
}
```
Any type with a matching `DoAuth()` method automatically satisfies `Auth`.
`ClassicAuth` and `Athena` both do this independently — no shared base class.

## 2. Method receiver — attaches a function to a type

```go
func (a *ClassicAuth) DoAuth() (...) { ... }
      └──────┬──────┘
           receiver: "this method belongs to *ClassicAuth"
```
`a` = the specific instance the method was called on (like `self`/`this`).

## 3. Pointer receiver vs value receiver — method SET differs

```go
func (c *Circle) Area() float64 { ... }   // pointer receiver
```
- `c.Area()` on a variable → works (Go auto-takes `&c`)
- `var s Shape = c` (value) → COMPILE ERROR — Circle's method set does NOT include pointer-receiver methods
- `var s Shape = &c` (pointer) → works

**Rule:** pointer receiver → only the pointer type satisfies the interface.

## 4. Type switch vs type assertion

```go
switch authIf.(type) {          // DISCOVER: what concrete type is inside?
case *auth.ClassicAuth:
    authIf.(*auth.ClassicAuth).DoAuth()   // EXTRACT: unwrap to use it
}
```
- `.(type)` — only valid inside `switch`, asks "what is it?"
- `.(ConcreteType)` — unwraps the interface to the real type so you can call its specific methods
- comma-ok form `v, ok := x.(string)` — never panics, `ok=false` + zero value if wrong type

## 5. Code-reading strategy (for real unfamiliar files)

Don't read top-to-bottom. Hunt in this order:
1. Every `type X struct` / `type X interface` — the nouns
2. Every `func (x *X) Method()` — what each noun can do
3. THEN read logic inside, function by function

## 6. Pattern spotted: thin method calls plain helper function

```go
func (a *ClassicAuth) DoAuth() (...) {
    authCookie, err := DoClassicAuth(a.VpnFQDN, a.Username, a.Password) // plain func, no receiver
    ...
}
```
The interface method is a thin wrapper; the real logic lives in a standalone function.
This is common — keeps the interface-facing method simple and testable.
