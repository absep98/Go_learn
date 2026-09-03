# xk6-wss Repo Coverage Map

Goal: full depth on this codebase, tracked so nothing gets lost between sessions.
Strategy: alternate between reading a file (hunt structs/methods first) and
building a small real change — building cements understanding faster than
reading alone (proven by the -timeout exercise).

## utils/ (standalone test-server binaries)
- [x] `echo-server.go` — DONE, fully read + modified (-timeout flag, -json-stats flag) + tested live
- [ ] `udp-echo-server.go` — not started
- [ ] `web-traffic-echo-server.go` — not started
- [ ] `spa-mock-server.go` — not started

## xk6-wss/internal/pkg/ngs-vpn-lib/auth/ — authentication
- [x] `auth.go` — DONE (Auth interface, Base struct)
- [x] `classic.go` — DONE (ClassicAuth: single HTTP POST login)
- [x] `athena.go` — DONE (Athena: multi-step OIDC-style handshake, struct embedding)
- [x] `configuration.go` — DONE (FromConfiguration factory: switch on "type" → athena/classic; factory pattern documented in 13_design_patterns)

## xk6-wss/internal/pkg/ngs-vpn-lib/vpn/ — core tunnel session
- [~] `vpn.go` — PARTIAL: Session struct, NewVPN, NewSession, SetClient, Cfg read.
      17 methods total, ~13 unread (CsTCP, MuxUDP, MuxDNS, HTTP, SecureBrowse, etc.)
- [x] `mux.go` — DONE (custom binary protocol: MuxMessage 20-byte header + payload; Serialize/Deserialize pair; slice range assignment needs `copy(dst, src)`; length-prefixed read pattern)

## xk6-wss/internal/pkg/ngs-vpn-lib/protocol/ — wire protocols
- [x] `dns/dns.go` — DONE (DNSQueryA builds packet, DNSResourceA parses response; nested struct literals; `r.A[:]` array-to-slice; type-assertion-with-comma-ok reinforced)
- [x] `tcp/tcp.go` + `tcp/configuration.go` — DONE (struct + constructor + mini-factory + config helpers; no new Go concepts)
- [x] `udp/udp.go` + `udp/configuration.go` — DONE (same shape as tcp but carries DNS bits since UDP tunnels DNS too)

## xk6-wss/internal/pkg/ngs-vpn-lib/error/ — error types
- [x] `error.go` — DONE (custom Error type satisfies Go's `error` interface via Error() string; Errorf variadic + type switch; type alias vs named type)
- [x] `auth.go` — DONE (just RootCauseT constants: StoreDiscoveryError, CgiLoginError, etc.)
- [x] `vpn.go` — DONE (same: VPN-side RootCauseT constants)

## xk6-wss/internal/pkg/ngs-vpn-lib/util/ — shared helpers
- [x] `http.go` — DONE (SendHTTPRequest steps, hidden CheckRedirect, DumpHttp type-switch)
- [x] `safemap.go` — DONE (RWMutex read/write split, cleanupFn callback, unsafeRemove convention, empty vs named interface)

## xk6-wss/internal/pkg/phobos-lib/pacer/ — load-test rate limiting
- [x] `pacer.go` — DONE (Pacer interface, 3-return contract, FromConfiguration factory)
- [x] `constant.go` — DONE (constant-rate pacer; defer-placement-as-logic; self-correcting sleep=0 when behind schedule)

## xk6-wss/pkg/configuration/ — top-level config parsing
- [x] `configuration.go` — DONE (typed getters via comma-ok + type-assert; GetInt uses type-switch to accept int/int64/float64 from JSON; GetSliceOfStrings composes smaller helpers)

## xk6-wss/wss.go — the big k6 module file (~3250 lines)
- [~] PARTIAL: read init(), RootModule/ModuleInstance, registerMetrics, Login(),
      SendDNSRequest, AccessUdpApp, AccessTcpApp, DirectTcpAccess, TLS transport setup
- [ ] Unread: CepGwAccess and other CEP-specific gateway functions (large chunk,
      likely 800+ lines — this is the newest/most actively developed part per the
      code comments about CEP/Hybrid deployment)

## Suggested order for future sessions
1. `util/http.go` + `util/safemap.go` — small, and used EVERYWHERE else (read once, understand every call site after)
2. `error/*.go` — tiny, quick win, explains the `lerr.Errorf(...)` pattern seen constantly
3. `auth/configuration.go` — closes the auth/ package fully
4. Remaining `vpn.go` methods — CsTCP, MuxUDP, HTTP, SecureBrowse
5. `protocol/tcp`, `protocol/udp`, `protocol/dns` — the wire-format details
6. `mux.go`, `pacer/*.go`
7. `wss.go` CepGwAccess section — the big one, save for last once everything it depends on is understood
