package handlers

/*
=== ADMIN-ONLY MIDDLEWARE TESTS ===

/failed-jobs returns dead-letter payloads, which contain users' journal text.
It was served to anyone, with no authentication. AdminOnly puts it behind an
operator token sent in the X-Admin-Token header:
- no token configured  → 404 (the endpoint does not exist; fail closed)
- header missing/wrong → 401
- correct token        → request reaches the handler
*/

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func adminProbe(t *testing.T, configured, sent string) (code int, reached bool) {
	t.Helper()
	next := func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}
	req := httptest.NewRequest(http.MethodGet, "/failed-jobs", nil)
	if sent != "" {
		req.Header.Set("X-Admin-Token", sent)
	}
	rec := httptest.NewRecorder()
	AdminOnly(configured, next)(rec, req)
	return rec.Code, reached
}

func TestAdminOnly_DisabledWhenNoTokenConfigured(t *testing.T) {
	code, reached := adminProbe(t, "", "anything")
	if code != http.StatusNotFound || reached {
		t.Fatalf("expected 404 and handler not reached, got %d reached=%v", code, reached)
	}
}

func TestAdminOnly_RejectsMissingToken(t *testing.T) {
	code, reached := adminProbe(t, "s3cret-operator-token", "")
	if code != http.StatusUnauthorized || reached {
		t.Fatalf("expected 401 and handler not reached, got %d reached=%v", code, reached)
	}
}

func TestAdminOnly_RejectsWrongToken(t *testing.T) {
	code, reached := adminProbe(t, "s3cret-operator-token", "s3cret-operator-tokeX")
	if code != http.StatusUnauthorized || reached {
		t.Fatalf("expected 401 and handler not reached, got %d reached=%v", code, reached)
	}
}

func TestAdminOnly_AllowsCorrectToken(t *testing.T) {
	code, reached := adminProbe(t, "s3cret-operator-token", "s3cret-operator-token")
	if code != http.StatusOK || !reached {
		t.Fatalf("expected 200 and handler reached, got %d reached=%v", code, reached)
	}
}
