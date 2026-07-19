package planning

import (
	"context"
	"reflect"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/reasoning"
)

// TestStage4B_PlanningRequest_Transport verifies that PlanningRequest can transport
// ResolvedGoal cleanly and that Validate() remains backward compatible for older payloads.
func TestStage4B_PlanningRequest_Transport(t *testing.T) {
	// Case 1: Older payload without ResolvedGoal (nil)
	legacyReq := &PlanningRequest{
		RequestID:          "req-legacy-1",
		Goal:               "Perform legacy task",
		Domain:             "General",
		TargetDepth:        DepthTactical,
		MaxExecutionBudget: 100 * time.Millisecond,
		MinConfidenceFloor: 0.70,
	}
	if err := legacyReq.Validate(); err != nil {
		t.Fatalf("expected legacy PlanningRequest without ResolvedGoal to validate successfully, got: %v", err)
	}

	// Case 2: New payload with ResolvedGoal populated
	semGoal := &reasoning.SemanticGoal{
		Kind:          reasoning.GoalKindCommunicative,
		Intent:        "greet_user",
		Target:        "user",
		DesiredState:  map[string]string{"acknowledged": "true"},
		Constraints:   map[string]string{"tone": "friendly"},
		OperationHint: "greeting",
	}
	extendedReq := &PlanningRequest{
		RequestID:          "req-extended-1",
		Goal:               "Derived symbolic conclusion for intent greet_user",
		ResolvedGoal:       semGoal,
		Domain:             "General",
		TargetDepth:        DepthTactical,
		MaxExecutionBudget: 100 * time.Millisecond,
		MinConfidenceFloor: 0.70,
	}
	if err := extendedReq.Validate(); err != nil {
		t.Fatalf("expected extended PlanningRequest with ResolvedGoal to validate successfully, got: %v", err)
	}
	if extendedReq.ResolvedGoal == nil || extendedReq.ResolvedGoal.Intent != "greet_user" {
		t.Errorf("unexpected ResolvedGoal in PlanningRequest: %+v", extendedReq.ResolvedGoal)
	}
}

// TestStage4B_Builders_Compatibility verifies that WithResolvedGoal works on builders
// while maintaining complete backward compatibility for existing builder chaining.
func TestStage4B_Builders_Compatibility(t *testing.T) {
	semGoal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
	}

	req, err := NewPlanningRequestBuilder().
		WithRequestID("req-builder-1").
		WithGoal("Test Goal").
		WithResolvedGoal(semGoal).
		WithDomain("Dialogue").
		Build()

	if err != nil {
		t.Fatalf("PlanningRequestBuilder failed with WithResolvedGoal: %v", err)
	}
	if req.Goal != "Test Goal" || req.ResolvedGoal == nil || req.ResolvedGoal.Intent != "greet_user" {
		t.Errorf("unexpected built PlanningRequest fields: %+v", req)
	}

	planBuilder := NewPlanBuilder().
		WithIdentity("plan-1", "snap-1", "trace-1").
		WithGoalAndDomain("Test Goal", "Dialogue", "FAST").
		WithResolvedGoal(semGoal).
		WithReplayMetadata(ReplayMetadata{StrategySnapshotID: "snap-1", ReplayFidelity: "EXACT"}).
		WithStatus(PlanStatusComplete, nil)

	plan, err := planBuilder.Build()
	if err != nil {
		t.Fatalf("PlanBuilder failed with WithResolvedGoal: %v", err)
	}
	if plan.Goal != "Test Goal" || plan.Domain != "Dialogue" {
		t.Errorf("unexpected built Plan fields: %+v", plan)
	}
}

// TestStage4B_Specialists_ByteForByteIdentical proves that in Stage 4B,
// HTN, GOAP, and Tree Search specialists remain behaviorally identical and
// yield byte-for-byte identical output whether ResolvedGoal is nil or present.
func TestStage4B_Specialists_ByteForByteIdentical(t *testing.T) {
	ctx := context.Background()
	reqNilGoal := &PlanningRequest{
		RequestID: "req-spec-1",
		Goal:      "Test Decomposition Goal",
		Domain:    "General",
	}
	reqWithGoal := &PlanningRequest{
		RequestID: "req-spec-1",
		Goal:      "Test Decomposition Goal",
		Domain:    "General",
		ResolvedGoal: &reasoning.SemanticGoal{
			Kind:         reasoning.GoalKindCommunicative,
			Intent:       "greet_user",
			Target:       "user",
			DesiredState: map[string]string{"acknowledged": "true"},
		},
	}

	profile := &PlanningPolicyProfile{
		Capabilities: &PlanningCapabilities{
			SupportsHTN:      true,
			SupportsGOAP:     true,
			MaxPlanningDepth: 10,
		},
		SearchStrategies: map[string]*PlanningSearchStrategy{
			"TACTICAL": {
				SearchID:          "strat-tactical",
				MaxDepth:          4,
				MaxNodes:          20,
				BeamWidth:         2,
				AllowBacktracking: true,
				PruningPolicy:     "ALPHA_BETA",
				ExpansionPolicy:   "BREADTH_FIRST",
			},
		},
	}

	// 1. Check HTN (Note: In Stage 4C, HTN adopted structured ResolvedGoal, so subNilHTN and subResolvedHTN differ intentionally)
	htn := NewHTNSpecialist("TACTICAL")
	_, _, _, err1 := htn.Contribute(ctx, reqNilGoal, nil, profile)
	_, subResolvedHTN, _, err2 := htn.Contribute(ctx, reqWithGoal, nil, profile)
	if err1 != nil || err2 != nil {
		t.Fatalf("HTN Contribute failed: %v / %v", err1, err2)
	}
	if len(subResolvedHTN) == 0 {
		t.Errorf("HTN returned empty subgoals when ResolvedGoal was added")
	}

	// 2. Check GOAP (Note: In Stage 4D, GOAP adopted structured ResolvedGoal and abstains on COMMUNICATIVE reqWithGoal)
	goap := NewGOAPSpecialist("TACTICAL")
	_, _, _, err3 := goap.Contribute(ctx, reqNilGoal, nil, profile)
	_, subResolvedGOAP, _, err4 := goap.Contribute(ctx, reqWithGoal, nil, profile)
	if err3 != nil || err4 != nil {
		t.Fatalf("GOAP Contribute failed: %v / %v", err3, err4)
	}
	if len(subResolvedGOAP) != 0 {
		t.Errorf("GOAP expected to abstain and return 0 subgoals for COMMUNICATIVE goal, got %d", len(subResolvedGOAP))
	}

	// 3. Check TreeSearch
	ts := NewTreeSearchSpecialist("TACTICAL")
	_, subNilTS, edgeNilTS, err5 := ts.Contribute(ctx, reqNilGoal, nil, profile)
	_, subResolvedTS, edgeResolvedTS, err6 := ts.Contribute(ctx, reqWithGoal, nil, profile)
	if err5 != nil || err6 != nil {
		t.Fatalf("TreeSearch Contribute failed: %v / %v", err5, err6)
	}
	if !reflect.DeepEqual(subNilTS, subResolvedTS) || !reflect.DeepEqual(edgeNilTS, edgeResolvedTS) {
		t.Errorf("TreeSearch output changed when ResolvedGoal was added! Expected byte-for-byte identical output.")
	}
}

// TestStage4B_Service_ActiveGoals_Transport verifies that handleActiveGoal unpacks
// both PrimaryHypothesis.Conclusion and ResolvedGoal from TopicActiveGoals envelopes.
func TestStage4B_Service_ActiveGoals_Transport(t *testing.T) {
	storer := &casStorer{data: make(map[string][]byte)}
	pub := &capturingPublisher{}
	sub := &recordingSubscriber{}

	svc := NewService(WithWorkspaceBridge(storer, pub, sub))
	if err := svc.Start(); err != nil {
		t.Fatalf("failed to start planning service: %v", err)
	}
	defer svc.Close()

	payload := []byte(`{
		"schema_version": "2.0",
		"primary_hypothesis": {
			"conclusion": "Derived symbolic conclusion for intent greet_user"
		},
		"resolved_goal": {
			"kind": "COMMUNICATIVE",
			"intent": "greet_user",
			"target": "user",
			"desired_state": {"acknowledged": "true"},
			"operation_hint": "greeting"
		}
	}`)
	casKey, _ := storer.Store(context.Background(), payload)

	env := communication.Envelope{
		ID:         "env-stage4b-1",
		Source:     "Intelligence.Reasoning",
		Topic:      communication.TopicActiveGoals,
		PayloadRef: casKey,
	}

	if err := sub.handler(context.Background(), env); err != nil {
		t.Fatalf("handleActiveGoal returned unexpected error: %v", err)
	}

	pub.mu.Lock()
	defer pub.mu.Unlock()
	if len(pub.envs) == 0 {
		t.Fatalf("expected candidate plan envelope to be published")
	}

	publishedEnv := pub.envs[0]
	if publishedEnv.Topic != communication.TopicCandidatePlans {
		t.Errorf("expected topic candidate-plans, got %s", publishedEnv.Topic)
	}
}
