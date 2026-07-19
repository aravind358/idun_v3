package planning

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestSearchEnumerations(t *testing.T) {
	// Test NodeStatus
	validNodeStatuses := []NodeStatus{
		NodeStatusUnexpanded, NodeStatusOpen, NodeStatusClosed,
		NodeStatusTerminalGoal, NodeStatusPrunedBudget, NodeStatusPrunedConstitutional,
	}
	for _, status := range validNodeStatuses {
		if !status.IsValid() {
			t.Errorf("expected NodeStatus %q to be valid", status)
		}
		if status.String() != string(status) {
			t.Errorf("expected String()=%q, got %q", string(status), status.String())
		}
	}
	if (NodeStatus("INVALID_STATUS")).IsValid() {
		t.Error("expected invalid NodeStatus to return false from IsValid()")
	}

	// Test EdgeType
	validEdgeTypes := []EdgeType{
		EdgeTypeStrategicOperator, EdgeTypeAdversarialContingency,
	}
	for _, edgeType := range validEdgeTypes {
		if !edgeType.IsValid() {
			t.Errorf("expected EdgeType %q to be valid", edgeType)
		}
		if edgeType.String() != string(edgeType) {
			t.Errorf("expected String()=%q, got %q", string(edgeType), edgeType.String())
		}
	}
	if (EdgeType("INVALID_EDGE")).IsValid() {
		t.Error("expected invalid EdgeType to return false from IsValid()")
	}

	// Test Reversibility
	validReversibilities := []Reversibility{
		ReversibilityTrivial, ReversibilityHighCost, ReversibilityCritical,
	}
	for _, rev := range validReversibilities {
		if !rev.IsValid() {
			t.Errorf("expected Reversibility %q to be valid", rev)
		}
		if rev.String() != string(rev) {
			t.Errorf("expected String()=%q, got %q", string(rev), rev.String())
		}
	}
	if (Reversibility("INVALID_REV")).IsValid() {
		t.Error("expected invalid Reversibility to return false from IsValid()")
	}
}

func TestCostVector_ValidationAndArithmetic(t *testing.T) {
	c1 := CostVector{
		Time:                   10 * time.Millisecond,
		Resources:              5.0,
		MonetaryCost:           0.10,
		Risk:                   0.05,
		IrreversibilityPenalty: 0.0,
	}
	if err := c1.Validate(); err != nil {
		t.Fatalf("expected valid CostVector, got err: %v", err)
	}
	if c1.IsZero() {
		t.Error("expected c1 to not be zero")
	}

	c2 := CostVector{
		Time:                   20 * time.Millisecond,
		Resources:              10.0,
		MonetaryCost:           0.20,
		Risk:                   0.10,
		IrreversibilityPenalty: 50.0,
	}

	sum := c1.Add(c2)
	if sum.Time != 30*time.Millisecond || math.Abs(sum.Resources-15.0) > 1e-9 || math.Abs(sum.MonetaryCost-0.30) > 1e-9 || math.Abs(sum.Risk-0.15) > 1e-9 || math.Abs(sum.IrreversibilityPenalty-50.0) > 1e-9 {
		t.Errorf("unexpected CostVector sum result: %+v", sum)
	}

	zero := CostVector{}
	if !zero.IsZero() {
		t.Error("expected zero vector to report true from IsZero()")
	}

	// Test bounds validation
	invalidRisk := CostVector{Risk: 1.5}
	if err := invalidRisk.Validate(); err == nil {
		t.Error("expected error for risk > 1.0")
	}
	invalidTime := CostVector{Time: -1 * time.Second}
	if err := invalidTime.Validate(); err == nil {
		t.Error("expected error for negative time")
	}
	invalidRes := CostVector{Resources: -5.0}
	if err := invalidRes.Validate(); err == nil {
		t.Error("expected error for negative resources")
	}
	invalidMoney := CostVector{MonetaryCost: -0.5}
	if err := invalidMoney.Validate(); err == nil {
		t.Error("expected error for negative monetary cost")
	}
}

func TestNodeCostProfile_Validation(t *testing.T) {
	ncp := NodeCostProfile{
		AccumulatedCost:        CostVector{Time: 10 * time.Millisecond, Risk: 0.1},
		EstimatedRemainingCost: CostVector{Time: 20 * time.Millisecond, Risk: 0.2},
		EvaluationScore:        15.5,
	}
	if err := ncp.Validate(); err != nil {
		t.Fatalf("expected valid NodeCostProfile, got: %v", err)
	}

	var nilProfile *NodeCostProfile
	if err := nilProfile.Validate(); err == nil {
		t.Error("expected error validating nil NodeCostProfile")
	}

	badProfile := NodeCostProfile{
		AccumulatedCost: CostVector{Risk: -0.1},
	}
	if err := badProfile.Validate(); err == nil {
		t.Error("expected error when accumulated cost is invalid")
	}
}

func TestSearchStep_ValidationAndCloning(t *testing.T) {
	step := SearchStep{
		StepID:        "step-1",
		StepIndex:     0,
		AppliedEdgeID: "edge-a",
		OperatorName:  "MigrateShard",
		TransitionCost: CostVector{
			Time:      5 * time.Millisecond,
			Resources: 2.0,
		},
		RiskIncurred:    0.05,
		SimulatedOffset: 5 * time.Millisecond,
		Metadata:        map[string]string{"region": "us-east-1"},
	}
	if err := step.Validate(); err != nil {
		t.Fatalf("expected valid SearchStep, got: %v", err)
	}

	clone := step.Clone()
	if clone.StepID != step.StepID || clone.OperatorName != step.OperatorName || clone.Metadata["region"] != "us-east-1" {
		t.Errorf("clone mismatch: %+v", clone)
	}

	// Mutate clone metadata to verify deep copy safety
	clone.Metadata["region"] = "eu-west-1"
	if step.Metadata["region"] != "us-east-1" {
		t.Error("mutating cloned SearchStep metadata tainted original")
	}

	var nilStep *SearchStep
	if err := nilStep.Validate(); err == nil {
		t.Error("expected error validating nil SearchStep")
	}
}

func TestSearchEdge_ConstructorAndValidation(t *testing.T) {
	edge := NewSearchEdge("edge-101", EdgeTypeStrategicOperator, "ProvisionFailover")
	if edge.Preconditions == nil || edge.RequiredAssumptions == nil || edge.Postconditions == nil {
		t.Fatal("expected non-nil maps in NewSearchEdge")
	}
	if edge.Reversibility != ReversibilityTrivial {
		t.Errorf("expected default reversibility %v, got %v", ReversibilityTrivial, edge.Reversibility)
	}
	if err := edge.Validate(); err != nil {
		t.Fatalf("expected new edge to be valid, got: %v", err)
	}

	edge.Preconditions["cluster_status"] = "inactive"
	edge.RequiredAssumptions["api_availability"] = "high"
	edge.Postconditions["cluster_status"] = "active"
	edge.RiskDelta = 0.10
	edge.EdgeCost = CostVector{Time: 100 * time.Millisecond, MonetaryCost: 1.50}

	if err := edge.Validate(); err != nil {
		t.Fatalf("expected populated edge to be valid, got: %v", err)
	}

	// Deep copy test
	clone := edge.Clone()
	clone.Preconditions["cluster_status"] = "degraded"
	if edge.Preconditions["cluster_status"] != "inactive" {
		t.Error("mutating clone tainted original Preconditions map")
	}

	// Test invalid scenarios
	badEdge := &SearchEdge{EdgeID: "", EdgeType: EdgeTypeStrategicOperator, OperatorName: "op"}
	if err := badEdge.Validate(); err == nil {
		t.Error("expected error when EdgeID is empty")
	}
}

func TestSearchState_ConstructorValidationAndCloning(t *testing.T) {
	state := NewSearchState("state-root")
	if state.SimulatedWorldState == nil || state.RemainingDesiredState == nil || state.ActiveConstraints == nil || state.Assumptions == nil {
		t.Fatal("expected non-nil maps from NewSearchState")
	}
	if state.ExecutedTrajectory == nil || state.ResourceReservations == nil {
		t.Fatal("expected non-nil slices from NewSearchState")
	}
	if state.EpistemicConfidence != 1.0 {
		t.Errorf("expected default confidence 1.0, got %f", state.EpistemicConfidence)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("expected root state to be valid, got: %v", err)
	}

	state.SimulatedWorldState["db_status"] = "running"
	state.RemainingDesiredState["migration_complete"] = "true"
	state.ActiveConstraints["max_downtime_sec"] = "0"
	state.Assumptions["network_connectivity"] = "stable"
	state.ExecutedTrajectory = append(state.ExecutedTrajectory, SearchStep{
		StepID:        "s-1",
		StepIndex:     0,
		AppliedEdgeID: "e-1",
		OperatorName:  "InitMigration",
	})
	state.ResourceReservations = append(state.ResourceReservations, ResourceRequirement{
		ResourceID:   "res-gpu",
		ResourceType: "GPU_UNITS",
		Quantity:     2.0,
	})
	state.AccumulatedRisk = 0.05
	state.EpistemicConfidence = 0.92

	if err := state.Validate(); err != nil {
		t.Fatalf("expected populated state to be valid, got: %v", err)
	}

	// Test deep cloning safety across all maps and slices
	clone := state.Clone()
	if clone.StateID != state.StateID || len(clone.ExecutedTrajectory) != 1 || len(clone.ResourceReservations) != 1 {
		t.Fatalf("clone structural mismatch: %+v", clone)
	}

	clone.SimulatedWorldState["db_status"] = "stopped"
	clone.Assumptions["network_connectivity"] = "degraded"
	clone.ExecutedTrajectory[0].OperatorName = "MutatedOp"
	clone.ResourceReservations[0].Quantity = 99.0

	if state.SimulatedWorldState["db_status"] != "running" {
		t.Error("mutating clone tainted original SimulatedWorldState")
	}
	if state.Assumptions["network_connectivity"] != "stable" {
		t.Error("mutating clone tainted original Assumptions")
	}
	if state.ExecutedTrajectory[0].OperatorName != "InitMigration" {
		t.Error("mutating clone step tainted original ExecutedTrajectory")
	}
	if state.ResourceReservations[0].Quantity != 2.0 {
		t.Error("mutating clone reservation tainted original ResourceReservations")
	}
}

func TestSearchNode_TopologyAndValidation(t *testing.T) {
	rootState := NewSearchState("state-root")
	rootNode := NewSearchNode("node-root", nil, rootState, nil)
	if rootNode.Status != NodeStatusUnexpanded {
		t.Errorf("expected unexpanded status, got %s", rootNode.Status)
	}
	if err := rootNode.Validate(); err != nil {
		t.Fatalf("expected valid root node, got: %v", err)
	}

	edge := NewSearchEdge("edge-1", EdgeTypeStrategicOperator, "Op1")
	childState := rootState.Clone()
	childState.StateID = "state-child-1"
	childNode := NewSearchNode("node-child-1", rootNode, childState, edge)
	childNode.Status = NodeStatusOpen
	childNode.CostProfile = NodeCostProfile{
		AccumulatedCost:        CostVector{Time: 10 * time.Millisecond},
		EstimatedRemainingCost: CostVector{Time: 5 * time.Millisecond},
		EvaluationScore:        15.0,
	}

	if err := childNode.Validate(); err != nil {
		t.Fatalf("expected valid child node, got: %v", err)
	}
	if childNode.Parent != rootNode {
		t.Error("expected child parent back-pointer to match rootNode")
	}
}

func TestSearchTypes_SerializationCompatibility(t *testing.T) {
	state := NewSearchState("state-serialize")
	state.SimulatedWorldState["foo"] = "bar"
	state.Assumptions["net"] = "ok"
	state.ExecutedTrajectory = append(state.ExecutedTrajectory, SearchStep{
		StepID:        "step-s1",
		StepIndex:     0,
		AppliedEdgeID: "e-1",
		OperatorName:  "Op",
	})

	edge := NewSearchEdge("edge-s1", EdgeTypeStrategicOperator, "Op")
	node := NewSearchNode("node-s1", nil, state, edge)
	node.Status = NodeStatusClosed

	// Marshal node to JSON (Parent is omitted via `json:"-"` to prevent circular serialization)
	data, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal SearchNode to JSON: %v", err)
	}

	var decoded SearchNode
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal SearchNode from JSON: %v", err)
	}

	if decoded.NodeID != node.NodeID || decoded.Status != node.Status {
		t.Errorf("decoded node mismatch: %+v vs %+v", decoded, *node)
	}
	if decoded.State == nil || decoded.State.StateID != state.StateID || decoded.State.SimulatedWorldState["foo"] != "bar" {
		t.Errorf("decoded state mismatch: %+v", decoded.State)
	}
	if decoded.IncomingEdge == nil || decoded.IncomingEdge.EdgeID != edge.EdgeID {
		t.Errorf("decoded incoming edge mismatch: %+v", decoded.IncomingEdge)
	}
}

func TestSearchNode_CloneAndImmutability(t *testing.T) {
	rootNode := NewSearchNode("node-root", nil, NewSearchState("state-root"), nil)
	state := NewSearchState("state-1")
	state.SimulatedWorldState["key1"] = "val1"
	edge := NewSearchEdge("edge-1", EdgeTypeStrategicOperator, "Op1")
	edge.Preconditions["cond1"] = "true"

	node := NewSearchNode("node-1", rootNode, state, edge)
	node.PlannerMetadata["meta1"] = "info1"

	clone := node.Clone()
	if clone.NodeID != node.NodeID || clone.Parent != rootNode {
		t.Fatalf("clone structural mismatch: %+v", clone)
	}

	// Mutate clone across all levels (node, state, edge, metadata)
	clone.PlannerMetadata["meta1"] = "mutated_info"
	clone.State.SimulatedWorldState["key1"] = "mutated_val"
	clone.IncomingEdge.Preconditions["cond1"] = "false"

	// Verify original remained completely untouched
	if node.PlannerMetadata["meta1"] != "info1" {
		t.Error("mutating clone PlannerMetadata tainted original node")
	}
	if node.State.SimulatedWorldState["key1"] != "val1" {
		t.Error("mutating clone State tainted original node State")
	}
	if node.IncomingEdge.Preconditions["cond1"] != "true" {
		t.Error("mutating clone IncomingEdge tainted original node IncomingEdge")
	}
}
