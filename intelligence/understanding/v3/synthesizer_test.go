package v3

import (
	"idun/boundary/perception"
	"idun/core/foundation"
	"testing"
	"time"
)

func TestSynthesize(t *testing.T) {
	envID, _ := foundation.NewUUID()
	artID, _ := foundation.NewUUID()
	env, _ := perception.NewBuilder().
		ArtifactID(artID).
		EnvelopeID(envID).
		RawInput("test").
		Version("3.0").
		Timestamp(time.Now()).
		Build()

	primary := NewHypothesis("primary_intent", 0.9, 0.0, LayerNeuralClassifier, nil)
	amb := NewHypothesis("secondary_intent", 0.8, 0.1, LayerNeuralClassifier, nil)

	result, err := Synthesize(env, primary, []Hypothesis{amb}, StatusAmbiguous, nil, nil, nil, nil, nil, 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.EnvelopeID() != env.EnvelopeID() {
		t.Errorf("expected EnvelopeID %s, got %s", env.EnvelopeID(), result.EnvelopeID())
	}
	if result.ParentArtifactID() != foundation.ParentArtifactID(env.ArtifactID()) {
		t.Errorf("expected ParentArtifactID %s, got %s", env.ArtifactID(), result.ParentArtifactID())
	}
	if result.PrimaryIntent() != "primary_intent" {
		t.Errorf("expected primary_intent, got %s", result.PrimaryIntent())
	}
	if len(result.AmbiguitySet()) != 1 {
		t.Errorf("expected 1 ambiguity, got %d", len(result.AmbiguitySet()))
	}
	
	// Validatable should pass
	if err := result.Validate(); err != nil {
		t.Errorf("synthesized result failed validation: %v", err)
	}
}

func TestSynthesize_Impasse(t *testing.T) {
	envID, _ := foundation.NewUUID()
	artID, _ := foundation.NewUUID()
	env, _ := perception.NewBuilder().
		ArtifactID(artID).
		EnvelopeID(envID).
		RawInput("test").
		Version("3.0").
		Timestamp(time.Now()).
		Build()

	primary := NewHypothesis("unresolved_intent", 0.0, 0.0, LayerReflexiveGrammar, nil)

	result, err := Synthesize(env, primary, nil, StatusFailed, nil, nil, nil, nil, nil, 0, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.PrimaryIntent() != "unresolved_intent" {
		t.Errorf("expected unresolved_intent, got %s", result.PrimaryIntent())
	}
	if result.Status() != StatusFailed {
		t.Errorf("expected FAILED_IMPASSE, got %s", result.Status())
	}
}
