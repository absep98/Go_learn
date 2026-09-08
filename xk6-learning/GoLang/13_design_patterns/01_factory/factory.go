package main

import "fmt"

// ── The interface: what all products share ──────────────────────────────
// (mirrors `Auth interface { DoAuth() }` in the real repo)
type Notifier interface {
	Send(msg string) string
}

// ── Concrete types: each builds/behaves differently ─────────────────────
type EmailNotifier struct{ address string }
type SMSNotifier struct{ phone string }

func (e *EmailNotifier) Send(msg string) string {
	return fmt.Sprintf("EMAIL to %s: %s", e.address, msg)
}
func (s *SMSNotifier) Send(msg string) string {
	return fmt.Sprintf("SMS to %s: %s", s.phone, msg)
}

// ── The FACTORY: decides which concrete type to build ───────────────────
// (mirrors FromConfiguration -> switch -> athena/classic in the real repo)
func NewNotifier(kind, target string) (Notifier, error) {
	switch kind {
	case "email":
		return &EmailNotifier{address: target}, nil
	case "sms":
		return &SMSNotifier{phone: target}, nil
	default:
		return nil, fmt.Errorf("unknown notifier type: %s", kind)
	}
}

func main() {
	// The caller hands the factory ONE string and gets back a Notifier.
	// It does NOT know or care which concrete type it received.
	for _, cfg := range []struct{ kind, target string }{
		{"email", "abhishek@example.com"},
		{"sms", "+91-99999-00000"},
	} {
		n, err := NewNotifier(cfg.kind, cfg.target)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		// Same call works for BOTH types — that's the interface hiding the difference.
		fmt.Println(n.Send("your build passed"))
	}

	// Adding a 3rd type (e.g. "slack") = add one case + one struct.
	// This main() would NOT change at all. That's the payoff.
}
