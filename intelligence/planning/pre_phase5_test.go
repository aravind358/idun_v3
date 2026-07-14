package planning

import (
	"context"
	"testing"
	"time"
)

// TestCapabilityFingerprint_ImmutableReproducibility verifies that CapabilityFingerprint
// provides deterministic, immutable identification of engine capabilities across time.
func TestCapabilityFingerprint_ImmutableReproducibility(t *testing.T) {
	caps1 := DefaultPlanningCapabilities()
	fp1 := ComputeCapabilityFingerprint(caps1)
	if fp1 == "" {
		t.Fatal("expected non-empty CapabilityFingerprint for default capabilities")
	}
	if caps1.CapabilityFingerprint != fp1 {
		t.Fatalf("expected DefaultPlanningCapabilities to populate matching fingerprint %s, got %s", fp1, caps1.CapabilityFingerprint)
	}

	// Modifying engine boundaries must alter the fingerprint to prevent historical replay ambiguity
	caps2 := DefaultPlanningCapabilities()
	caps2.MaxParallelWorkers = 32 // Changed from default 16
	fp2 := ComputeCapabilityFingerprint(caps2)
	if fp1 == fp2 {
		t.Fatalf("expected different fingerprint after mutating MaxParallelWorkers: fp1=%s, fp2=%s", fp1, fp2)
	}

	// Validation must reject mismatched capability fingerprints
	caps2.CapabilityFingerprint = fp1 // Intentional mismatch
	if err := caps2.Validate(); err == nil {
		t.Fatal("expected validation error when CapabilityFingerprint does not match structural digest, got nil")
	}

	// Fixing the fingerprint must restore validation pass
	caps2.CapabilityFingerprint = fp2
	if err := caps2.Validate(); err != nil {
		t.Fatalf("expected valid capability profile with updated fingerprint, got error: %v", err)
	}
}

// TestPlanningTrace_CompleteProvenanceAndSpecialistUsage verifies that every planning episode
// embeds complete replay provenance and factual specialist usage telemetry inside PlanningTrace.
func TestPlanningTrace_CompleteProvenanceAndSpecialistUsage(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("STRATEGIC"))

	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	defer service.Close()

	if err := service.Start(); err != nil {
		t.Fatalf("failed to start planning service: %v", err)
	}

	req := &PlanningRequest{
		RequestID:          "req-prov-101",
		Goal:               "Deploy autonomous navigation mission",
		Domain:             "General",
		TargetDepth:        DepthTactical,
		MaxExecutionBudget: 500 * time.Millisecond,
		MinConfidenceFloor: 0.70,
	}

	res, err := service.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical execution failed: %v", err)
	}
	if len(res.Traces) == 0 || res.Traces[0] == nil {
		t.Fatal("expected non-nil PlanningTrace returned in result")
	}

	trace := res.Traces[0]

	// 1. Verify Complete Replay Provenance
	if trace.PolicyFingerprint == "" {
		t.Error("expected non-empty PolicyFingerprint in trace")
	}
	if trace.CapabilityFingerprint == "" {
		t.Error("expected non-empty CapabilityFingerprint in trace")
	}
	if trace.SearchStrategyID == "" {
		t.Error("expected non-empty SearchStrategyID in trace")
	}
	if trace.ReplayMetadata.StrategySnapshotID == "" {
		t.Error("expected populated StrategySnapshotID in trace ReplayMetadata")
	}
	if trace.ReplayMetadata.ReplayFidelity != "EXACT" {
		t.Errorf("expected ReplayFidelity EXACT, got: %s", trace.ReplayMetadata.ReplayFidelity)
	}

	// 2. Verify Bounded Specialist Usage Telemetry
	if len(trace.SpecialistUsage) == 0 {
		t.Fatal("expected populated SpecialistUsage slice in trace")
	}

	// Check that registered specialists (HTN, GOAP, TreeSearch) have recorded usage facts
	foundHTN := false
	for _, usage := range trace.SpecialistUsage {
		if usage.SpecialistID == "" {
			t.Error("found SpecialistUsage with empty SpecialistID")
		}
		if usage.ContributionScore < 0.0 || usage.ContributionScore > 1.0 {
			t.Errorf("SpecialistUsage %s ContributionScore out of bounds: %f", usage.SpecialistID, usage.ContributionScore)
		}
		if usage.Invoked && usage.SkipReason != SkipNone {
			t.Errorf("expected SkipNone for invoked specialist %s, got %s", usage.SpecialistID, usage.SkipReason)
		}
		if !usage.Invoked && usage.SkipReason == "" {
			t.Errorf("expected non-empty SkipReason for uninvoked specialist %s", usage.SpecialistID)
		}
		if usage.SpecialistID == "HTNDecompositionSpecialist" {
			foundHTN = true
			if !usage.Invoked {
				t.Error("expected HTNDecompositionSpecialist to be invoked during General tactical episode")
			}
			if usage.ExecutionTimeUs == 0 {
				t.Error("expected non-zero ExecutionTimeUs for invoked HTNDecompositionSpecialist")
			}
			if !usage.Success {
				t.Error("expected Success true for invoked HTNDecompositionSpecialist")
			}
			if usage.ContributionScore <= 0.0 {
				t.Errorf("expected positive ContributionScore for HTN when subgoals generated, got %f", usage.ContributionScore)
			}
		}
	}

	if !foundHTN {
		t.Errorf("expected to find HTNDecompositionSpecialist in trace SpecialistUsage, got: %+v", trace.SpecialistUsage)
	}

	// 3. Verify Trace Structural Validation
	if err := trace.Validate(); err != nil {
		t.Fatalf("trace structural validation failed: %v", err)
	}
}

// TestSpecialistSkipReason_ObservationalTelemetry verifies that uninvoked specialists
// accurately record observational skip reasons without mutation or quality evaluation.
func TestSpecialistSkipReason_ObservationalTelemetry(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("STRATEGIC"))

	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	defer service.Close()
	_ = service.Start()

	// Request with domain mismatch for TACTICAL HTN/GOAP if they don't support "Quantum"
	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-skip-1").
		WithGoal("Execute quantum simulation").
		WithDomain("Quantum").
		Build()

	res, err := service.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical execution failed: %v", err)
	}
	trace := res.Traces[0]

	for _, usage := range trace.SpecialistUsage {
		if !usage.Invoked {
			if usage.SkipReason != SkipDomainMismatch && usage.SkipReason != SkipCapabilityDisabled && usage.SkipReason != SkipHigherPrioritySpecialist && usage.SkipReason != SkipNoApplicableGoal {
				t.Errorf("unexpected skip reason for uninvoked %s: %s", usage.SpecialistID, usage.SkipReason)
			}
		}
	}
}
