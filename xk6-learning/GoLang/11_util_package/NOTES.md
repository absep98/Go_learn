# util/ package — http.go + safemap.go

Real code: `xk6-wss/internal/pkg/ngs-vpn-lib/util/`
These two files are used EVERYWHERE else — reading them once explains every
`SendHTTPRequest`, `DumpHttp`, and `SafeMap` call in auth.go / athena.go / vpn.go.

## http.go — the shared HTTP helper

- `SendHTTPRequest(...)` is the one function every auth/vpn call actually uses.
  Steps: build request → combine headers → attach cookies → build client → `client.Do(req)` (sends it).
- **The hidden `CheckRedirect`:** classic.go never writes `CheckRedirect`, yet its
  `resp.StatusCode == 302` check works. Why? Because `SendHTTPRequest` builds the
  `http.Client` internally WITH `CheckRedirect: return http.ErrUseLastResponse`,
  which means "don't auto-follow redirects, hand me the raw 302." Without it, Go's
  client would silently follow the redirect and you'd never see the 302 or its cookie.
  Lesson: behavior can come from a shared helper, not the file you're reading.
- `CheckRedirect func(req *http.Request, via []*http.Request) error` — 2 args:
  `req` = the next hop about to happen, `via` = breadcrumb trail of everywhere
  you've already been (grows +1 per redirect). This code ignores both (unnamed
  params) because its policy is a blanket "never follow." You'd use `via` for rules
  like "stop after 5 redirects" or "don't leave this domain."
- `DumpHttp(obj interface{})` — uses the SAME type-switch from 07_interfaces, but on
  real code. Takes `interface{}` so ONE function debugs both *http.Request and
  *http.Response. `switch obj.(type)` figures out which it got.

## safemap.go — thread-safe expiring map

- **`RWMutex`** (new): allows MANY simultaneous readers OR one exclusive writer.
  `RLock()`/`RUnlock()` = reading; `Lock()`/`Unlock()` = writing. Chosen here because
  the map is read constantly (every connection lookup) but written rarely — reads
  run in parallel instead of serializing. Only `Get` uses RLock; all others use Lock.
- **Callback via `cleanupFn func(interface{})`** (the key new idea): SafeMap stores
  `interface{}` (anything) so it can't know HOW to clean up a value on removal —
  TCP conn needs `.Close()`, a string needs nothing. So the CALLER hands it a cleanup
  function at creation; SafeMap just calls `m.cleanupFn(value)` when removing.
  "I handle the WHEN, you define the WHAT." Same shape as CheckRedirect (function
  stored as data). `noOpCleanup` is the do-nothing fallback if caller passes nil.
- **`unsafeRemove`** — the `unsafe` prefix is a Go naming convention meaning "this
  does NOT lock; the caller must already hold the lock." Callers (`Remove`,
  `CleanupLoop`) lock first, then call it. Prevents a double-lock deadlock (a
  goroutine trying to Lock() a mutex it already holds freezes forever).

## Two meanings of "interface" in Go (cleared up a confusion)
- `interface{}` (empty, no methods) = "ANY value at all" — string, int, conn, anything.
- `Auth interface { DoAuth() }` (named, has methods) = a CONTRACT — only types with
  those methods qualify.
Same keyword, totally different jobs. `cleanupFn func(interface{})` uses the empty
"any value" meaning, NOT the contract meaning.
