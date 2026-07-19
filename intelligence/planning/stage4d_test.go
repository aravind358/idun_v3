package planning

import (
	"context"
	"reflect"
	"testing"
	"time"

	"idun/intelligence/reasoning"
)

func buildTestProfile4D() *PlanningPolicyProfile {
	return &PlanningPolicyProfile{
		ProfileID: "prof-stage4d",
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
}

// Test A — Structured DesiredState
func TestStage4D_GOAP_StructuredDesiredState(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4D()
	goap := NewGOAPSpecialist("TACTICAL")

	goal1 := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindToolExecution,
		Intent: "set_alarm",
		Target: "alarm_service",
		DesiredState: map[string]string{
			"exists": "true",
			"time":   "07:00",
		},
	}

	goal2 := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindToolExecution,
		Intent: "set_alarm",
		Target: "alarm_service",
		DesiredState: map[string]string{
			"exists": "true",
			"time":   "08:00",
		},
	}

	req1 := &PlanningRequest{
		RequestID:    "req-goap-ds1",
		Goal:         "Set alarm to 07:00",
		ResolvedGoal: goal1,
		Domain:       "Automation",
	}

	req2 := &PlanningRequest{
		RequestID:    "req-goap-ds2",
		Goal:         "Set alarm to 08:00",
		ResolvedGoal: goal2,
		Domain:       "Automation",
	}

	_, subgoals1, _, err1 := goap.Contribute(ctx, req1, nil, profile)
	_, subgoals2, _, err2 := goap.Contribute(ctx, req2, nil, profile)
	if err1 != nil || err2 != nil {
		t.Fatalf("GOAP Contribute failed: %v / %v", err1, err2)
	}

	if len(subgoals1) == 0 || len(subgoals2) == 0 {
		t.Fatalf("expected subgoals, got %d / %d", len(subgoals1), len(subgoals2))
	}

	ds1 := subgoals1[len(subgoals1)-1].Parameters["desired_state"]
	ds2 := subgoals2[len(subgoals2)-1].Parameters["desired_state"]

	if ds1 == ds2 {
		t.Fatalf("expected different target state conditions when DesiredState changes, got identical: %q", ds1)
	}
	if ds1 != "exists=true, time=07:00" || ds2 != "exists=true, time=08:00" {
		t.Errorf("unexpected desired states: ds1=%q, ds2=%q", ds1, ds2)
	}
}

// Test B — Conflicting Legacy Goal
func TestStage4D_GOAP_ConflictingLegacyGoal(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4D()
	goap := NewGOAPSpecialist("TACTICAL")

	goal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindToolExecution,
		Intent: "set_alarm",
		Target: "alarm_service",
		DesiredState: map[string]string{
			"exists": "true",
			"time":   "07:00",
		},
	}

	req := &PlanningRequest{
		RequestID:    "req-goap-conflict",
		Goal:         "Derived symbolic conclusion for wrong_intent",
		ResolvedGoal: goal,
		Domain:       "Automation",
	}

	_, subgoals, _, err := goap.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected GOAP error: %v", err)
	}

	for i, sg := range subgoals {
		if stringsContains(sg.Title, "wrong_intent") || stringsContains(sg.Parameters["postcondition"], "wrong_intent") {
			t.Errorf("subgoal[%d] leaked legacy wrong_intent: title=%q, post=%q", i, sg.Title, sg.Parameters["postcondition"])
		}
		if sg.Parameters["desired_state"] != "exists=true, time=07:00" {
			t.Errorf("subgoal[%d] expected desired_state 'exists=true, time=07:00', got %q", i, sg.Parameters["desired_state"])
		}
	}
}

// Test C — Constraints
func TestStage4D_GOAP_Constraints(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4D()
	goap := NewGOAPSpecialist("TACTICAL")

	goal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindInformationRetrieval,
		Intent: "diagnose_performance",
		Target: "local_host",
		DesiredState: map[string]string{
			"root_cause_identified": "true",
		},
		Constraints: map[string]string{
			"read_only": "true",
		},
	}

	req := &PlanningRequest{
		RequestID:    "req-goap-const",
		Goal:         "Diagnose performance",
		ResolvedGoal: goal,
		Domain:       "Observability",
	}

	_, subgoals, _, err := goap.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected GOAP error: %v", err)
	}

	for i, sg := range subgoals {
		if sg.Parameters["constraints"] != "read_only=true" {
			t.Errorf("subgoal[%d] expected constraints 'read_only=true', got %q", i, sg.Parameters["constraints"])
		}
		if sg.Parameters["desired_state"] != "root_cause_identified=true" {
			t.Errorf("subgoal[%d] expected desired_state 'root_cause_identified=true', got %q", i, sg.Parameters["desired_state"])
		}
		if sg.Parameters["constraint_enforcement"] != "enforced_read_only" {
			t.Errorf("subgoal[%d] expected constraint_enforcement 'enforced_read_only', got %q", i, sg.Parameters["constraint_enforcement"])
		}
		// Verify mutating transitions are rejected under read_only=true
		if stringsContains(sg.Title, "Configure operator state transitions") || stringsContains(sg.Title, "Execute state transition establishing") {
			t.Errorf("subgoal[%d] illegally selected mutating operator transition under read_only=true: %q", i, sg.Title)
		}
	}
}

// Test D — Determinism
func TestStage4D_GOAP_Determinism(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4D()
	goap := NewGOAPSpecialist("TACTICAL")

	goal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindStateChange,
		Intent: "scale_cluster",
		Target: "k8s_pool_1",
		DesiredState: map[string]string{
			"replicas": "5",
			"status":   "healthy",
		},
		Constraints: map[string]string{
			"max_surge": "2",
		},
	}

	req := &PlanningRequest{
		RequestID:    "req-goap-det",
		Goal:         "Scale k8s cluster",
		ResolvedGoal: goal,
		Domain:       "Infrastructure",
	}

	_, sub1, edges1, err1 := goap.Contribute(ctx, req, nil, profile)
	if err1 != nil {
		t.Fatalf("run 1 failed: %v", err1)
	}

	time.Sleep(10 * time.Millisecond)

	_, sub2, edges2, err2 := goap.Contribute(ctx, req, nil, profile)
	if err2 != nil {
		t.Fatalf("run 2 failed: %v", err2)
	}

	if !reflect.DeepEqual(sub1, sub2) {
		t.Errorf("GOAP subgoals not deterministic across identical runs!\nRun 1: %+v\nRun 2: %+v", sub1, sub2)
	}
	if !reflect.DeepEqual(edges1, edges2) {
		t.Errorf("GOAP edges not deterministic across identical runs!\nRun 1: %+v\nRun 2: %+v", edges1, edges2)
	}
}

// Test E — Legacy Compatibility
func TestStage4D_GOAP_LegacyCompatibility(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4D()
	goap := NewGOAPSpecialist("TACTICAL")

	req := &PlanningRequest{
		RequestID:    "req-goap-legacy",
		Goal:         "Migrate database schema",
		ResolvedGoal: nil,
		Domain:       "Database",
	}

	_, subgoals, edges, err := goap.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected GOAP error: %v", err)
	}

	if len(subgoals) != 4 || len(edges) != 3 {
		t.Fatalf("expected 4 subgoals, 3 edges in legacy fallback, got %d and %d", len(subgoals), len(edges))
	}

	expectedActions := []string{
		"GOAP Action: Acquire resources and environment lock",
		"GOAP Action: Configure state parameters",
		"GOAP Action: Execute target goal transition",
		"GOAP Action: Verify state persistence and cleanup",
	}

	for i, sg := range subgoals {
		if sg.Title != expectedActions[i] {
			t.Errorf("subgoal[%d] expected title %q, got %q", i, expectedActions[i], sg.Title)
		}
		if sg.Parameters["goal_kind"] != "" {
			t.Errorf("subgoal[%d] unexpected structured parameter in legacy mode: %+v", i, sg.Parameters)
		}
	}
	if subgoals[2].Parameters["postcondition"] != "Migrate database schema satisfied" {
		t.Errorf("subgoal[2] expected legacy postcondition, got %q", subgoals[2].Parameters["postcondition"])
	}
}

// Test F — Semantic Ownership
func TestStage4D_GOAP_SemanticOwnership(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4D()
	goap := NewGOAPSpecialist("TACTICAL")

	goal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindToolExecution,
		Intent: "restart_service",
		Target: "nginx_gateway",
		DesiredState: map[string]string{
			"running": "true",
		},
		Constraints: map[string]string{
			"graceful": "true",
		},
	}

	beforeCopy := goal.Clone()

	req := &PlanningRequest{
		RequestID:    "req-goap-owner",
		Goal:         "Restart nginx",
		ResolvedGoal: goal,
		Domain:       "Networking",
	}

	_, _, _, err := goap.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected GOAP error: %v", err)
	}

	if !reflect.DeepEqual(goal, beforeCopy) {
		t.Fatalf("GOAP illegally mutated req.ResolvedGoal!\nBefore: %+v\nAfter:  %+v", beforeCopy, goal)
	}
}

// Test G — DesiredState Dominance
func TestStage4D_GOAP_DesiredStateDominance(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4D()
	goap := NewGOAPSpecialist("TACTICAL")

	goalA := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindStateChange,
		Intent: "reboot_server",
		Target: "server-1",
		DesiredState: map[string]string{
			"status": "online",
		},
	}

	goalB := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindStateChange,
		Intent: "reboot_server",
		Target: "server-1",
		DesiredState: map[string]string{
			"status": "maintenance",
		},
	}

	reqA := &PlanningRequest{RequestID: "req-dom-a", Goal: "Reboot", ResolvedGoal: goalA, Domain: "Ops"}
	reqB := &PlanningRequest{RequestID: "req-dom-b", Goal: "Reboot", ResolvedGoal: goalB, Domain: "Ops"}

	_, subA, _, _ := goap.Contribute(ctx, reqA, nil, profile)
	_, subB, _, _ := goap.Contribute(ctx, reqB, nil, profile)

	if subA[len(subA)-1].Parameters["postcondition"] == subB[len(subB)-1].Parameters["postcondition"] {
		t.Fatalf("expected different postconditions when DesiredState changes from online to maintenance, got identical: %q", subA[len(subA)-1].Parameters["postcondition"])
	}
}

// Test H — Intent Independence
func TestStage4D_GOAP_IntentIndependence(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4D()
	goap := NewGOAPSpecialist("TACTICAL")

	goal1 := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindToolExecution,
		Intent: "set_alarm",
		Target: "alarm_service",
		DesiredState: map[string]string{
			"exists": "true",
			"time":   "07:00",
		},
		Constraints: map[string]string{
			"max_retries": "3",
		},
	}

	goal2 := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindToolExecution,
		Intent: "configure_alarm", // different intent label
		Target: "alarm_service",
		DesiredState: map[string]string{
			"exists": "true",
			"time":   "07:00",
		},
		Constraints: map[string]string{
			"max_retries": "3",
		},
	}

	req1 := &PlanningRequest{RequestID: "req-ind-1", Goal: "Alarm 1", ResolvedGoal: goal1, Domain: "Automation"}
	req2 := &PlanningRequest{RequestID: "req-ind-2", Goal: "Alarm 2", ResolvedGoal: goal2, Domain: "Automation"}

	_, sub1, _, _ := goap.Contribute(ctx, req1, nil, profile)
	_, sub2, _, _ := goap.Contribute(ctx, req2, nil, profile)

	for i := range sub1 {
		if sub1[i].Parameters["desired_state"] != sub2[i].Parameters["desired_state"] {
			t.Errorf("node[%d] desired_state mismatch: %q vs %q", i, sub1[i].Parameters["desired_state"], sub2[i].Parameters["desired_state"])
		}
		if sub1[i].Parameters["precondition"] != sub2[i].Parameters["precondition"] {
			t.Errorf("node[%d] precondition mismatch: %q vs %q", i, sub1[i].Parameters["precondition"], sub2[i].Parameters["precondition"])
		}
		if sub1[i].Parameters["postcondition"] != sub2[i].Parameters["postcondition"] {
			t.Errorf("node[%d] postcondition mismatch: %q vs %q", i, sub1[i].Parameters["postcondition"], sub2[i].Parameters["postcondition"])
		}
		if sub1[i].Parameters["constraints"] != sub2[i].Parameters["constraints"] {
			t.Errorf("node[%d] constraints mismatch: %q vs %q", i, sub1[i].Parameters["constraints"], sub2[i].Parameters["constraints"])
		}
		// Verify intent parameter differs correctly reflecting metadata difference
		if sub1[i].Parameters["intent"] == sub2[i].Parameters["intent"] {
			t.Errorf("expected intent metadata to differ between set_alarm and configure_alarm, both were %q", sub1[i].Parameters["intent"])
		}
	}
}
