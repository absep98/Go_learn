package handlers

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
)

// AdminOnly protects operator endpoints such as /failed-jobs, whose dead-letter
// payloads contain users' journal text and were previously public.
//
// The caller must send the configured token in the X-Admin-Token header.
//   - token == "" → 404: the endpoint is disabled unless an operator opts in (fail closed)
//   - header missing or wrong → 401
//
// subtle.ConstantTimeCompare keeps response timing independent of how many
// leading characters of a guess were right.
//
// Trade-off: a single shared static token, not per-operator identities or roles.
// Enough for one operator; a team would need real admin users (RBAC) and audit logs.
func AdminOnly(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			http.NotFound(w, r)
			return
		}

		got := r.Header.Get("X-Admin-Token")
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			slog.Warn("Admin endpoint access denied", "path", r.URL.Path, "remote_addr", r.RemoteAddr)
			errorResponse(w, http.StatusUnauthorized, "Admin token required")
			return
		}

		next(w, r)
	}
}
