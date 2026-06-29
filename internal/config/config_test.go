package config

import "testing"

func TestDefaultHasMVPProfiles(t *testing.T) {
	cfg := Default()
	if cfg.Provider != "e2b" {
		t.Fatalf("provider = %q, want e2b", cfg.Provider)
	}
	if cfg.SoftwareProfile("claude").Command != "claude" {
		t.Fatalf("missing claude profile")
	}
	if cfg.SoftwareProfile("aider").Command != "aider" {
		t.Fatalf("generic fallback should use software name")
	}
}
