package reasoning

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultConfig_Valid(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultConfig failed validation: %v", err)
	}
	if cfg.MaxBeamWidth != MaxBeamWidth {
		t.Errorf("expected MaxBeamWidth %d, got %d", MaxBeamWidth, cfg.MaxBeamWidth)
	}
}

func TestConfigOptions_AndValidationFailures(t *testing.T) {
	cfg := DefaultConfig()

	WithTimeout(10 * time.Second)(&cfg)
	WithEscalationThreshold(0.60)(&cfg)
	WithGraphLimits(300, 1000, 2)(&cfg)

	if err := cfg.Validate(); err != nil {
		t.Fatalf("modified config failed validation: %v", err)
	}
	if cfg.DefaultTimeout != 10*time.Second || cfg.EscalationThreshold != 0.60 {
		t.Errorf("config options did not apply correctly: %+v", cfg)
	}

	badCfg := cfg
	badCfg.EscalationThreshold = 1.5
	if err := badCfg.Validate(); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig for bad escalation threshold, got %v", err)
	}
}
