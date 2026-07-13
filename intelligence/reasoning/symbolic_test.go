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
