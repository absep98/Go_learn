package main

import (
	"fmt"
	"time"
)

// Base holds shared fields — like Base struct in ngs-vpn-lib/auth/auth.go
type Base struct {
	FQDN    string
	Timeout time.Duration
}

func (b *Base) Describe() string {
	return fmt.Sprintf("connects to %s (timeout %v)", b.FQDN, b.Timeout)
}

// Athena EMBEDS Base — no field name in front of "Base".
// This promotes Base's fields AND methods onto Athena directly.
type Athena struct {
	Base // ← embedding, not a named field
	Username string
}

func main() {
	a := Athena{
		Base:     Base{FQDN: "auth.example.com", Timeout: 30 * time.Second},
		Username: "abhishek",
	}

	// Promoted field — access WITHOUT a.Base.FQDN
	fmt.Println(a.FQDN) // works directly

	// Promoted method — access WITHOUT a.Base.Describe()
	fmt.Println(a.Describe()) // works directly

	// The long form still works too — embedding doesn't remove it
	fmt.Println(a.Base.FQDN)
}

// THE RULE:
//   type Athena struct { Base }        → EMBEDDING (no field name) → fields/methods promoted
//   type Athena struct { b Base }      → NORMAL FIELD (has a name) → need a.b.FQDN
//
// This is Go's replacement for class inheritance — composition, not "is-a".
// Real example: xk6-wss/internal/pkg/ngs-vpn-lib/auth/auth.go
//   type Base struct { authFQDN, vpnFQDN, timeout, transport ... }
//   type Athena struct { Base; PrimaryToken, ServiceToken, Username ... }
// Athena.vpnFQDN works directly even though vpnFQDN is declared inside Base.
