package config

import "testing"

func TestLoad_ReadsAdminToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ADMIN_TOKEN", "operator-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AdminToken != "operator-token" {
		t.Fatalf("expected AdminToken from ADMIN_TOKEN, got %q", cfg.AdminToken)
	}
}

func TestLoad_AdminTokenDefaultsToEmpty(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ADMIN_TOKEN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AdminToken != "" {
		t.Fatalf("expected empty AdminToken (admin endpoints disabled), got %q", cfg.AdminToken)
	}
}
