package reasoning

import (
	"context"
	"testing"
)

func TestBayesianFusionSpecialist_FuseEvidence(t *testing.T) {
	specialist := NewBayesianFusionSpecialist()
	if specialist.ID() != StageS4EvidenceFusion {
		t.Errorf("expected stage ID S4, got %s", specialist.ID())
	}

	inputs := []ReasoningHypothesis{
		{
			ID:                  "hyp-support",
			Type:                HypothesisInference,
			Conclusion:          "Subject authorized",
			ReasoningConfidence: 0.80,
			SupportingPremises:  []string{"slot:role=admin", "memory:fact(auth_policy)"},
			ContributingStages:  []StageIdentifier{StageS1SymbolicFast, StageS2RelationalGraph},
		},
		{
			ID:                  "hyp-conflict",
			Type:                HypothesisInference,
			Conclusion:          "Subject revoked",
			ReasoningConfidence: 0.70,
			SupportingPremises:  []string{"contradiction:revoked_token"},
			ContributingStages:  []StageIdentifier{StageS1SymbolicFast},
		},
	}

	fused, err := specialist.FuseEvidence(context.Background(), inputs)
	if err != nil {
		t.Fatalf("expected fusion to succeed, got %v", err)
	}
	if len(fused) != 2 {
		t.Fatalf("expected 2 fused hypotheses, got %d", len(fused))
	}

	// Supported hypothesis confidence should increase
	if fused[0].ReasoningConfidence <= 0.80 {
		t.Errorf("expected fused support confidence > 0.80, got %f", fused[0].ReasoningConfidence)
	}
	// Conflicted hypothesis confidence should decrease
	if fused[1].ReasoningConfidence >= 0.70 {
		t.Errorf("expected fused conflict confidence < 0.70, got %f", fused[1].ReasoningConfidence)
	}
	// CalibratedConfidence must remain 0.0 until Stage S7 Calibration writes it
	if fused[0].CalibratedConfidence != 0.0 || fused[1].CalibratedConfidence != 0.0 {
		t.Errorf("expected CalibratedConfidence == 0.0 prior to Stage S7, got %f / %f", fused[0].CalibratedConfidence, fused[1].CalibratedConfidence)
	}

	for _, h := range fused {
		if err := h.Validate(); err != nil {
			t.Fatalf("fused hypothesis failed validation: %v", err)
		}
	}
}
