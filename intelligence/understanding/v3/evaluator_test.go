package v3

import (
	"testing"
)

func TestEvaluator_Unambiguous(t *testing.T) {
	hyps := []Hypothesis{
		NewHypothesis("turn_on", 0.95, 0.0, LayerNeuralClassifier, nil),
		NewHypothesis("turn_off", 0.70, 0.0, LayerNeuralClassifier, nil),
	}

	primary, ambSet, status := EvaluateHypotheses(hyps)

	if status != StatusUnambiguous {
		t.Errorf("expected UNAMBIGUOUS, got %s", status)
	}
	if primary.Intent() != "turn_on" {
		t.Errorf("expected turn_on, got %s", primary.Intent())
	}
	if len(ambSet) != 0 {
		t.Errorf("expected 0 ambiguities, got %d", len(ambSet))
	}
}

func TestEvaluator_BeamWidth(t *testing.T) {
	hyps := []Hypothesis{
		NewHypothesis("play_music", 0.90, 0.0, LayerNeuralClassifier, nil),
		NewHypothesis("play_podcast", 0.88, 0.0, LayerNeuralClassifier, nil),
		NewHypothesis("play_radio", 0.86, 0.0, LayerNeuralClassifier, nil),
		NewHypothesis("play_news", 0.85, 0.0, LayerNeuralClassifier, nil), // Should be truncated
	}

	primary, ambSet, status := EvaluateHypotheses(hyps)

	if status != StatusAmbiguous {
		t.Errorf("expected AMBIGUOUS_BEAM, got %s", status)
	}
	if len(ambSet) != 2 { // MaxBeamWidth is 3, so primary + 2 amb
		t.Fatalf("expected 2 ambiguities, got %d", len(ambSet))
	}

	// Verify deltas were calculated correctly
	if ambSet[0].Intent() != "play_podcast" || ambSet[0].DeltaFromPrimary() > 0.021 {
		t.Errorf("bad delta for 1st amb: %v", ambSet[0].DeltaFromPrimary())
	}
	if ambSet[1].Intent() != "play_radio" || ambSet[1].DeltaFromPrimary() > 0.041 {
		t.Errorf("bad delta for 2nd amb: %v", ambSet[1].DeltaFromPrimary())
	}
	
	// Primary delta should be 0
	if primary.DeltaFromPrimary() != 0.0 {
		t.Errorf("primary delta should be 0.0, got %v", primary.DeltaFromPrimary())
	}
}

func TestEvaluator_ClampAndImpasse(t *testing.T) {
	hyps := []Hypothesis{
		NewHypothesis("weather", 1.5, 0.0, LayerNeuralClassifier, nil), // Over 1.0 -> 1.0
		NewHypothesis("news", 0.9, 0.0, LayerNeuralClassifier, nil),    // Delta = 0.1
	}

	primary, ambSet, status := EvaluateHypotheses(hyps)
	if status != StatusAmbiguous {
		t.Errorf("expected AMBIGUOUS_BEAM, got %s", status)
	}
	if primary.Confidence() != 1.0 {
		t.Errorf("expected clamped 1.0, got %v", primary.Confidence())
	}
	if len(ambSet) != 1 || ambSet[0].DeltaFromPrimary() > 0.101 {
		t.Errorf("expected delta 0.1, got %v", ambSet[0].DeltaFromPrimary())
	}
}
