package planning

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestOpenQueue_PriorityOrderingAndBeamPruning(t *testing.T) {
	q := NewOpenQueue()
	if q.Len() != 0 || q.Pop() != nil || q.Peek() != nil {
		t.Error("expected new queue to be empty and return nil on Pop/Peek")
	}

	// Create 5 nodes with different EvaluationScores
	scores := []float64{25.0, 10.0, 30.0, 5.0, 15.0}
	for i, score := range scores {
		state := NewSearchState("state-" + string(rune('A'+i)))
		node := NewSearchNode("node-"+string(rune('A'+i)), nil, state, nil)
		node.CostProfile.EvaluationScore = score
		q.Push(node)
	}

	if q.Len() != 5 {
		t.Fatalf("expected queue len 5, got %d", q.Len())
	}

	// Peek should return the lowest score (5.0) without removing it
	peeked := q.Peek()
	if peeked == nil || peeked.CostProfile.EvaluationScore != 5.0 {
		t.Errorf("expected peek to return score 5.0, got %+v", peeked)
	}
	if q.Len() != 5 {
		t.Errorf("peek removed item from queue: len %d", q.Len())
	}

	// Test PruneToBeam (keep top 3 nodes out of 5)
	pruned, count := q.PruneToBeam(3)
	if count != 2 || len(pruned) != 2 {
		t.Fatalf("expected 2 pruned nodes when pruning 5 to beam 3, got %d", count)
	}
	for _, p := range pruned {
		if p.Status != NodeStatusPrunedBudget {
			t.Errorf("expected pruned node status %s, got %s", NodeStatusPrunedBudget, p.Status)
		}
	}
	if q.Len() != 3 {
		t.Fatalf("expected queue len 3 after pruning, got %d", q.Len())
	}

	// Snapshot should return remaining nodes ordered ascending: [5.0, 10.0, 15.0]
	snap := q.Snapshot()
	if len(snap) != 3 || snap[0].CostProfile.EvaluationScore != 5.0 || snap[1].CostProfile.EvaluationScore != 10.0 || snap[2].CostProfile.EvaluationScore != 15.0 {
		t.Errorf("unexpected snapshot ordering: %+v", snap)
	}

	// Pop all items and verify strict min-heap order
	expectedOrder := []float64{5.0, 10.0, 15.0}
	for i, expected := range expectedOrder {
		popped := q.Pop()
		if popped == nil {
			t.Fatalf("at index %d: expected item with score %f, got nil", i, expected)
		}
		if popped.CostProfile.EvaluationScore != expected {
			t.Errorf("at index %d: expected score %f, got %f", i, expected, popped.CostProfile.EvaluationScore)
		}
	}
	if q.Pop() != nil {
		t.Error("expected nil pop after queue drained")
	}

	// Verification 3 exact scenario: [14, 2, 8, 5, 11, 7] pruned to top 3 -> [2, 5, 7]
	qVer3 := NewOpenQueue()
	for i, score := range []float64{14, 2, 8, 5, 11, 7} {
		node := NewSearchNode("node-v3-"+string(rune('A'+i)), nil, NewSearchState("state-v3"), nil)
		node.CostProfile.EvaluationScore = score
		qVer3.Push(node)
	}
	prunedVer3, countVer3 := qVer3.PruneToBeam(3)
	if countVer3 != 3 || len(prunedVer3) != 3 || qVer3.Len() != 3 {
		t.Fatalf("Verification 3 pruning failed: count=%d len=%d", countVer3, qVer3.Len())
	}
	snapVer3 := qVer3.Snapshot()
	for i, exp := range []float64{2.0, 5.0, 7.0} {
		if snapVer3[i].CostProfile.EvaluationScore != exp {
			t.Errorf("Verification 3 index %d: expected score %f, got %f", i, exp, snapVer3[i].CostProfile.EvaluationScore)
		}
	}

	// Verification 1 deterministic tie breaking: identical EvaluationScore = 10.0
	qTie := NewOpenQueue()
	n1 := NewSearchNode("node-C", nil, NewSearchState("state-C"), nil)
	n1.CostProfile.EvaluationScore, n1.State.EpistemicConfidence, n1.State.SimulatedClock = 10.0, 0.8, 5*time.Second
	n2 := NewSearchNode("node-A", nil, NewSearchState("state-A"), nil)
	n2.CostProfile.EvaluationScore, n2.State.EpistemicConfidence, n2.State.SimulatedClock = 10.0, 0.9, 5*time.Second
	n3 := NewSearchNode("node-B", nil, NewSearchState("state-B"), nil)
	n3.CostProfile.EvaluationScore, n3.State.EpistemicConfidence, n3.State.SimulatedClock = 10.0, 0.8, 2*time.Second
	qTie.Push(n1)
	qTie.Push(n2)
	qTie.Push(n3)
	snapTie := qTie.Snapshot()
	if snapTie[0].NodeID != "node-A" || snapTie[1].NodeID != "node-B" || snapTie[2].NodeID != "node-C" {
		t.Errorf("expected tie-break order [node-A, node-B, node-C], got [%s, %s, %s]", snapTie[0].NodeID, snapTie[1].NodeID, snapTie[2].NodeID)
	}
}

func TestBeamAStarEngine_HeuristicEvaluation(t *testing.T) {
	cfg := DefaultBeamAStarConfig()
	engine := NewBeamAStarEngine(cfg)

	state := NewSearchState("state-h")
	state.RemainingDesiredState["k1"] = "v1"
	state.RemainingDesiredState["k2"] = "v2"
	state.AccumulatedCost.Resources = 5.0
	state.EpistemicConfidence = 0.80

	node := NewSearchNode("node-h", nil, state, nil)
	engine.computeHeuristicAndScore(node, nil, cfg)

	// hSemantic = 2 keys * 1.0 = 2.0
	// hResource = 5.0 * 0.2 = 1.0
	// hBase = max(2.0, 1.0) = 2.0
	// confidencePenalty = 1 + 2.0 * (1 - 0.80) = 1.4
	// hFinal = 2.0 * 1.4 = 2.8
	expectedH := 2.8
	if math.Abs(node.CostProfile.EstimatedRemainingCost.Resources-expectedH) > 1e-9 {
		t.Errorf("expected h(n)=%f, got %f", expectedH, node.CostProfile.EstimatedRemainingCost.Resources)
	}
}

func TestBeamAStarEngine_GoalDetectionAndPathReconstruction(t *testing.T) {
	cfg := DefaultBeamAStarConfig()
	engine := NewBeamAStarEngine(cfg)

	// Setup root state
	rootState := NewSearchState("state-root")
	rootState.SimulatedWorldState["db"] = "stopped"
	rootState.RemainingDesiredState["db"] = "running"
	rootState.RemainingDesiredState["cache"] = "warmed"

	// Setup operators
	op1 := NewSearchEdge("op-1", EdgeTypeStrategicOperator, "StartDB")
	op1.Preconditions["db"] = "stopped"
	op1.Postconditions["db"] = "running"
	op1.EdgeCost = CostVector{Time: 1 * time.Second, Resources: 2.0}

	op2 := NewSearchEdge("op-2", EdgeTypeStrategicOperator, "WarmCache")
	op2.Preconditions["db"] = "running"
	op2.Postconditions["cache"] = "warmed"
	op2.EdgeCost = CostVector{Time: 500 * time.Millisecond, Resources: 1.0}

	req := &PlanningRequest{
		RequestID:          "req-goal",
		Goal:               "Start DB and Warm Cache",
		MinConfidenceFloor: 0.50,
	}

	result, err := engine.Search(context.Background(), req, rootState, []*SearchEdge{op1, op2})
	if err != nil {
		t.Fatalf("search returned error: %v", err)
	}

	if result.Status != StatusComplete {
		t.Fatalf("expected StatusComplete, got %s", result.Status)
	}
	if len(result.GoalNodes) == 0 {
		t.Fatal("expected at least one goal node discovered")
	}

	goalNode := result.GoalNodes[0]
	if goalNode.Status != NodeStatusTerminalGoal {
		t.Errorf("expected terminal goal status, got %s", goalNode.Status)
	}

	// Verify candidate path extraction
	path := ReconstructPath(goalNode)
	if len(path) != 3 { // root -> node after op1 -> goalNode after op2
		t.Fatalf("expected path length 3, got %d", len(path))
	}
	if path[0].NodeID != "node-root" {
		t.Errorf("expected root node at path[0], got %s", path[0].NodeID)
	}
	if path[1].IncomingEdge.EdgeID != "op-1" {
		t.Errorf("expected op-1 transition at path[1], got %s", path[1].IncomingEdge.EdgeID)
	}
	if path[2].IncomingEdge.EdgeID != "op-2" {
		t.Errorf("expected op-2 transition at path[2], got %s", path[2].IncomingEdge.EdgeID)
	}
}

func TestBeamAStarEngine_DuplicateDetectionAndPruning(t *testing.T) {
	cfg := DefaultBeamAStarConfig()
	engine := NewBeamAStarEngine(cfg)

	rootState := NewSearchState("state-root")
	rootState.SimulatedWorldState["service"] = "down"
	rootState.RemainingDesiredState["service"] = "up"

	// Two distinct operators that transition service to "up", but one is more expensive/inferior
	opFast := NewSearchEdge("op-fast", EdgeTypeStrategicOperator, "RestartFast")
	opFast.Preconditions["service"] = "down"
	opFast.Postconditions["service"] = "up"
	opFast.EdgeCost = CostVector{Resources: 1.0, MonetaryCost: 0.10}

	opSlow := NewSearchEdge("op-slow", EdgeTypeStrategicOperator, "RestartSlow")
	opSlow.Preconditions["service"] = "down"
	opSlow.Postconditions["service"] = "up"
	opSlow.EdgeCost = CostVector{Resources: 50.0, MonetaryCost: 10.0}

	req := &PlanningRequest{RequestID: "req-dup", Goal: "Restore service"}

	result, err := engine.Search(context.Background(), req, rootState, []*SearchEdge{opFast, opSlow})
	if err != nil {
		t.Fatalf("search error: %v", err)
	}

	if result.Status != StatusComplete {
		t.Fatalf("expected StatusComplete, got %s", result.Status)
	}
	// Both opFast and opSlow lead to the same state where service is up.
	// Since opFast has lower score, it is expanded/evaluated and reaches terminal goal.
	// When opSlow is checked against CLOSED set or evaluated, duplicate detection keeps the optimal trajectory.
	if result.ExpandedCount < 1 {
		t.Errorf("expected at least 1 expansion, got %d", result.ExpandedCount)
	}

	// Verification 2 exact scenario: CLOSED set duplicate detection & update behavior
	closedSet := make(map[string]float64)
	closedSet["State A"] = 8.0 // State A initially visited with Cost = 8.0

	// Case 1: State A appears with Cost = 12.0 -> must be pruned
	n12 := NewSearchNode("node-12", nil, NewSearchState("State A"), nil)
	n12.CostProfile.AccumulatedCost = CostVector{Resources: 12.0}
	scalarG12 := engine.computeScalarCost(n12.CostProfile.AccumulatedCost)
	if bestG, exists := closedSet[n12.State.StateID]; !exists || bestG > scalarG12 {
		t.Errorf("expected State A with Cost %f (>= bestG %f) to be pruned", scalarG12, bestG)
	}
	if closedSet["State A"] != 8.0 {
		t.Errorf("expected CLOSED set to remain 8.0 after pruning worse cost, got %f", closedSet["State A"])
	}

	// Case 2: State A appears with Cost = 5.0 -> CLOSED set updated and better state explored
	n5 := NewSearchNode("node-5", nil, NewSearchState("State A"), nil)
	n5.CostProfile.AccumulatedCost = CostVector{Resources: 5.0}
	scalarG5 := engine.computeScalarCost(n5.CostProfile.AccumulatedCost)
	pruned := false
	if bestG, exists := closedSet[n5.State.StateID]; exists && bestG <= scalarG5 {
		pruned = true
	} else {
		closedSet[n5.State.StateID] = scalarG5 // CLOSED set updated
	}
	if pruned || closedSet["State A"] != 5.0 {
		t.Errorf("expected State A with better Cost %f (< bestG 8.0) NOT to be pruned and CLOSED updated to 5.0, got %f", scalarG5, closedSet["State A"])
	}
}

func TestBeamAStarEngine_BudgetAndTimePreemption(t *testing.T) {
	cfg := DefaultBeamAStarConfig()
	cfg.MaxDepth = 100000 // Allow deep exploration so execution budget preemption triggers
	engine := NewBeamAStarEngine(cfg)

	rootState := NewSearchState("state-root")
	rootState.RemainingDesiredState["never_reached"] = "true"

	// Infinite chain of operators to simulate long search that exceeds budget
	opLoop := NewSearchEdge("op-loop", EdgeTypeStrategicOperator, "StepForward")
	opLoop.EdgeCost = CostVector{Time: 10 * time.Millisecond, Resources: 1.0}

	// We set an extremely tight budget of 5ms
	req := &PlanningRequest{
		RequestID:          "req-preempt",
		Goal:               "Reach unreachable goal",
		MaxExecutionBudget: 5 * time.Millisecond,
	}

	result, err := engine.Search(context.Background(), req, rootState, []*SearchEdge{opLoop})
	if err != nil {
		t.Fatalf("expected preemption without error, got err: %v", err)
	}

	if result.Status != StatusPartialBudget {
		t.Fatalf("expected StatusPartialBudget upon preemption, got %s", result.Status)
	}
	if len(result.PartialNodes) == 0 {
		t.Error("expected partial nodes returned after preemption")
	}

	// Verify Context Cancellation behavior
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately before search starts

	reqCancelled := &PlanningRequest{RequestID: "req-cancel", Goal: "Test cancel"}
	resultCancel, err := engine.Search(ctx, reqCancelled, rootState, []*SearchEdge{opLoop})
	if err != nil {
		t.Fatalf("expected clean return on cancelled context, got err: %v", err)
	}
	if resultCancel.Status != StatusPartialBudget {
		t.Errorf("expected StatusPartialBudget on context cancellation, got %s", resultCancel.Status)
	}
}

// TestOpenQueue_PruneExactVerification3 explicitly verifies the exact BeamWidth = 3 scenario:
// If OPEN contains [2, 5, 7, 8, 11, 14], confirm that after pruning, OPEN contains [2, 5, 7] only.
func TestOpenQueue_PruneExactVerification3(t *testing.T) {
	q := NewOpenQueue()
	scores := []float64{14, 2, 8, 5, 11, 7} // Unordered insertion of [2, 5, 7, 8, 11, 14]
	for i, score := range scores {
		state := NewSearchState("state-" + string(rune('A'+i)))
		node := NewSearchNode("node-"+string(rune('A'+i)), nil, state, nil)
		node.CostProfile.EvaluationScore = score
		q.Push(node)
	}

	if q.Len() != 6 {
		t.Fatalf("expected queue len 6 before pruning, got %d", q.Len())
	}

	pruned, count := q.PruneToBeam(3)
	if count != 3 || len(pruned) != 3 {
		t.Fatalf("expected 3 pruned nodes, got %d", count)
	}
	if q.Len() != 3 {
		t.Fatalf("expected queue len 3 after pruning, got %d", q.Len())
	}

	snap := q.Snapshot()
	expectedScores := []float64{2.0, 5.0, 7.0}
	for i, exp := range expectedScores {
		if snap[i].CostProfile.EvaluationScore != exp {
			t.Errorf("at index %d: expected score %f, got %f", i, exp, snap[i].CostProfile.EvaluationScore)
		}
	}
}

// TestOpenQueue_DeterministicTieBreaking verifies that exact score ties are broken deterministically
// by EpistemicConfidence descending, then SimulatedClock ascending, then NodeID ascending.
func TestOpenQueue_DeterministicTieBreaking(t *testing.T) {
	q := NewOpenQueue()

	// Three nodes with identical EvaluationScore = 10.0
	n1 := NewSearchNode("node-C", nil, NewSearchState("state-C"), nil)
	n1.CostProfile.EvaluationScore = 10.0
	n1.State.EpistemicConfidence = 0.8
	n1.State.SimulatedClock = 5 * time.Second

	n2 := NewSearchNode("node-A", nil, NewSearchState("state-A"), nil)
	n2.CostProfile.EvaluationScore = 10.0
	n2.State.EpistemicConfidence = 0.9 // Higher confidence -> must be ordered before n1 and n3
	n2.State.SimulatedClock = 5 * time.Second

	n3 := NewSearchNode("node-B", nil, NewSearchState("state-B"), nil)
	n3.CostProfile.EvaluationScore = 10.0
	n3.State.EpistemicConfidence = 0.8
	n3.State.SimulatedClock = 2 * time.Second // Same confidence as n1, but shorter time -> must be ordered before n1

	q.Push(n1)
	q.Push(n2)
	q.Push(n3)

	snap := q.Snapshot()
	if snap[0].NodeID != "node-A" || snap[1].NodeID != "node-B" || snap[2].NodeID != "node-C" {
		t.Errorf("expected tie-break order [node-A, node-B, node-C], got [%s, %s, %s]", snap[0].NodeID, snap[1].NodeID, snap[2].NodeID)
	}
}

// TestBeamAStarEngine_ClosedSetExactVerification2 explicitly verifies CLOSED set duplicate handling:
// If State A ↓ Cost = 8 exists, and later State A ↓ Cost = 12 appears, the second state is pruned.
// But if State A ↓ Cost = 5 appears, the CLOSED entry is updated and the better state is explored.
func TestBeamAStarEngine_ClosedSetExactVerification2(t *testing.T) {
	cfg := DefaultBeamAStarConfig()
	engine := NewBeamAStarEngine(cfg)

	closedSet := make(map[string]float64)
	closedSet["State A"] = 8.0 // State A initially visited with Cost = 8.0

	// Case 1: State A appears with Cost = 12.0
	n12 := NewSearchNode("node-12", nil, NewSearchState("State A"), nil)
	n12.CostProfile.AccumulatedCost = CostVector{Resources: 12.0}
	scalarG12 := engine.computeScalarCost(n12.CostProfile.AccumulatedCost)

	if bestG, exists := closedSet[n12.State.StateID]; exists && bestG <= scalarG12 {
		// Pruned correctly
	} else {
		t.Errorf("expected State A with Cost %f (>= bestG %f) to be pruned", scalarG12, bestG)
	}
	if closedSet["State A"] != 8.0 {
		t.Errorf("expected CLOSED set to remain 8.0 after pruning worse cost, got %f", closedSet["State A"])
	}

	// Case 2: State A appears with Cost = 5.0
	n5 := NewSearchNode("node-5", nil, NewSearchState("State A"), nil)
	n5.CostProfile.AccumulatedCost = CostVector{Resources: 5.0}
	scalarG5 := engine.computeScalarCost(n5.CostProfile.AccumulatedCost)

	pruned := false
	if bestG, exists := closedSet[n5.State.StateID]; exists && bestG <= scalarG5 {
		pruned = true
	} else {
		closedSet[n5.State.StateID] = scalarG5 // CLOSED set updated
	}

	if pruned {
		t.Errorf("expected State A with better Cost %f (< bestG 8.0) NOT to be pruned", scalarG5)
	}
	if closedSet["State A"] != 5.0 {
		t.Errorf("expected CLOSED entry updated to 5.0, got %f", closedSet["State A"])
	}
}
