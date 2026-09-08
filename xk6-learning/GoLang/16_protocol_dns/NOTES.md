# protocol/dns/dns.go — DNS query builder + response parser

Real code: `xk6-wss/internal/pkg/ngs-vpn-lib/protocol/dns/dns.go` (88 lines)
Completes the entire `protocol/` folder.

## Mirror-image pair
- `DNSQueryA(id, domain)` — BUILDS a DNS "what's the IP of X?" packet as `[]byte`
- `DNSResourceA(buf)` — PARSES a DNS response packet, extracts the IP

Uses the external `golang.org/x/net/dns/dnsmessage` package — you don't need
to memorize the DNS wire format; the library handles it.

## Nested struct literals (looks scary, isn't)
```go
message := dnsmessage.Message{
    Header: dnsmessage.Header{
        ID: id,
        RecursionDesired: true,
    },
    Questions: []dnsmessage.Question{
        { Name: name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET },
    },
}
```
Same `Struct{Field: value}` shape as `Stats{ActiveConns: ...}` — just nested.
An outer struct's field value happens to BE another struct literal. Peel one
layer at a time. The `Questions: []dnsmessage.Question{ {...} }` bit is a
slice literal holding one struct literal. All the same shape all the way down.

## `r.A[:]` — array-to-slice conversion (small but new)
`r.A` is `[4]byte` — a FIXED-SIZE array (a DNS A-record IP is always 4 bytes).
`r.A[:]` slices the whole array → `[]byte`. Needed because `net.IP()` accepts
a slice, not an array. Same "explicit conversion" spirit as `uint16(port)`.

**Array vs slice reminder:** `[4]byte` (fixed size, part of the type) is a
different type from `[]byte` (dynamic size). Not interchangeable — you must
slice with `[:]` to bridge them.

## Loop pattern — filter then extract
```go
for _, answer := range answers {
    if answer.Header.Type != dnsmessage.TypeA || answer.Header.Class != dnsmessage.ClassINET {
        continue    // skip non-A records — DNS can return CNAME, AAAA, MX, etc.
    }
    r, ok := answer.Body.(*dnsmessage.AResource)   // type assertion, comma-ok
    if !ok { continue }
    return net.IP(r.A[:]), nil
}
```
`continue` (not `return`) because DNS responses often contain multiple record
types — we want to skip past the ones we don't care about and keep scanning
until we find an A record.

The `r, ok := answer.Body.(*dnsmessage.AResource)` line is the
**type-assertion-with-comma-ok** pattern seen everywhere (`v, ok := If.(string)`
in configuration.go, etc.). "Try to unwrap this interface as SomeType; tell me
if it worked." Third+ time seeing it now — should feel automatic.
