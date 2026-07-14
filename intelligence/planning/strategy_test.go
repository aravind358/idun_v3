package planning

import (
	"testing"
	"time"
)

func TestStrategyProvider_AtomicLifecycle(t *testing.T) {
	prov := NewDefaultStrategyProvider(nil)
	snap := prov.ActiveSnapshot()
	if snap == nil || snap.ActiveProfile() == nil {
		t.Fatal("expected non-nil default snapshot and profile upon nil initialization")
	}
	if snap.ActiveProfile().ProfileID != "PROFILE_GENERAL_BASE" {
		t.Errorf("unexpected profile ID: %s", snap.ActiveProfile().ProfileID)
	}

	newProfile := *DefaultPlanningPolicyProfile()
	newProfile.ProfileID = "PROFILE_HIGH_PERF"
	newProfile.ProfileVersion = "2.1"
	newSnap, err := NewPlanningStrategySnapshot("snap-21", "2.1", &newProfile)
	if err != nil {
		t.Fatalf("failed to create new snapshot: %v", err)
	}

	if err := prov.UpdateSnapshot(newSnap); err != nil {
		t.Fatalf("failed to update snapshot: %v", err)
	}
	if prov.ActiveSnapshot().ActiveProfile().ProfileID != "PROFILE_HIGH_PERF" {
		t.Errorf("expected updated profile ID, got %s", prov.ActiveSnapshot().ActiveProfile().ProfileID)
	}

	if err := prov.UpdateSnapshot(nil); err == nil {
		t.Error("expected error when updating to nil snapshot, got nil")
	}
}

func TestPlanFingerprinter_StructuralInvariance(t *testing.T) {
	fpGen := NewDefaultPlanFingerprinter()

	basePlan := &Plan{
		PlanID:        "plan-hash-1",
		SchemaVersion: SchemaVersion2_0_0,
		CreatedAt:     time.Now(),
		Domain:        "Coding",
		Goal:          "Build feature",
		Subgoals: []Subgoal{
			{SubgoalID: "sg-1", Title: "Analyze"},
		},
		EstimatedCost:     10.0,
		EstimatedDuration: 1 * time.Hour,
		Status:            PlanStatusComplete,
	}

	fp1, err := fpGen.ComputeFingerprint(basePlan)
	if err != nil || fp1 == "" {
		t.Fatalf("failed to compute fingerprint: %v", err)
	}

	// Modify non-structural estimates and status
	basePlan.EstimatedCost = 500.0
	basePlan.EstimatedDuration = 48 * time.Hour
	basePlan.Status = PlanStatusPartialBudgetExhausted
	basePlan.CreatedAt = time.Now().Add(10 * time.Hour)

	fp2, err := fpGen.ComputeFingerprint(basePlan)
	if err != nil {
		t.Fatalf("failed to compute second fingerprint: %v", err)
	}
	if fp1 != fp2 {
		t.Errorf("fingerprint changed after modifying only non-structural estimates! fp1=%s fp2=%s", fp1, fp2)
	}

	// Modify structural content (add dependency)
	basePlan.Dependencies = append(basePlan.Dependencies, DependencyEdge{
		EdgeID: "e-1", SourceNodeID: "sg-1", TargetNodeID: "sg-2",
	})
	fp3, err := fpGen.ComputeFingerprint(basePlan)
	if err != nil {
		t.Fatalf("failed to compute third fingerprint: %v", err)
	}
	if fp1 == fp3 {
		t.Errorf("fingerprint did NOT change after modifying structural graph dependencies! fp=%s", fp1)
	}
}
