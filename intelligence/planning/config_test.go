package planning

import (
	"testing"
)

func TestDefaultConfig_Valid(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected DefaultConfig to be valid, got: %v", err)
	}

	if cfg.DefaultProfile == nil {
		t.Fatal("expected DefaultProfile in DefaultConfig, got nil")
	}
	if cfg.DefaultProfile.ProfileID != "PROFILE_GENERAL_BASE" {
		t.Errorf("unexpected DefaultProfile ID: %s", cfg.DefaultProfile.ProfileID)
	}
}

func TestConfigValidation_Invalid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxTraceRetention = -1
	if err := cfg.Validate(); err == nil {
		t.Error("expected error on negative MaxTraceRetention, got nil")
	}

	cfg = DefaultConfig()
	cfg.DefaultProfile = nil
	if err := cfg.Validate(); err == nil {
		t.Error("expected error on nil DefaultProfile, got nil")
	}
}

func TestPlanningPolicyProfile_Validation(t *testing.T) {
	profile := DefaultPlanningPolicyProfile()
	if err := profile.Validate(); err != nil {
		t.Fatalf("expected DefaultPlanningPolicyProfile to be valid, got: %v", err)
	}

	profile.MaxAlternatives = 0
	if err := profile.Validate(); err == nil {
		t.Error("expected error when MaxAlternatives is 0, got nil")
	}

	profile = DefaultPlanningPolicyProfile()
	profile.SearchStrategies["TACTICAL"].BeamWidth = 0
	if err := profile.Validate(); err == nil {
		t.Error("expected error when SearchStrategy BeamWidth is 0, got nil")
	}
}

func TestPlanningSearchStrategy_Validation(t *testing.T) {
	strat := &PlanningSearchStrategy{
		SearchID:               "strat-1",
		SearchType:             "HTN",
		MaxDepth:               10,
		MaxNodes:               100,
		BeamWidth:              3,
		AllowBacktracking:      true,
		AllowParallelExpansion: true,
	}
	if err := strat.Validate(); err != nil {
		t.Fatalf("expected valid strategy, got: %v", err)
	}

	strat.SearchID = ""
	if err := strat.Validate(); err == nil {
		t.Error("expected error when SearchID is empty, got nil")
	}
}
