package planning

import (
	"context"
	"reflect"
	"testing"
	"time"

	"idun/intelligence/reasoning"
)

func buildTestProfile4C() *PlanningPolicyProfile {
	return &PlanningPolicyProfile{
		ProfileID: "prof-stage4c",
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

// Test A — Structured Communicative Goal
func TestStage4C_HTN_StructuredCommunicativeGoal(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4C()
	htn := NewHTNSpecialist("TACTICAL")

	semGoal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
		},
	}

	req := &PlanningRequest{
		RequestID:    "req-comm-1",
		Goal:         "Deliberately incorrect diagnostic string that should be ignored",
		ResolvedGoal: semGoal,
		Domain:       "Dialogue",
	}

	_, subgoals, edges, err := htn.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected HTN error: %v", err)
	}
	if len(subgoals) == 0 || len(edges) == 0 {
		t.Fatalf("expected subgoals and edges to be generated, got %d subgoals, %d edges", len(subgoals), len(edges))
	}

	// Verify structured goal wins and diagnostic string is ignored
	for i, sg := range subgoals {
		if stringsContains(sg.Title, "Deliberately incorrect") {
			t.Errorf("subgoal[%d] title incorrectly leaked diagnostic string: %q", i, sg.Title)
		}
		if sg.Parameters["goal_kind"] != string(reasoning.GoalKindCommunicative) {
			t.Errorf("subgoal[%d] expected goal_kind %s, got %s", i, reasoning.GoalKindCommunicative, sg.Parameters["goal_kind"])
		}
		if sg.Parameters["intent"] != "greet_user" || sg.Parameters["target"] != "user" {
			t.Errorf("subgoal[%d] unexpected parameters: %+v", i, sg.Parameters)
		}
		if sg.Parameters["desired_state"] != "acknowledged=true" {
			t.Errorf("subgoal[%d] expected desired_state acknowledged=true, got %s", i, sg.Parameters["desired_state"])
		}
	}

	// Verify specific communicative phase titles
	expectedPhase2Title := "Produce an approved acknowledgement of the user's greeting."
	if len(subgoals) > 2 && subgoals[2].Title != expectedPhase2Title {
		t.Errorf("expected phase 2 title %q, got %q", expectedPhase2Title, subgoals[2].Title)
	}
}

// Test B — Structured Tool Goal
func TestStage4C_HTN_StructuredToolGoal(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4C()
	htn := NewHTNSpecialist("TACTICAL")

	semGoal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindToolExecution,
		Intent: "set_alarm",
		Target: "alarm_service",
		DesiredState: map[string]string{
			"exists": "true",
			"time":   "07:00",
		},
	}

	req := &PlanningRequest{
		RequestID:    "req-tool-1",
		Goal:         "Derived symbolic conclusion for intent set_alarm",
		ResolvedGoal: semGoal,
		Domain:       "Automation",
	}

	_, subgoals, _, err := htn.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected HTN error: %v", err)
	}

	for i, sg := range subgoals {
		if sg.Parameters["desired_state"] != "exists=true, time=07:00" {
			t.Errorf("subgoal[%d] expected deterministic desired_state 'exists=true, time=07:00', got %q", i, sg.Parameters["desired_state"])
		}
		if sg.Parameters["intent"] != "set_alarm" || sg.Parameters["target"] != "alarm_service" {
			t.Errorf("subgoal[%d] mismatch parameters: %+v", i, sg.Parameters)
		}
	}
}

// Test C — Constraints
func TestStage4C_HTN_Constraints(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4C()
	htn := NewHTNSpecialist("TACTICAL")

	semGoal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindToolExecution,
		Intent: "send_email",
		Target: "email_service",
		DesiredState: map[string]string{
			"sent": "true",
		},
		Constraints: map[string]string{
			"max_retries": "3",
			"timeout":     "10s",
		},
	}

	req := &PlanningRequest{
		RequestID:    "req-const-1",
		Goal:         "Send email",
		ResolvedGoal: semGoal,
		Domain:       "Messaging",
	}

	_, subgoals, _, err := htn.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected HTN error: %v", err)
	}

	for i, sg := range subgoals {
		if sg.Parameters["constraints"] != "max_retries=3, timeout=10s" {
			t.Errorf("subgoal[%d] expected deterministic constraints 'max_retries=3, timeout=10s', got %q", i, sg.Parameters["constraints"])
		}
	}
}

// Test D — Determinism
func TestStage4C_HTN_Determinism(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4C()
	htn := NewHTNSpecialist("TACTICAL")

	semGoal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindStateChange,
		Intent: "reboot_node",
		Target: "node_cluster_1",
		DesiredState: map[string]string{
			"status": "online",
			"health": "ok",
		},
		Constraints: map[string]string{
			"maintenance_window": "active",
		},
	}

	req := &PlanningRequest{
		RequestID:    "req-det-1",
		Goal:         "Reboot node cluster",
		ResolvedGoal: semGoal,
		Domain:       "Infrastructure",
	}

	_, subgoals1, edges1, err1 := htn.Contribute(ctx, req, nil, profile)
	if err1 != nil {
		t.Fatalf("run 1 failed: %v", err1)
	}

	// Sleep briefly to ensure any timestamp-based logic would diverge if not strictly deterministic
	time.Sleep(10 * time.Millisecond)

	_, subgoals2, edges2, err2 := htn.Contribute(ctx, req, nil, profile)
	if err2 != nil {
		t.Fatalf("run 2 failed: %v", err2)
	}

	if !reflect.DeepEqual(subgoals1, subgoals2) {
		t.Errorf("HTN subgoals are not byte-for-byte deterministic across identical runs!\nRun 1: %+v\nRun 2: %+v", subgoals1, subgoals2)
	}
	if !reflect.DeepEqual(edges1, edges2) {
		t.Errorf("HTN edges are not byte-for-byte deterministic across identical runs!\nRun 1: %+v\nRun 2: %+v", edges1, edges2)
	}
}

// Test E — Legacy Compatibility
func TestStage4C_HTN_LegacyCompatibility(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4C()
	htn := NewHTNSpecialist("TACTICAL")

	// ResolvedGoal = nil
	req := &PlanningRequest{
		RequestID:    "req-legacy-1",
		Goal:         "Format disk drive",
		ResolvedGoal: nil,
		Domain:       "Storage",
	}

	_, subgoals, edges, err := htn.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected HTN error: %v", err)
	}

	if len(subgoals) != 4 || len(edges) != 3 {
		t.Fatalf("expected 4 subgoals, 3 edges in legacy fallback, got %d and %d", len(subgoals), len(edges))
	}

	expectedPhases := []string{
		"Analyze scope and preconditions",
		"Formulate core structural design",
		"Execute primary task operations",
		"Verify outcome invariants and postconditions",
	}

	for i, sg := range subgoals {
		expectedTitle := expectedPhases[i] + " for [Format disk drive]"
		if sg.Title != expectedTitle {
			t.Errorf("subgoal[%d] expected legacy title %q, got %q", i, expectedTitle, sg.Title)
		}
		if sg.Parameters["goal_kind"] != "" {
			t.Errorf("subgoal[%d] unexpected structured parameter in legacy mode: %+v", i, sg.Parameters)
		}
	}
}

// Test F — Semantic Ownership
func TestStage4C_HTN_SemanticOwnership(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4C()
	htn := NewHTNSpecialist("TACTICAL")

	semGoal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindInformationRetrieval,
		Intent: "query_logs",
		Target: "auth_service",
		DesiredState: map[string]string{
			"records_found": "true",
		},
		Constraints: map[string]string{
			"limit": "100",
		},
		OperationHint: "fast-query",
	}

	beforeCopy := semGoal.Clone()

	req := &PlanningRequest{
		RequestID:    "req-owner-1",
		Goal:         "Query auth service logs",
		ResolvedGoal: semGoal,
		Domain:       "Observability",
	}

	_, _, _, err := htn.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected HTN error: %v", err)
	}

	if !reflect.DeepEqual(semGoal, beforeCopy) {
		t.Fatalf("HTN illegally mutated req.ResolvedGoal!\nBefore: %+v\nAfter:  %+v", beforeCopy, semGoal)
	}
}

// Test G — Conflicting Legacy String
func TestStage4C_HTN_ConflictingLegacyString(t *testing.T) {
	ctx := context.Background()
	profile := buildTestProfile4C()
	htn := NewHTNSpecialist("TACTICAL")

	semGoal := &reasoning.SemanticGoal{
		Kind:   reasoning.GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
		},
	}

	req := &PlanningRequest{
		RequestID:    "req-conflict-1",
		Goal:         `Derived symbolic conclusion for intent "wrong_intent"`,
		ResolvedGoal: semGoal,
		Domain:       "Dialogue",
	}

	_, subgoals, _, err := htn.Contribute(ctx, req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected HTN error: %v", err)
	}

	for i, sg := range subgoals {
		if stringsContains(sg.Title, "wrong_intent") {
			t.Errorf("subgoal[%d] title leaked wrong_intent from diagnostic goal string: %q", i, sg.Title)
		}
		if sg.Parameters["intent"] != "greet_user" {
			t.Errorf("subgoal[%d] expected intent greet_user, got %q", i, sg.Parameters["intent"])
		}
	}
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && (s+substr[:1] != s && indexOfSubstr(s, substr) >= 0)))
}

func indexOfSubstr(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
