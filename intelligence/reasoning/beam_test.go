package reasoning

import (
	"testing"
)

func TestBeamSelectionSpecialist_SelectBeam(t *testing.T) {
	specialist := NewBeamSelectionSpecialist()
	if specialist.ID() != StageS6BeamSelection {
		t.Errorf("expected stage ID S6, got %s", specialist.ID())
	}

	hyps := []ReasoningHypothesis{
		{ID: "h3", ReasoningConfidence: 0.70, Conclusion: "Lower rank"},
		{ID: "h1", ReasoningConfidence: 0.90, Conclusion: "Top winner"},
		{ID: "h2", ReasoningConfidence: 0.82, Conclusion: "Close runner-up within threshold"},
		{ID: "h4", ReasoningConfidence: 0.40, Conclusion: "Too low"},
	}

	primary, beam, err := specialist.SelectBeam(hyps, 3, 0.25)
	if err != nil {
		t.Fatalf("expected SelectBeam to succeed, got %v", err)
	}

	if primary.ID != "h1" {
		t.Errorf("expected primary h1, got %s", primary.ID)
	}

	if len(beam) != 2 {
		t.Fatalf("expected 2 runners-up in beam within ambiguity threshold, got %d", len(beam))
	}
	if beam[0].ID != "h2" || beam[1].ID != "h3" {
		t.Errorf("expected beam runner-ups h2 and h3, got %v", beam)
	}
}

func TestBeamSelectionSpecialist_EmptyList(t *testing.T) {
	specialist := NewBeamSelectionSpecialist()
	_, _, err := specialist.SelectBeam(nil, 3, 0.25)
	if err != ErrEmptyHypotheses {
		t.Errorf("expected ErrEmptyHypotheses, got %v", err)
	}
}
