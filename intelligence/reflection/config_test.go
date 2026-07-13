package reflection

import (
	"testing"
)

func TestDefaultConfig_Valid(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected DefaultConfig to be valid, got: %v", err)
	}
}

func TestConfigOptions_AndValidationFailures(t *testing.T) {
	cfg := DefaultConfig()
	WithEnabled(false)(&cfg)
	WithMaxConcurrentReflections(8)(&cfg)
	WithDefaultReflectionConfidence(0.90)(&cfg)

	if cfg.Enabled {
		t.Error("expected Enabled=false")
	}
	if cfg.MaxConcurrentReflections != 8 {
		t.Errorf("got MaxConcurrentReflections=%d, want 8", cfg.MaxConcurrentReflections)
	}
	if cfg.DefaultReflectionConfidence != 0.90 {
		t.Errorf("got DefaultReflectionConfidence=%f, want 0.90", cfg.DefaultReflectionConfidence)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validation failure: %v", err)
	}

	// Invalid confidence
	badConf := cfg
	WithDefaultReflectionConfidence(2.0)(&badConf)
	if err := badConf.Validate(); err == nil {
		t.Error("expected validation failure on out-of-bounds confidence")
	}

	// Invalid max concurrent
	badConcurrency := cfg
	WithMaxConcurrentReflections(0)(&badConcurrency)
	if err := badConcurrency.Validate(); err == nil {
		t.Error("expected validation failure on non-positive max concurrent reflections")
	}
}
