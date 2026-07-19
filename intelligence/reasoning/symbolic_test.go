package reasoning

import (
	"context"
	"testing"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/understanding"
)

func TestSymbolicSpecialist_EvaluateWithFrame(t *testing.T) {
	specialist := NewSymbolicSpecialist()
	if specialist.ID() != StageS1SymbolicFast {
		t.Errorf("expected stage ID %s, got %s", StageS1SymbolicFast, specialist.ID())
	}

	frame := &understanding.SemanticFrame{
		PrimaryHypothesis: understanding.Hypothesis{
			Intent:               "authorize_user",
			CalibratedConfidence: 0.95,
			Slots: []understanding.Slot{
				{Name: "role", Value: "admin"},
			},
		},
	}

	memRecords := []memory.Record{
		{ID: "bel-role-admin", Type: "belief"},
	}

	hyps, err := specialist.Evaluate(context.Background(), communication.Envelope{ID: "env-s1"}, frame, memRecords)
	if err != nil {
		t.Fatalf("expected symbolic evaluation to succeed, got %v", err)
	}
	if len(hyps) != 1 {
		t.Fatalf("expected 1 hypothesis, got %d", len(hyps))
	}
	if err := hyps[0].Validate(); err != nil {
		t.Fatalf("generated hypothesis failed validation: %v", err)
	}
	if hyps[0].ReasoningConfidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", hyps[0].ReasoningConfidence)
	}
	if hyps[0].CalibratedConfidence != 0.0 {
		t.Errorf("expected CalibratedConfidence == 0 prior to Stage S7, got %f", hyps[0].CalibratedConfidence)
	}
}

func TestSymbolicSpecialist_EvaluateWithoutFrame(t *testing.T) {
	specialist := NewSymbolicSpecialist()
	env := communication.Envelope{ID: "env-raw", RawConfidence: 0.88}

	hyps, err := specialist.Evaluate(context.Background(), env, nil, nil)
	if err != nil {
		t.Fatalf("expected symbolic evaluation without frame to succeed, got %v", err)
	}
	if len(hyps) != 1 {
		t.Fatalf("expected 1 hypothesis, got %d", len(hyps))
	}
	if hyps[0].ReasoningConfidence != 0.88 {
		t.Errorf("expected confidence 0.88, got %f", hyps[0].ReasoningConfidence)
	}
	if hyps[0].CalibratedConfidence != 0.0 {
		t.Errorf("expected CalibratedConfidence == 0 prior to Stage S7, got %f", hyps[0].CalibratedConfidence)
	}
}

func TestSymbolicSpecialist_DialogueIntentsGenerateProposedGoals(t *testing.T) {
	specialist := NewSymbolicSpecialist()
	tests := []struct {
		intent         string
		wantTarget     string
		wantStateKey   string
		wantStateVal   string
		wantConclusion string
	}{
		{
			intent:         "greet_user",
			wantTarget:     "user",
			wantStateKey:   "acknowledged",
			wantStateVal:   "true",
			wantConclusion: `Derived symbolic conclusion for intent "greet_user"`,
		},
		{
			intent:         "query_identity",
			wantTarget:     "system_identity",
			wantStateKey:   "identity_communicated",
			wantStateVal:   "true",
			wantConclusion: `Derived symbolic conclusion for intent "query_identity"`,
		},
		{
			intent:         "query_wellbeing",
			wantTarget:     "system_status",
			wantStateKey:   "status_communicated",
			wantStateVal:   "true",
			wantConclusion: `Derived symbolic conclusion for intent "query_wellbeing"`,
		},
		{
			intent:         "farewell_user",
			wantTarget:     "user",
			wantStateKey:   "session_concluded",
			wantStateVal:   "true",
			wantConclusion: `Derived symbolic conclusion for intent "farewell_user"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.intent, func(t *testing.T) {
			frame := &understanding.SemanticFrame{
				PrimaryHypothesis: understanding.Hypothesis{
					Intent:               tc.intent,
					CalibratedConfidence: 0.90,
				},
			}
			hyps, err := specialist.Evaluate(context.Background(), communication.Envelope{ID: "env-" + tc.intent}, frame, nil)
			if err != nil {
				t.Fatalf("Evaluate failed: %v", err)
			}
			if len(hyps) != 1 {
				t.Fatalf("expected 1 hypothesis, got %d", len(hyps))
			}
			hyp := hyps[0]
			if hyp.Conclusion != tc.wantConclusion {
				t.Errorf("expected Conclusion %q, got %q", tc.wantConclusion, hyp.Conclusion)
			}
			if hyp.ProposedGoal == nil {
				t.Fatalf("expected ProposedGoal != nil for dialogue intent %s", tc.intent)
			}
			if err := hyp.ProposedGoal.Validate(); err != nil {
				t.Fatalf("ProposedGoal failed validation: %v", err)
			}
			if hyp.ProposedGoal.Kind != GoalKindCommunicative {
				t.Errorf("expected Kind COMMUNICATIVE, got %s", hyp.ProposedGoal.Kind)
			}
			if hyp.ProposedGoal.Intent != tc.intent {
				t.Errorf("expected Intent %q, got %q", tc.intent, hyp.ProposedGoal.Intent)
			}
			if hyp.ProposedGoal.Target != tc.wantTarget {
				t.Errorf("expected Target %q, got %q", tc.wantTarget, hyp.ProposedGoal.Target)
			}
			if val, ok := hyp.ProposedGoal.DesiredState[tc.wantStateKey]; !ok || val != tc.wantStateVal {
				t.Errorf("expected DesiredState[%s] == %s, got %v", tc.wantStateKey, tc.wantStateVal, hyp.ProposedGoal.DesiredState)
			}
		})
	}
}

func TestSymbolicSpecialist_MemoryKnowledgeRuleMatching(t *testing.T) {
	specialist := NewSymbolicSpecialist()
	ruleJSON := []byte(`{
		"rule_id": "rule-custom-action",
		"intent": "custom_action",
		"required_slots": [{"slot": "permission", "operator": "eq", "value": "granted"}],
		"consequent": {
			"kind": "INFORMATION_RETRIEVAL",
			"intent": "custom_action",
			"target": "system",
			"desired_state": {"executed": "true"}
		}
	}`)

	memRecords := []memory.Record{
		{ID: "rule-custom-action", Type: "knowledge_rule", Payload: ruleJSON},
	}

	// Case 1: Slot matches condition -> rule triggers
	frameMatch := &understanding.SemanticFrame{
		PrimaryHypothesis: understanding.Hypothesis{
			Intent:               "custom_action",
			CalibratedConfidence: 0.95,
			Slots: []understanding.Slot{
				{Name: "permission", Value: "granted"},
			},
		},
	}
	hyps, err := specialist.Evaluate(context.Background(), communication.Envelope{ID: "env-match"}, frameMatch, memRecords)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if hyps[0].ProposedGoal == nil {
		t.Fatalf("expected matching memory rule to attach ProposedGoal")
	}
	if hyps[0].ProposedGoal.Kind != GoalKindInformationRetrieval || hyps[0].ProposedGoal.Intent != "custom_action" {
		t.Errorf("unexpected ProposedGoal: %+v", hyps[0].ProposedGoal)
	}
	if hyps[0].Conclusion != `Derived symbolic conclusion for intent "custom_action"` {
		t.Errorf("expected diagnostic conclusion preserved, got %q", hyps[0].Conclusion)
	}

	// Case 2: Slot does NOT match condition -> rule does not trigger, ProposedGoal is nil
	frameNoMatch := &understanding.SemanticFrame{
		PrimaryHypothesis: understanding.Hypothesis{
			Intent:               "custom_action",
			CalibratedConfidence: 0.95,
			Slots: []understanding.Slot{
				{Name: "permission", Value: "denied"},
			},
		},
	}
	hyps2, err := specialist.Evaluate(context.Background(), communication.Envelope{ID: "env-nomatch"}, frameNoMatch, memRecords)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if hyps2[0].ProposedGoal != nil {
		t.Errorf("expected ProposedGoal == nil when rule condition unmet, got %+v", hyps2[0].ProposedGoal)
	}
	if hyps2[0].Conclusion != `Derived symbolic conclusion for intent "custom_action"` {
		t.Errorf("expected diagnostic conclusion preserved when rule condition unmet")
	}
}
