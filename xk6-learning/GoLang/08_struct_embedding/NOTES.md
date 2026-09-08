# Struct Embedding + Reading a Complex Real File (athena.go)

Real code studied: `xk6-wss/internal/pkg/ngs-vpn-lib/auth/athena.go` (~520 lines)

## 1. Struct embedding = Go's alternative to inheritance

```go
type Athena struct {
    Base              // embedded — no field name
    PrimaryToken string
}
```
- `Base` has no name in front → its fields AND methods get **promoted** onto `Athena`
- `a.vpnFQDN` works directly, even though `vpnFQDN` is physically declared inside `Base`
- Long form `a.Base.vpnFQDN` still works too — embedding adds a shortcut, doesn't remove access
- Contrast: `type Athena struct { b Base }` (named field) → MUST write `a.b.vpnFQDN`, no promotion

**Why Go does this instead of class inheritance:** composition over "is-a" hierarchies.
`Athena` HAS-A `Base` (shares its fields), it doesn't pretend to BE a `Base`.

## 2. How a big/scary file (500+ lines) actually breaks down

`athena.go` looks huge, but the `DoAuth()` method (the ONE method required by the `Auth`
interface) is only 4 steps:

```go
func (a *Athena) DoAuth() (...) {
    a.storeDiscovery()        // step 1: discover the auth domain/endpoint
    a.GetPrimaryToken(...)    // step 2: OIDC-style token handshake (4 sub-steps)
    a.CgiLogin()               // step 3: actual login, get session cookie
    vpn.NewSession(...)        // step 4: wrap cookie into a VPN session
}
```

Everything else in the file (`storeDiscovery`, `authEndpointsV1`, `domainProtocols`,
`explicitFormsStart`, `explicitForms`, `GetServiceToken`, `CgiLogin`) is just the
**implementation detail** behind those 4 steps — each one does an HTTP call, parses
an XML or cookie response, and feeds the next step.

**Lesson: a big file is not necessarily a hard file.** Find the ONE entry point method
(here: `DoAuth`, since that's what the `Auth` interface requires), trace what it calls
in order, and treat everything else as supporting detail you fill in only when needed.

## 3. Classic vs Athena — same interface, wildly different implementation

| | ClassicAuth | Athena |
|---|---|---|
| HTTP calls | 1 (login POST) | 4-6, chained (discovery → token → login) |
| Auth style | username/password form POST | OIDC-style token handshake |
| Both satisfy `Auth` interface | ✅ `DoAuth()` method | ✅ `DoAuth()` method |

This is the payoff of interfaces: `wss.go`'s `Login()` function calls `authIf.DoAuth()`
without caring which of these two wildly different flows is actually running underneath.
