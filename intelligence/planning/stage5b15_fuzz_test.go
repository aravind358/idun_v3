package planning

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// FuzzPlanningResultParsing fuzzes deserialization and structural validation of PlanningResult payloads.
func FuzzPlanningResultParsing(f *testing.F) {
	validRes := PlanningResult{
		ResultID:      "res-1",
		RequestID:     "req-1",
		ResultStatus:  ResultSuccess,
		PrimaryPlanID: "plan-1",
		Plans: []*Plan{
			{
				PlanID:             "plan-1",
				SchemaVersion:      SchemaVersion2_0_0,
				StrategySnapshotID: "snap-1",
				Goal:               "Goal 1",
				EstimatedCost:      10.0,
				Subgoals:           []Subgoal{{SubgoalID: "sg-1", Title: "Title 1"}},
			},
		},
	}
	rawValid, _ := json.Marshal(validRes)
	f.Add(string(rawValid))
	f.Add("")
	f.Add("not-json-data")
	f.Add(`{"result_id":"res-2","request_id":""}`)
	f.Add(`{"result_id":"res-3","plans":[{"plan_id":"","estimated_cost":-100}]}`)

	f.Fuzz(func(t *testing.T, data string) {
		var res PlanningResult
		if err := json.Unmarshal([]byte(data), &res); err == nil {
			_ = res.Validate()
		}
	})
}

// FuzzSearchStateCloning fuzzes deep-copy isolation and validation of SearchState across arbitrary map entries and vectors.
func FuzzSearchStateCloning(f *testing.F) {
	f.Add("state-fuzz-1", 0.85, uint64(100), "goal_key", "goal_val", "constraint_key", "constraint_val", 0.05)
	f.Add("", 0.0, uint64(0), "", "", "", "", -1.0)
	f.Add("state-fuzz-2", 1.5, uint64(999999), "k1", "v1", "c1", "cv1", 2.0)

	f.Fuzz(func(t *testing.T, stateID string, confidence float64, clock uint64, gKey, gVal, cKey, cVal string, risk float64) {
		state := NewSearchState(stateID)
		state.EpistemicConfidence = confidence
		state.SimulatedClock = time.Duration(clock)
		state.AccumulatedRisk = risk
		if gKey != "" {
			state.RemainingDesiredState[gKey] = gVal
		}
		if cKey != "" {
			state.ActiveConstraints[cKey] = cVal
		}

		clone := state.Clone()
		if clone == nil {
			t.Fatalf("SearchState.Clone() returned nil for state: %v", stateID)
		}
		if clone == state {
			t.Fatalf("SearchState.Clone() returned shallow pointer copy")
		}

		// Verify deep copy mutation isolation
		if gKey != "" {
			clone.RemainingDesiredState[gKey] = gVal + "-mutated"
			if state.RemainingDesiredState[gKey] == clone.RemainingDesiredState[gKey] {
				t.Errorf("RemainingDesiredState map shared between state and clone")
			}
		}
		_ = clone.Validate()
	})
}

// FuzzBeamAStarEngine fuzzes Beam A* execution under arbitrary ceiling, weight, and risk configurations.
func FuzzBeamAStarEngine(f *testing.F) {
	f.Add(int(15), int(50), 1.0, 10.0, 2.0, 0.80, "op-1", 0.05, string(ReversibilityHighCost))
	f.Add(int(0), int(0), -1.0, -5.0, 0.0, -0.1, "", -1.0, string(ReversibilityCritical))
	f.Add(int(1000), int(5), 10.0, 1.0, 5.0, 1.5, "op-huge", 0.5, string(ReversibilityTrivial))

	f.Fuzz(func(t *testing.T, beamWidth, maxDepth int, lambda, gamma, beta, riskCeiling float64, opID string, riskDelta float64, rev string) {
		cfg := BeamAStarConfig{
			BeamWidth:      beamWidth,
			MaxDepth:       maxDepth,
			Lambda:         lambda,
			Gamma:          gamma,
			Beta:           beta,
			MaxRiskCeiling: riskCeiling,
		}
		engine := NewBeamAStarEngine(cfg)
		req, _ := NewPlanningRequestBuilder().
			WithRequestID("req-fuzz-beam").
			WithGoal("Fuzz Goal").
			WithDomain("General").
			WithTargetDepth(DepthStrategic).
			Build()

		rootState := NewSearchState("root-fuzz")
		rootState.EpistemicConfidence = 0.90
		rootState.RemainingDesiredState["goal"] = "Fuzz Goal"

		ops := make([]*SearchEdge, 0, 1)
		if opID != "" {
			op := NewSearchEdge(opID, EdgeTypeStrategicOperator, "Fuzz Operator")
			op.RiskDelta = riskDelta
			if Reversibility(rev).IsValid() {
				op.Reversibility = Reversibility(rev)
			} else {
				op.Reversibility = ReversibilityHighCost
			}
			op.Postconditions["goal"] = "Fuzz Goal"
			ops = append(ops, op)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, _ = engine.Search(ctx, req, rootState, ops)
	})
}

// FuzzTreeSearchSpecialist fuzzes TreeSearchSpecialist contribution across arbitrary goals, domains, and operator costs.
func FuzzTreeSearchSpecialist(f *testing.F) {
	f.Add("req-tree-1", "Strategic Expansion", "General", string(DepthStrategic), float64(10.0), 0.05)
	f.Add("", "", "", "INVALID_DEPTH", float64(-100.0), -1.5)
	f.Add("req-tree-2", "Contingency Check", "Robotics", string(DepthTactical), float64(0.0), 2.0)

	f.Fuzz(func(t *testing.T, reqID, goal, domain, depth string, opCost, opRisk float64) {
		s := NewTreeSearchSpecialist("STRATEGIC")
		req := &PlanningRequest{
			RequestID:   reqID,
			Goal:        goal,
			Domain:      domain,
			TargetDepth: PlanningDepth(depth),
		}

		graph := &DependencyGraphSnapshot{Nodes: make(map[string]string), Edges: []DependencyEdge{}}
		profile := DefaultPlanningPolicyProfile()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, _, _, _ = s.Contribute(ctx, req, graph, profile)
	})
}

// FuzzPlanConversionAndRollbackExtraction fuzzes ConvertPathToPlan across arbitrary multi-node search paths and reversibility levels.
func FuzzPlanConversionAndRollbackExtraction(f *testing.F) {
	f.Add("node-0", "node-1", "op-1", string(ReversibilityHighCost), 0.05, float64(15.0))
	f.Add("", "", "", string(ReversibilityCritical), -0.5, -10.0)
	f.Add("root", "child", "branch-op", string(ReversibilityTrivial), 0.9, float64(0.0))

	f.Fuzz(func(t *testing.T, rootID, childID, opID, rev string, risk float64, cost float64) {
		s := NewTreeSearchSpecialist("STRATEGIC")
		req, _ := NewPlanningRequestBuilder().
			WithRequestID("req-fuzz-conv").
			WithGoal("Fuzz conversion").
			WithDomain("General").
			WithTargetDepth(DepthStrategic).
			Build()

		rootState := NewSearchState("state-root")
		rootState.EpistemicConfidence = 0.95
		rootNode := NewSearchNode(rootID, nil, rootState, nil)

		path := []*SearchNode{rootNode}
		if childID != "" && opID != "" {
			childState := rootState.Clone()
			childState.StateID = "state-child"
			op := NewSearchEdge(opID, EdgeTypeStrategicOperator, "Fuzz Op")
			op.RiskDelta = risk
			op.EdgeCost = CostVector{Time: 1 * time.Second, Resources: cost}
			if Reversibility(rev).IsValid() {
				op.Reversibility = Reversibility(rev)
			} else {
				op.Reversibility = ReversibilityHighCost
			}
			childNode := NewSearchNode(childID, rootNode, childState, op)
			path = append(path, childNode)
		}

		plan, err := s.ConvertPathToPlan(path, req, "snap-fuzz", "trace-fuzz", 0)
		if err == nil && plan != nil {
			_ = plan.Validate()
		}
	})
}

// FuzzPlannerRouting fuzzes planner identification and routing helpers against arbitrary strings.
func FuzzPlannerRouting(f *testing.F) {
	f.Add("HTNDecompositionSpecialist", "General")
	f.Add("GOAPActionSpecialist", "Robotics")
	f.Add("MultiAlternativeTreeSearchSpecialist", "Strategic")
	f.Add("UnknownSpecialist", "UnknownDomain")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, specName, domain string) {
		// Test routing check across all registered specialist types
		htn := NewHTNSpecialist("TACTICAL")
		goap := NewGOAPSpecialist("TACTICAL")
		tree := NewTreeSearchSpecialist("STRATEGIC")

		for _, dom := range htn.SupportedDomains() {
			if dom == domain {
				_ = htn.Name()
			}
		}
		for _, dom := range goap.SupportedDomains() {
			if dom == domain {
				_ = goap.Name()
			}
		}
		for _, dom := range tree.SupportedDomains() {
			if dom == domain {
				_ = tree.Name()
			}
		}
	})
}
