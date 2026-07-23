package planning

import (
	"sync"
	"testing"
	"time"
)

func TestPlanValidation_Valid(t *testing.T) {
	plan := &CandidatePlan{
		PlanID:             "plan-101",
		SchemaVersion:      SchemaVersion2_0_0,
		CreatedAt:          time.Now().UTC(),
		StrategySnapshotID: "snap-200",
		PlanFingerprint:    "a1b2c3d4",
		Domain:             "Coding",
		Goal:               "Refactor database module",
		Subgoals: []Subgoal{
			{SubgoalID: "sg-1", Title: "Analyze schema", Description: "Inspect SQL tables"},
		},
		Dependencies: []DependencyEdge{
			{EdgeID: "e-1", SourceNodeID: "sg-1", TargetNodeID: "sg-2", DependencyType: "HARD_PREREQUISITE"},
		},
		EstimatedCost:     15.5,
		EstimatedDuration: 2 * time.Hour,
		ConfidenceProfile: ConfidenceProfile{
			GoalConfidence:         0.90,
			PreconditionConfidence: 0.85,
			DependencyConfidence:   0.88,
			ResourceConfidence:     0.82,
			TimingConfidence:       0.80,
			ConstraintConfidence:   0.90,
			OverallConfidence:      0.80, // min is 0.80
		},
		PlanStatus: PlanStatusComplete,
		ReplayMetadata: ReplayMetadata{
			StrategySnapshotID: "snap-200",
			ReplayFidelity:     "EXACT",
			ReplaySeed:         42,
		},
		TraceID: "trace-999",
	}

	if err := plan.Validate(); err != nil {
		t.Fatalf("expected valid plan, got error: %v", err)
	}

	// Test alias schema version
	plan.SchemaVersion = SchemaVersion2_0
	if err := plan.Validate(); err != nil {
		t.Fatalf("expected valid plan with schema alias, got error: %v", err)
	}
}

func TestPlanValidation_InvalidSchema(t *testing.T) {
	plan := &CandidatePlan{
		PlanID:             "plan-101",
		SchemaVersion:      "1.0.0-OLD",
		StrategySnapshotID: "snap-200",
		Goal:               "Test goal",
		TraceID:            "trace-999",
	}
	if err := plan.Validate(); err == nil {
		t.Fatal("expected error on invalid schema version, got nil")
	}
}

func TestConfidenceProfileValidation(t *testing.T) {
	tests := []struct {
		name    string
		cp      ConfidenceProfile
		wantErr bool
	}{
		{
			name: "valid_minimum_bound",
			cp: ConfidenceProfile{
				GoalConfidence:         0.9,
				PreconditionConfidence: 0.8,
				DependencyConfidence:   0.8,
				ResourceConfidence:     0.7,
				TimingConfidence:       0.8,
				ConstraintConfidence:   0.9,
				OverallConfidence:      0.7,
			},
			wantErr: false,
		},
		{
			name: "invalid_overall_exceeds_min",
			cp: ConfidenceProfile{
				GoalConfidence:         0.9,
				PreconditionConfidence: 0.8,
				DependencyConfidence:   0.8,
				ResourceConfidence:     0.5,
				TimingConfidence:       0.8,
				ConstraintConfidence:   0.9,
				OverallConfidence:      0.8, // Exceeds min dimension (0.5)
			},
			wantErr: true,
		},
		{
			name: "invalid_dimension_out_of_bounds",
			cp: ConfidenceProfile{
				GoalConfidence:         1.2,
				PreconditionConfidence: 0.8,
				DependencyConfidence:   0.8,
				ResourceConfidence:     0.7,
				TimingConfidence:       0.8,
				ConstraintConfidence:   0.9,
				OverallConfidence:      0.7,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cp.Validate()
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestPlanningTraceValidation(t *testing.T) {
	trace := &PlanningTrace{
		TraceID:             "trace-1",
		PlanID:              "plan-1",
		SchemaVersion:       SchemaVersion2_0_0,
		StrategySnapshotID:  "snap-1",
		TerminationReason:   TerminationGoalFound,
		EstimatedComplexity: 12.5,
		RejectedBranches: []RejectedBranch{
			{BranchID: "b-alt", DiscardReason: "ResourceConflict: GPU limit"},
		},
		ConfidenceProfile: ConfidenceProfile{
			GoalConfidence: 0.9, PreconditionConfidence: 0.9, DependencyConfidence: 0.9,
			ResourceConfidence: 0.9, TimingConfidence: 0.9, ConstraintConfidence: 0.9,
			OverallConfidence: 0.9,
		},
		QualityMetrics: QualityMetrics{
			Completeness: 0.95, Efficiency: 0.88, Robustness: 0.92,
		},
		SearchStatistics: SearchStatistics{
			NodesExpanded: 120, NodesPruned: 45, CacheHits: 3,
		},
	}

	if err := trace.Validate(); err != nil {
		t.Fatalf("expected valid trace, got: %v", err)
	}

	// Test missing termination reason
	trace.TerminationReason = ""
	if err := trace.Validate(); err == nil {
		t.Fatal("expected error on missing termination reason, got nil")
	}
}

func TestStrategySnapshotConcurrentSwap(t *testing.T) {
	baseProfile := DefaultPlanningPolicyProfile()
	snap, err := NewPlanningStrategySnapshot("snap-1", "1.0", baseProfile)
	if err != nil {
		t.Fatalf("failed to create snapshot: %v", err)
	}

	var wg sync.WaitGroup
	// Run concurrent readers and writers to verify atomic pointer safety under -race
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%5 == 0 {
				newP := *baseProfile
				newP.ProfileVersion = "2.0"
				_ = snap.SwapProfile(&newP)
			} else {
				p := snap.ActiveProfile()
				if p == nil || p.ProfileID == "" {
					t.Errorf("concurrent reader got nil or invalid profile")
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestSupportingStructValidations(t *testing.T) {
	// Subgoal
	sg := Subgoal{SubgoalID: "sg-1", Title: "Test"}
	if err := sg.Validate(); err != nil {
		t.Errorf("expected valid subgoal: %v", err)
	}
	sg.Title = ""
	if err := sg.Validate(); err == nil {
		t.Error("expected error for missing subgoal title")
	}

	// DependencyEdge
	dep := DependencyEdge{EdgeID: "e-1", SourceNodeID: "a", TargetNodeID: "b"}
	if err := dep.Validate(); err != nil {
		t.Errorf("expected valid edge: %v", err)
	}
	dep.TargetNodeID = "a" // Self-loop
	if err := dep.Validate(); err == nil {
		t.Error("expected error for self-loop dependency edge")
	}

	// ResourceRequirement
	res := ResourceRequirement{ResourceID: "res-1", ResourceType: "GPU", Quantity: 4.0}
	if err := res.Validate(); err != nil {
		t.Errorf("expected valid resource: %v", err)
	}
	res.Quantity = -1.0
	if err := res.Validate(); err == nil {
		t.Error("expected error for negative resource quantity")
	}
}

