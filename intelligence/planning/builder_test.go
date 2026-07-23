package planning

import (
	"testing"
	"time"
)

func TestPlanningRequestBuilder_Fluent(t *testing.T) {
	req, err := NewPlanningRequestBuilder().
		WithRequestID("req-500").
		WithGoal("Build multi-agent collaboration workflow").
		WithDomain("SoftwareEngineering").
		WithContextRef("storage://cas/frame-888").
		WithConstraints([]string{"No downtime"}, []string{"Fast execution"}).
		WithTargetDepth(DepthStrategic).
		WithBudget(400*time.Millisecond, 0.75).
		Build()

	if err != nil {
		t.Fatalf("expected builder success, got error: %v", err)
	}

	if req.RequestID != "req-500" || req.Goal != "Build multi-agent collaboration workflow" {
		t.Errorf("unexpected field values in built PlanningRequest: %+v", req)
	}
	if req.Domain != "SoftwareEngineering" || req.MinConfidenceFloor != 0.75 {
		t.Errorf("unexpected domain or confidence floor: %s / %f", req.Domain, req.MinConfidenceFloor)
	}
}

func TestPlanningRequestBuilder_ErrorChaining(t *testing.T) {
	_, err := NewPlanningRequestBuilder().
		WithRequestID("req-err").
		WithBudget(-10*time.Millisecond, 0.50). // Invalid budget
		WithGoal("Should fail").
		Build()

	if err == nil {
		t.Fatal("expected builder to fail due to negative budget, got nil")
	}
}

func TestPlanBuilder_Fluent(t *testing.T) {
	cp := ConfidenceProfile{
		GoalConfidence: 0.9, PreconditionConfidence: 0.85, DependencyConfidence: 0.88,
		ResourceConfidence: 0.82, TimingConfidence: 0.80, ConstraintConfidence: 0.90,
		OverallConfidence: 0.80,
	}

	plan, err := NewPlanBuilder().
		WithIdentity("plan-777", "snap-1", "trace-777").
		WithGoalAndDomain("Deploy database schema", "Coding", "TACTICAL").
		AddSubgoal(Subgoal{SubgoalID: "sg-1", Title: "Check migrations"}).
		AddDependency(DependencyEdge{EdgeID: "e-1", SourceNodeID: "sg-1", TargetNodeID: "sg-2"}).
		WithEstimates(10.0, 15*time.Minute, []ResourceRequirement{
			{ResourceID: "res-db", ResourceType: "DATABASE_CONN", Quantity: 1.0},
		}).
		WithConfidenceProfile(cp).
		WithReplayMetadata(ReplayMetadata{StrategySnapshotID: "snap-1", ReplayFidelity: "EXACT"}).
		WithFingerprint("abc123hash").
		Build()

	if err != nil {
		t.Fatalf("expected valid CandidatePlan built, got: %v", err)
	}

	if plan.PlanID != "plan-777" || plan.TraceID != "trace-777" || len(plan.Subgoals) != 1 {
		t.Errorf("built plan fields mismatch: %+v", plan)
	}
	if plan.SchemaVersion != SchemaVersion2_0_0 {
		t.Errorf("expected frozen schema version %s, got %s", SchemaVersion2_0_0, plan.SchemaVersion)
	}
}

func TestPlanningTraceBuilder_Fluent(t *testing.T) {
	cp := ConfidenceProfile{
		GoalConfidence: 0.9, PreconditionConfidence: 0.9, DependencyConfidence: 0.9,
		ResourceConfidence: 0.9, TimingConfidence: 0.9, ConstraintConfidence: 0.9,
		OverallConfidence: 0.9,
	}
	qm := QualityMetrics{
		Completeness: 0.95, Efficiency: 0.90, Robustness: 0.92,
	}
	stats := SearchStatistics{
		NodesExpanded: 50, NodesPruned: 12, CacheHits: 1,
	}

	trace, err := NewPlanningTraceBuilder().
		WithIdentity("trace-101", "plan-101", "snap-1").
		AddStepLog(PlanningStepLog{StepIndex: 1, SpecialistName: "GoalDecomposition", Status: "DONE"}).
		AddRejectedBranch(RejectedBranch{BranchID: "alt-2", DiscardReason: "High latency"}).
		WithDiagnostics(TerminationGoalFound, stats, 5.0, cp, qm).
		Build()

	if err != nil {
		t.Fatalf("expected valid PlanningTrace built, got: %v", err)
	}

	if trace.TraceID != "trace-101" || trace.TerminationReason != TerminationGoalFound {
		t.Errorf("built trace fields mismatch: %+v", trace)
	}
	if len(trace.RejectedBranches) != 1 || trace.SearchStatistics.NodesExpanded != 50 {
		t.Errorf("built trace diagnostic mismatch")
	}
}

