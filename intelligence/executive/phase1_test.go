package executive_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"idun/intelligence/calibration"
	"idun/intelligence/constitution"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

func TestExecutivePolicyProfileBuilderAndValidation(t *testing.T) {
	t.Run("valid default profile builds cleanly", func(t *testing.T) {
		profile, err := executive.NewExecutivePolicyProfileBuilder().
			WithProfileID("test-profile").
			WithVersion("1.0.0").
			WithSource("Learning").
			Build()
		if err != nil {
			t.Fatalf("unexpected error building default profile: %v", err)
		}
		if profile.SchemaVersion != executive.SchemaVersion {
			t.Errorf("expected schema version %q, got %q", executive.SchemaVersion, profile.SchemaVersion)
		}
		if profile.PolicyFingerprint == "" {
			t.Error("expected non-empty PolicyFingerprint")
		}
	})

	t.Run("validation firewall rejects invalid schema version", func(t *testing.T) {
		profile := executive.DefaultExecutivePolicyProfile()
		profile.SchemaVersion = "1.0-INVALID"
		if err := profile.Validate(); !errors.Is(err, executive.ErrInvalidPolicy) {
			t.Errorf("expected ErrInvalidPolicy for wrong schema version, got: %v", err)
		}
	})

	t.Run("validation firewall rejects negative budget units", func(t *testing.T) {
		profile := executive.DefaultExecutivePolicyProfile()
		profile.BudgetPolicies.MaxCycleBudgetUnits = -100
		if err := profile.Validate(); !errors.Is(err, executive.ErrInvalidPolicy) {
			t.Errorf("expected ErrInvalidPolicy for negative budget units, got: %v", err)
		}
	})

	t.Run("validation firewall rejects out-of-bounds admission threshold", func(t *testing.T) {
		profile := executive.DefaultExecutivePolicyProfile()
		profile.WorkspacePolicies.AdmissionThreshold = 1.5
		if err := profile.Validate(); !errors.Is(err, executive.ErrInvalidPolicy) {
			t.Errorf("expected ErrInvalidPolicy for admission threshold > 1.0, got: %v", err)
		}
	})

	t.Run("fingerprint tampering detected by validation firewall", func(t *testing.T) {
		profile := executive.DefaultExecutivePolicyProfile()
		profile.PolicyFingerprint = "deadbeefdeadbeef"
		if err := profile.Validate(); !errors.Is(err, executive.ErrInvalidPolicy) {
			t.Errorf("expected ErrInvalidPolicy for fingerprint mismatch, got: %v", err)
		}
	})
}

func TestExecutiveCapabilitiesBuilderAndValidation(t *testing.T) {
	t.Run("valid default capabilities build cleanly", func(t *testing.T) {
		caps, err := executive.NewExecutiveCapabilitiesBuilder().
			WithMaxConcurrentEpisodes(50).
			WithMaxRetryBudget(10).
			Build()
		if err != nil {
			t.Fatalf("unexpected error building capabilities: %v", err)
		}
		if caps.CapabilityFingerprint == "" {
			t.Error("expected non-empty CapabilityFingerprint")
		}
	})

	t.Run("validation firewall rejects negative max retry budget", func(t *testing.T) {
		caps := executive.DefaultExecutiveCapabilities()
		caps.MaxRetryBudget = -1
		if err := caps.Validate(); !errors.Is(err, executive.ErrInvalidCapabilities) {
			t.Errorf("expected ErrInvalidCapabilities for negative max retry budget, got: %v", err)
		}
	})

	t.Run("fingerprint mismatch detected by validation firewall", func(t *testing.T) {
		caps := executive.DefaultExecutiveCapabilities()
		caps.CapabilityFingerprint = "badfingerprint"
		if err := caps.Validate(); !errors.Is(err, executive.ErrInvalidCapabilities) {
			t.Errorf("expected ErrInvalidCapabilities for fingerprint mismatch, got: %v", err)
		}
	})
}

func TestExecutiveRequestBuilderAndValidation(t *testing.T) {
	t.Run("valid request build", func(t *testing.T) {
		req, err := executive.NewExecutiveRequestBuilder().
			WithRequestID("req-101").
			WithEpisodeID("ep-202").
			WithPriority(executive.PriorityBand1RealTime).
			WithBudget(executive.BudgetStandard).
			Build()
		if err != nil {
			t.Fatalf("unexpected error building request: %v", err)
		}
		if req.RequestID != "req-101" {
			t.Errorf("got request ID %q, expected req-101", req.RequestID)
		}
	})

	t.Run("request builder rejects missing episode id", func(t *testing.T) {
		_, err := executive.NewExecutiveRequestBuilder().
			WithRequestID("req-101").
			Build()
		if !errors.Is(err, executive.ErrInvalidResult) {
			t.Errorf("expected ErrInvalidResult for missing episode id, got: %v", err)
		}
	})
}

func TestExecutiveTraceAndResultValidation(t *testing.T) {
	replayMeta, err := executive.NewReplayMetadataBuilder().
		WithPolicyFingerprint("fp-policy").
		WithCapabilityFingerprint("fp-caps").
		WithReplaySeed(12345).
		WithExecutiveVersion(executive.DefaultExecutiveVersion).
		Build()
	if err != nil {
		t.Fatalf("unexpected error building replay metadata: %v", err)
	}

	t.Run("valid trace builds cleanly", func(t *testing.T) {
		trace, err := executive.NewExecutiveTraceBuilder().
			WithTraceID("trace-1").
			WithEpisodeID("ep-1").
			WithExecutionDuration(100 * time.Millisecond).
			WithProvenance("fp-policy", "fp-caps", executive.DefaultExecutiveVersion).
			WithReplayMetadata(*replayMeta).
			WithBudgetsAllocated([]executive.BudgetAllocationEvent{
				{WorkflowID: "wf-1", NodeID: "node-1", TierAssigned: executive.BudgetStandard, UnitsAllocated: 10, Timestamp: time.Now()},
			}).
			Build()
		if err != nil {
			t.Fatalf("unexpected error building trace: %v", err)
		}
		if len(trace.BudgetsAllocated) != 1 {
			t.Errorf("expected 1 budget event, got %d", len(trace.BudgetsAllocated))
		}
	})

	t.Run("trace validation firewall rejects missing trace_id", func(t *testing.T) {
		_, err := executive.NewExecutiveTraceBuilder().
			WithEpisodeID("ep-1").
			WithProvenance("fp-policy", "fp-caps", executive.DefaultExecutiveVersion).
			WithReplayMetadata(*replayMeta).
			Build()
		if !errors.Is(err, executive.ErrInvalidTrace) {
			t.Errorf("expected ErrInvalidTrace when trace_id is missing, got: %v", err)
		}
	})

	t.Run("valid result builds cleanly with coordination summary", func(t *testing.T) {
		summary := executive.ExecutiveCoordinationSummary{
			EpisodesCoordinated:         1,
			AverageCoordinationDuration: 50 * time.Millisecond,
			TotalCoordinationDuration:   50 * time.Millisecond,
			SuccessfulCoordinations:     1,
		}
		result, err := executive.NewExecutiveResultBuilder().
			WithEpisodeID("ep-1").
			WithWorkflowID("wf-1").
			WithStatus(executive.StatusSuccess).
			WithTerminationReason(executive.ReasonSuccess).
			WithOutputRef("output://resource-1").
			WithCoordinationSummary(summary).
			WithTraceID("trace-1").
			WithProvenance("fp-policy", "fp-caps", executive.DefaultExecutiveVersion).
			WithReplayMetadata(*replayMeta).
			Build()
		if err != nil {
			t.Fatalf("unexpected error building result: %v", err)
		}
		if result.CoordinationSummary.SuccessfulCoordinations != 1 {
			t.Errorf("expected 1 successful coordination, got %d", result.CoordinationSummary.SuccessfulCoordinations)
		}
	})

	t.Run("result validation rejects unknown status enum", func(t *testing.T) {
		summary := executive.ExecutiveCoordinationSummary{}
		_, err := executive.NewExecutiveResultBuilder().
			WithEpisodeID("ep-1").
			WithWorkflowID("wf-1").
			WithStatus(executive.ExecutiveResultStatus("UNKNOWN_STATUS")).
			WithTerminationReason(executive.ReasonSuccess).
			WithCoordinationSummary(summary).
			WithProvenance("fp-policy", "fp-caps", executive.DefaultExecutiveVersion).
			WithReplayMetadata(*replayMeta).
			Build()
		if !errors.Is(err, executive.ErrInvalidResult) {
			t.Errorf("expected ErrInvalidResult for unknown status enum, got: %v", err)
		}
	})
}

func TestConfigurationValidation(t *testing.T) {
	t.Run("default configuration passes validation", func(t *testing.T) {
		cfg := executive.DefaultConfiguration()
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected default configuration to be valid, got: %v", err)
		}
	})

	t.Run("configuration validation rejects nil policy", func(t *testing.T) {
		cfg := executive.DefaultConfiguration()
		cfg.Policy = nil
		if err := cfg.Validate(); !errors.Is(err, executive.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig when policy is nil, got: %v", err)
		}
	})
}

func TestSnapshotHoldersConcurrencyAndValidation(t *testing.T) {
	t.Run("policy snapshot holder stores and loads atomically under concurrency", func(t *testing.T) {
		holder, err := executive.NewPolicySnapshotHolder(executive.DefaultExecutivePolicyProfile())
		if err != nil {
			t.Fatalf("unexpected error creating holder: %v", err)
		}

		// Reject unvalidated/tampered profile store
		badProfile := executive.DefaultExecutivePolicyProfile()
		badProfile.PolicyFingerprint = "corrupt"
		if err := holder.Store(badProfile); err == nil {
			t.Fatal("expected Store to reject corrupt profile, got nil")
		}

		var wg sync.WaitGroup
		// 20 concurrent readers and writers
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				if idx%2 == 0 {
					p := holder.Load()
					if p == nil || p.SchemaVersion != executive.SchemaVersion {
						t.Errorf("concurrent load read invalid profile")
					}
				} else {
					newProfile, _ := executive.NewExecutivePolicyProfileBuilder().
						WithProfileID("profile-updated").
						WithVersion("1.0.1").
						WithSource("Learning").
						Build()
					_ = holder.Store(newProfile)
				}
			}(i)
		}
		wg.Wait()
	})

	t.Run("capabilities snapshot holder concurrency", func(t *testing.T) {
		holder, err := executive.NewCapabilitiesSnapshotHolder(executive.DefaultExecutiveCapabilities())
		if err != nil {
			t.Fatalf("unexpected error creating holder: %v", err)
		}

		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				if idx%2 == 0 {
					c := holder.Load()
					if c == nil || c.MaxConcurrentEpisodes < 1 {
						t.Errorf("concurrent load read invalid capabilities")
					}
				} else {
					newCaps, _ := executive.NewExecutiveCapabilitiesBuilder().
						WithMaxConcurrentEpisodes(200).
						Build()
					_ = holder.Store(newCaps)
				}
			}(i)
		}
		wg.Wait()
	})
}

func TestServiceV2Phase1Integration(t *testing.T) {
	ws := workspace.NewEngine()
	cal := calibration.NewService()
	constGate := constitution.NewGate()

	customPolicy, _ := executive.NewExecutivePolicyProfileBuilder().
		WithProfileID("custom-init-profile").
		Build()
	customCaps, _ := executive.NewExecutiveCapabilitiesBuilder().
		WithMaxConcurrentEpisodes(500).
		Build()

	svc, err := executive.NewServiceV2(ws, cal, constGate, 1000,
		executive.WithPolicy(customPolicy),
		executive.WithCapabilities(customCaps),
	)
	if err != nil {
		t.Fatalf("failed to construct ServiceV2: %v", err)
	}

	if svc.Policy().ProfileID != "custom-init-profile" {
		t.Errorf("expected profile ID custom-init-profile, got %q", svc.Policy().ProfileID)
	}
	if svc.Capabilities().MaxConcurrentEpisodes != 500 {
		t.Errorf("expected MaxConcurrentEpisodes 500, got %d", svc.Capabilities().MaxConcurrentEpisodes)
	}

	// Test atomic policy update from Learning
	newPolicy, _ := executive.NewExecutivePolicyProfileBuilder().
		WithProfileID("learning-published-profile").
		Build()
	if err := svc.UpdatePolicy(newPolicy); err != nil {
		t.Fatalf("unexpected error updating policy: %v", err)
	}
	if svc.Policy().ProfileID != "learning-published-profile" {
		t.Errorf("expected updated profile ID learning-published-profile, got %q", svc.Policy().ProfileID)
	}
}
