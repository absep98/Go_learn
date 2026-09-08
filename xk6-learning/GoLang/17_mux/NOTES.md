# vpn/mux.go — custom binary protocol for VPN packet multiplexing

Real code: `xk6-wss/internal/pkg/ngs-vpn-lib/vpn/mux.go` (79 lines)

## What "mux" means here
Multiplex = multiple traffic types (UDP, ICMP, DNS) traveling through ONE VPN
tunnel. To tell them apart at the other end, each packet gets a small header
stamped on it. `MuxMessage` = that header + payload.

## The 20-byte header layout
```
byte 0     : Version (1B)
byte 1     : Type    (1B — UDP=0x12, ICMP=0x13, DNS=0x15)
bytes 2-3  : Length  (2B)
bytes 4-7  : Destination IP (4B, IPv4)
bytes 8-11 : Source IP      (4B)
bytes 12-13: Destination port (2B)
bytes 14-15: Source port     (2B)
bytes 16-19: unused
+ Payload  : variable-length actual data appended after
```
Mirror-image pair: `Serialize` (struct → bytes) and `Deserialize` (bytes → struct).
Same shape idea as dns.go's query/response pair.

## THE key concept from this file — writing into slice ranges

**You CAN assign to a single slice index:**
```go
b[4] = 42       // ✅ works
```

**You CANNOT assign to a slice RANGE:**
```go
b[4:8] = someSlice   // ❌ compile error — not valid Go syntax
```

**To write into a range, use `copy(dst, src)`:**
```go
copy(b[4:8], m.DstIP.To4())   // writes 4 bytes into positions 4,5,6,7
```
`copy()` is a built-in function that fills the destination window with the
source bytes. The `b[4:8]` slice expression creates the "window"; copy fills it.

## Two other new tiny things
- `binary.BigEndian.PutUint16(b[2:4], value)` — packs a 2-byte integer into
  positions 2-3. Big-endian = most-significant byte first = the network standard.
- `append(b[:], m.Payload...)` — concatenates two slices. The `...` spreads
  the second slice into individual bytes (same `...` as `Errorf("Status %d", args...)`).

## Why Deserialize does TWO reads instead of one
- 1st read: fixed 20 bytes for the header (`b := make([]byte, muxHeaderLen)`)
- Header contains a `Length` field telling how big the payload is
- 2nd read: exactly that many bytes for the payload (`p := make([]byte, m.Length)`)

You can't read the payload without first knowing how big it is — and the
size is inside the header. Classic "read the length prefix, then read that
many more bytes" pattern used in almost every binary protocol on Earth.
