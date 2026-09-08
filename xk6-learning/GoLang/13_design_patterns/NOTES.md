# Design Patterns (learned from real wss-tools code)

Design patterns = reusable, named solutions to problems that show up again and
again. Knowing the names lets you SPOT them in code and DESCRIBE your own designs
fast. This folder grows as we hit each pattern in the real codebase.

---

## 01_factory — Factory Pattern

**One-line:** a function that decides WHICH concrete type to build and hands it
back behind a common interface, so the caller doesn't need construction logic.

**Real code it came from:** `auth/configuration.go` → `FromConfiguration()`
```go
switch typeStr {
case TYPE_ATHENA:  return athenaFromConfiguration(...)   // builds *Athena
case TYPE_CLASSIC: return classicAuthFromConfiguration(...) // builds *ClassicAuth
}
```
Caller (`Login()` in wss.go) gets back "an Auth" and just calls `.DoAuth()` —
never knowing or caring which concrete type came out.

**The two ingredients (they work as a pair):**
- FACTORY decides *which* type to create (the `switch`)
- INTERFACE lets the caller *use* it without knowing which (`DoAuth()` / `Send()`)

**Why it's worth it — the payoff:** adding a NEW type later means:
- add one `case` to the factory switch
- add the new type's file
- ...and NOTHING else changes — not the caller, not the other types.
This "add without breaking existing code" property is the whole point.

**Coffee-machine analogy:** you press "latte" or "espresso" — you don't assemble
the drink. Hand the machine one word, it builds the right thing and hands it back.

**Runnable demo:** `01_factory/factory.go` (Notifier: email vs sms) — verified
running. Same `Send()` call works on both concrete types.

---

## Patterns seen but not yet in their own folder
- **Callback / strategy-ish**: `SafeMap.cleanupFn func(interface{})` and
  `http.Client.CheckRedirect` — inject behavior by passing a function. (see 11_util_package)
- **Constructor function**: `NewClassicAuth`, `NewVPN`, `NewSafeMap` — Go's
  stand-in for constructors (plain `New<Type>()` returning a pointer). (see 07, 08)
