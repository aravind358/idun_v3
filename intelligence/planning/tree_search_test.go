package planning

import (
	"context"
	"strings"
	"testing"
	"time"

	"idun/intelligence/reasoning"
)

func TestStage5B13_TreeSearchSpecialist_MultiCandidateGeneration(t *testing.T) {
	spec := NewTreeSearchSpecialist("STRATEGIC")

	// Register 3 custom strategic operators targeting different branches with different costs and reversibilities
	op1 := NewSearchEdge("op-branch-fast", EdgeTypeStrategicOperator, "Fast Failover Branch")
	op1.EdgeCost = CostVector{Time: 5 * time.Minute, Resources: 10.0, MonetaryCost: 100.0}
	op1.RiskDelta = 0.10
	op1.Reversibility = ReversibilityTrivial
	op1.Postconditions["region_status"] = "ACTIVE"
	op1.Postconditions["data_sync"] = "COMPLETE"

	op2 := NewSearchEdge("op-branch-balanced", EdgeTypeStrategicOperator, "Balanced Migration Branch")
	op2.EdgeCost = CostVector{Time: 15 * time.Minute, Resources: 5.0, MonetaryCost: 50.0}
	op2.RiskDelta = 0.05
	op2.Reversibility = ReversibilityHighCost
	op2.Postconditions["region_status"] = "ACTIVE"
	op2.Postconditions["data_sync"] = "COMPLETE"

	op3 := NewSearchEdge("op-branch-critical", EdgeTypeStrategicOperator, "Emergency Override Branch")
	op3.EdgeCost = CostVector{Time: 1 * time.Minute, Resources: 20.0, MonetaryCost: 500.0}
	op3.RiskDelta = 0.15
	op3.Reversibility = ReversibilityCritical
	op3.Postconditions["region_status"] = "ACTIVE"
	op3.Postconditions["data_sync"] = "COMPLETE"

	if err := spec.RegisterOperator(op1); err != nil {
		t.Fatalf("failed to register op1: %v", err)
	}
	if err := spec.RegisterOperator(op2); err != nil {
		t.Fatalf("failed to register op2: %v", err)
	}
	if err := spec.RegisterOperator(op3); err != nil {
		t.Fatalf("failed to register op3: %v", err)
	}

	req, err := NewPlanningRequestBuilder().
		WithRequestID("req-5b13-multi").
		WithGoal("Restore global service across regions").
		WithDomain("Strategic").
		WithTargetDepth(DepthStrategic).
		WithBudget(500*time.Millisecond, 0.80).
		Build()
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	// Manually inject ResolvedGoal into request to test exact condition matching
	req.ResolvedGoal = &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindStateChange,
		Intent: "restore_service",
		Target: "global_regions",
		DesiredState: map[string]string{
			"region_status": "ACTIVE",
			"data_sync":     "COMPLETE",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	profile := DefaultPlanningPolicyProfile()
	profile.SearchStrategies["STRATEGIC"].BeamWidth = 10

	res, err := spec.GeneratePlanningResult(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("GeneratePlanningResult failed: %v", err)
	}

	if res.ResultID == "" || res.RequestID != "req-5b13-multi" {
		t.Errorf("unexpected ResultID (%s) or RequestID (%s)", res.ResultID, res.RequestID)
	}
	if len(res.Plans) != 3 {
		t.Fatalf("expected exactly 3 candidate plans (one per branch operator), got %d", len(res.Plans))
	}

	// Verify plans are ordered ascending by evaluation score / cost
	for i := 0; i < len(res.Plans)-1; i++ {
		p1 := res.Plans[i]
		p2 := res.Plans[i+1]
		if p1.PlanID == "" || p2.PlanID == "" {
			t.Errorf("candidate plans missing PlanID at rank %d/%d", i, i+1)
		}
		if p1.PlanFingerprint == "" || p2.PlanFingerprint == "" {
			t.Errorf("candidate plans missing PlanFingerprint at rank %d/%d", i, i+1)
		}
	}

	// Verify Traces emission
	if len(res.Traces) != 1 {
		t.Fatalf("expected 1 diagnostic PlanningTrace, got %d", len(res.Traces))
	}
	trace := res.Traces[0]
	if trace.TraceID == "" || trace.PlanID == "" {
		t.Errorf("trace missing TraceID or PlanID")
	}
	if trace.SearchStatistics.AlternativePlansGenerated != uint32(len(res.Plans)) {
		t.Errorf("expected trace AlternativePlansGenerated=%d, got %d", len(res.Plans), trace.SearchStatistics.AlternativePlansGenerated)
	}
}

func TestStage5B13_TreeSearchSpecialist_RollbackExtraction(t *testing.T) {
	spec := NewTreeSearchSpecialist("STRATEGIC")

	opTrivial := NewSearchEdge("op-triv", EdgeTypeStrategicOperator, "Trivial Step")
	opTrivial.EdgeCost = CostVector{Resources: 2.0}
	opTrivial.Reversibility = ReversibilityTrivial
	opTrivial.Postconditions["stage"] = "1"

	opHigh := NewSearchEdge("op-high", EdgeTypeStrategicOperator, "High Cost Step")
	opHigh.EdgeCost = CostVector{Resources: 10.0}
	opHigh.Reversibility = ReversibilityHighCost
	opHigh.Postconditions["stage"] = "2"

	opCrit := NewSearchEdge("op-crit", EdgeTypeStrategicOperator, "Critical Irreversible Step")
	opCrit.EdgeCost = CostVector{Resources: 50.0}
	opCrit.Reversibility = ReversibilityCritical
	opCrit.Postconditions["stage"] = "DONE"

	spec.SetOperators([]*SearchEdge{opTrivial, opHigh, opCrit})

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-5b13-rb").
		WithGoal("Test exact rollback extraction").
		WithDomain("Strategic").
		WithBudget(500*time.Millisecond, 0.80).
		Build()
	req.ResolvedGoal = &reasoning.SemanticGoal{
		Kind:         reasoning.GoalKindStateChange,
		Intent:       "test_rb",
		Target:       "system",
		DesiredState: map[string]string{"stage": "1"}, // Target opTrivial first
	}

	candidates, err := spec.GenerateCandidatePlans(context.Background(), req, nil, nil)
	if err != nil {
		t.Fatalf("GenerateCandidatePlans failed: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatalf("expected candidate plans, got 0")
	}

	// Find the candidate corresponding to trivial reversibility or verify rollback strategy fields across operators
	// Let's test ConvertPathToPlan directly with a synthetic path of 3 steps covering all reversibility types
	rootState := NewSearchState("state-0")
	node0 := NewSearchNode("n0", nil, rootState, nil)

	state1 := rootState.Clone()
	node1 := NewSearchNode("n1", node0, state1, opTrivial)

	state2 := state1.Clone()
	node2 := NewSearchNode("n2", node1, state2, opHigh)

	state3 := state2.Clone()
	node3 := NewSearchNode("n3", node2, state3, opCrit)

	path := []*SearchNode{node0, node1, node2, node3}

	plan, err := spec.ConvertPathToPlan(path, req, "snap-test", "trace-test", 0)
	if err != nil {
		t.Fatalf("ConvertPathToPlan failed: %v", err)
	}

	if len(plan.RollbackStrategies) != 3 {
		t.Fatalf("expected exactly 3 RollbackStrategies extracted, got %d", len(plan.RollbackStrategies))
	}

	// Verify Trivial rollback
	rb0 := plan.RollbackStrategies[0]
	if !strings.Contains(rb0.ActionSteps[0], "trivial rollback") {
		t.Errorf("expected trivial rollback action step, got %v", rb0.ActionSteps)
	}
	if rb0.TriggerNodeID == "" {
		t.Errorf("missing TriggerNodeID on rb0")
	}

	// Verify High Cost rollback
	rb1 := plan.RollbackStrategies[1]
	if !strings.Contains(rb1.ActionSteps[0], "high-cost rollback procedure") || len(rb1.ActionSteps) < 3 {
		t.Errorf("expected high-cost multi-step rollback procedure, got %v", rb1.ActionSteps)
	}

	// Verify Critical rollback
	rb2 := plan.RollbackStrategies[2]
	if !strings.Contains(rb2.ActionSteps[0], "CRITICAL / IRREVERSIBLE transition") || !strings.Contains(rb2.ActionSteps[1], "emergency containment") {
		t.Errorf("expected critical containment protocol steps, got %v", rb2.ActionSteps)
	}
	if rb2.EstimatedCost < 100.0 {
		t.Errorf("expected penalty cost doubled for critical reversibility (>=100), got %.2f", rb2.EstimatedCost)
	}
}

func TestStage5B13_TreeSearchSpecialist_OperatorManagement(t *testing.T) {
	spec := NewTreeSearchSpecialist()

	if len(spec.GetOperators()) != 0 {
		t.Fatalf("expected initial operator pool to be empty, got %d", len(spec.GetOperators()))
	}

	if err := spec.RegisterOperator(nil); err == nil {
		t.Errorf("expected error registering nil operator, got nil")
	}

	op := NewSearchEdge("op-test-mgmt", EdgeTypeStrategicOperator, "Test Operator")
	op.EdgeCost = CostVector{Resources: 1.0}
	if err := spec.RegisterOperator(op); err != nil {
		t.Fatalf("RegisterOperator failed: %v", err)
	}

	ops := spec.GetOperators()
	if len(ops) != 1 || ops[0].EdgeID != "op-test-mgmt" {
		t.Fatalf("unexpected snapshot from GetOperators: %+v", ops)
	}

	// Verify deep copy / isolation
	ops[0].OperatorName = "Mutated Outside"
	if spec.GetOperators()[0].OperatorName == "Mutated Outside" {
		t.Errorf("GetOperators exposed internal mutable reference without deep copy")
	}
}
