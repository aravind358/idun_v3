package v3

import (
	"encoding/json"
	"idun/intelligence/understanding/v3/ontology"
	"testing"
)

func TestSemanticInterpretation_Builder(t *testing.T) {
	obj, err := NewBuilder().
		EnvelopeID("env-123").
		Status(StatusUnambiguous).
		PrimaryIntent("book_flight").
		CompoundIntentCount(1).
		Confidence(0.95).
		Build()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if obj.SpecVersion() != "3.0" {
		t.Errorf("expected 3.0, got %s", obj.SpecVersion())
	}
	if obj.PrimaryIntent() != "book_flight" {
		t.Errorf("expected book_flight, got %s", obj.PrimaryIntent())
	}
}

func TestSemanticInterpretation_ValidationFailures(t *testing.T) {
	tests := []struct {
		name    string
		builder *Builder
	}{
		{"missing envelope ID", NewBuilder().Status(StatusUnambiguous).PrimaryIntent("x").CompoundIntentCount(1)},
		{"missing primary intent", NewBuilder().EnvelopeID("x").Status(StatusUnambiguous).CompoundIntentCount(1)},
		{"invalid confidence", NewBuilder().EnvelopeID("x").Status(StatusUnambiguous).PrimaryIntent("x").CompoundIntentCount(1).Confidence(1.5)},
		{"invalid status", NewBuilder().EnvelopeID("x").Status("UNKNOWN").PrimaryIntent("x").CompoundIntentCount(1)},
		{"impasse requires unresolved", NewBuilder().EnvelopeID("x").Status(StatusFailed).PrimaryIntent("x").CompoundIntentCount(1)},
		{"compound mismatch", NewBuilder().EnvelopeID("x").Status(StatusUnambiguous).PrimaryIntent("x").CompoundIntentCount(2)}, // No secondary intents added
		{"conditional mismatch", NewBuilder().EnvelopeID("x").Status(StatusUnambiguous).PrimaryIntent("x").CompoundIntentCount(1).IsConditional(true)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder.Build()
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestSemanticInterpretation_Serialization(t *testing.T) {
	obj, _ := NewBuilder().
		EnvelopeID("env-123").
		Status(StatusAmbiguous).
		PrimaryIntent("book_flight").
		CompoundIntentCount(1).
		Confidence(0.85).
		PrimaryHypothesis(NewHypothesis("book_flight", 0.85, 0.0, LayerNeuralClassifier, []Slot{
			NewSlot("destination", "Paris", "loc-123", 0.9),
		})).
		AmbiguitySet([]Hypothesis{
			NewHypothesis("check_flight", 0.80, 0.05, LayerNeuralClassifier, []Slot{}),
		}).
		Entities([]Entity{
			NewEntity("Paris", ontology.EntityLocation, "Paris, France", "loc-123", 0.95),
		}).
		Polarity(NewPolarity(true, "not")).
		Build()

	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var obj2 SemanticInterpretation
	if err := json.Unmarshal(data, &obj2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if obj2.EnvelopeID() != obj.EnvelopeID() {
		t.Errorf("mismatch EnvelopeID: got %s, want %s", obj2.EnvelopeID(), obj.EnvelopeID())
	}
	if obj2.PrimaryHypothesis().SourceLayer() != obj.PrimaryHypothesis().SourceLayer() {
		t.Errorf("mismatch SourceLayer: got %s, want %s", obj2.PrimaryHypothesis().SourceLayer(), obj.PrimaryHypothesis().SourceLayer())
	}
	if len(obj2.Entities()) != 1 || obj2.Entities()[0].CanonicalName() != "Paris, France" {
		t.Errorf("mismatch Entities")
	}
	if !obj2.Polarity().Negated() {
		t.Errorf("mismatch Polarity")
	}
}
