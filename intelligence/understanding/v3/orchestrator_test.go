package v3

import (
	"context"
	"idun/boundary/perception"
	"idun/core/foundation"
	"testing"
	"time"
)

type mockSpecialist struct {
	hyps []Hypothesis
	err  error
}

func (m *mockSpecialist) Analyze(ctx context.Context, env *perception.PerceptionEnvelope) ([]Hypothesis, error) {
	return m.hyps, m.err
}

func TestOrchestrator_Cascade(t *testing.T) {
	envID, _ := foundation.NewUUID()
	artID, _ := foundation.NewUUID()
	env, _ := perception.NewBuilder().
		ArtifactID(artID).
		EnvelopeID(envID).
		RawInput("turn on the lights").
		Version("3.0").
		Timestamp(time.Now()).
		Build()

	// Setup mock specialists
	grammar := &mockSpecialist{
		hyps: nil, // Grammar fails
	}
	neural := &mockSpecialist{
		hyps: []Hypothesis{
			NewHypothesis("turn_on_lights", 0.90, 0.0, LayerNeuralClassifier, nil),
			NewHypothesis("turn_off_lights", 0.85, 0.0, LayerNeuralClassifier, nil),
		},
	}
	deliberative := &mockSpecialist{
		hyps: nil,
	}
	orc := NewOrchestrator(grammar, neural, deliberative, &mockExtractor{}, nil, nil, nil)
	result, err := orc.Analyze(context.Background(), env)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Validate artifact lineage
	if result.EnvelopeID() != env.EnvelopeID() {
		t.Errorf("mismatch EnvelopeID")
	}
	if result.ParentArtifactID() != foundation.ParentArtifactID(env.ArtifactID()) {
		t.Errorf("expected parent ID %s, got %s", env.ArtifactID(), result.ParentArtifactID())
	}
	if result.ArtifactID() == "" {
		t.Errorf("expected new artifact ID")
	}

	// Validate evaluator logic
	interp := result.Interpretations()[0]
	if interp.PrimaryIntent() != "turn_on_lights" {
		t.Errorf("expected turn_on_lights, got %s", interp.PrimaryIntent())
	}
	if interp.Status() != StatusAmbiguous {
		t.Errorf("expected AMBIGUOUS_BEAM status due to 0.05 delta, got %s", interp.Status())
	}
	if len(interp.AmbiguitySet()) != 1 {
		t.Errorf("expected 1 ambiguity, got %d", len(interp.AmbiguitySet()))
	}
	if interp.AmbiguitySet()[0].DeltaFromPrimary() < 0.049 || interp.AmbiguitySet()[0].DeltaFromPrimary() > 0.051 {
		t.Errorf("expected delta 0.05, got %v", interp.AmbiguitySet()[0].DeltaFromPrimary())
	}
}

func TestOrchestrator_Impasse(t *testing.T) {
	envID, _ := foundation.NewUUID()
	artID, _ := foundation.NewUUID()
	env, _ := perception.NewBuilder().
		ArtifactID(artID).
		EnvelopeID(envID).
		RawInput("asdfasdf").
		Version("3.0").
		Timestamp(time.Now()).
		Build()

	// Low confidence across the board
	neural := &mockSpecialist{
		hyps: []Hypothesis{
			NewHypothesis("unknown_intent", 0.35, 0.0, LayerNeuralClassifier, nil),
		},
	}
	orc := NewOrchestrator(nil, neural, nil, &mockExtractor{}, nil, nil, nil)
	result, err := orc.Analyze(context.Background(), env)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	interp := result.Interpretations()[0]
	if interp.Status() != StatusFailed {
		t.Errorf("expected FAILED_IMPASSE, got %s", interp.Status())
	}
	if interp.PrimaryIntent() != "unresolved_intent" {
		t.Errorf("expected unresolved_intent, got %s", interp.PrimaryIntent())
	}
}
type mockExtractor struct{}
func (m *mockExtractor) Run(hyp Hypothesis, b *Builder) {}
