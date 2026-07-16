package learning

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected default config to validate cleanly, got %v", err)
	}

	if cfg.ServiceVersion != SchemaVersion {
		t.Errorf("expected version %q, got %q", SchemaVersion, cfg.ServiceVersion)
	}
	if cfg.PolicyProfile.Author != "Executive" {
		t.Errorf("expected policy author Executive, got %q", cfg.PolicyProfile.Author)
	}
	if !cfg.Capabilities.SupportsOfflineLearning {
		t.Error("expected offline learning supported in default config")
	}
}

func TestConfigValidationErrors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxWorkers = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero workers")
	}

	cfg = DefaultConfig()
	cfg.ServiceVersion = "wrong.version"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid version string")
	}

	cfg = DefaultConfig()
	cfg.PolicyProfile = nil
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for nil policy profile")
	}
}
