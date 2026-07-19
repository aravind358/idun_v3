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

func TestBayesianFusionSpecialist_CorroboratingEquivalentGoals(t *testing.T) {
	// Proves Phase 3, 4, and 9: Corroboration of equivalent semantic goals across S1 and S5
	specialist := NewBayesianFusionSpecialist()

	goalS1 := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
		},
	}
	goalS5 := goalS1.Clone() // identical goal

	inputs := []ReasoningHypothesis{
		{
			ID:                  "hyp-s1",
			Type:                HypothesisType("Symbolic"),
			Conclusion:          `Derived symbolic conclusion for intent "greet_user"`,
			ReasoningConfidence: 0.75,
			ProposedGoal:        goalS1,
			SupportingPremises:  []string{"rule_match=dialogue_intent:greet_user"},
			ContributingStages:  []StageIdentifier{StageS1SymbolicFast},
		},
		{
			ID:                  "hyp-s5",
			Type:                HypothesisType("Analogy"),
			Conclusion:          "Analogical match from historical case",
			ReasoningConfidence: 0.78,
			ProposedGoal:        goalS5,
			SupportingPremises:  []string{"analogical_case=case-001", "stored_goal_recovered=true"},
			ContributingStages:  []StageIdentifier{StageIdentifier("STAGE_S5_ANALOGY")},
		},
	}

	if len(inputs) != 2 {
		t.Fatalf("expected 2 initial hypotheses before fusion, got %d", len(inputs))
	}

	fpS1 := goalS1.Fingerprint()
	fpS5 := goalS5.Fingerprint()
	if fpS1 != fpS5 {
		t.Fatalf("expected identical fingerprints, got %q vs %q", fpS1, fpS5)
	}

	fused, err := specialist.FuseEvidence(context.Background(), inputs)
	if err != nil {
		t.Fatalf("FuseEvidence failed: %v", err)
	}

	if len(fused) != 1 {
		t.Fatalf("expected exactly 1 fused hypothesis representing corroborating goals, got %d", len(fused))
	}

	fusedHyp := fused[0]
	if fusedHyp.ProposedGoal == nil || fusedHyp.ProposedGoal.Fingerprint() != fpS1 {
		t.Errorf("expected fused hypothesis to retain canonical goal fingerprint %q", fpS1)
	}
	if fusedHyp.ReasoningConfidence <= 0.78 {
		t.Errorf("expected Bayesian corroboration to increase confidence above max initial (0.78), got %f", fusedHyp.ReasoningConfidence)
	}

	// Verify merged unique premises
	premiseSet := make(map[string]bool)
	for _, p := range fusedHyp.SupportingPremises {
		premiseSet[p] = true
	}
	if !premiseSet["rule_match=dialogue_intent:greet_user"] || !premiseSet["analogical_case=case-001"] {
		t.Errorf("expected merged premises from both S1 and S5, got %v", fusedHyp.SupportingPremises)
	}

	// Verify contributing stages
	stageSet := make(map[StageIdentifier]bool)
	for _, st := range fusedHyp.ContributingStages {
		stageSet[st] = true
	}
	if !stageSet[StageS1SymbolicFast] || !stageSet[StageIdentifier("STAGE_S5_ANALOGY")] || !stageSet[StageS4EvidenceFusion] {
		t.Errorf("expected all contributing stages including S4 fusion, got %v", fusedHyp.ContributingStages)
	}
}

func TestBayesianFusionSpecialist_ConflictingGoals(t *testing.T) {
	// Proves Phase 5 and 10: Conflicting goals remain separate hypotheses
	specialist := NewBayesianFusionSpecialist()

	goalGreet := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
		},
	}
	goalFarewell := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "farewell_user",
		Target: "user",
		DesiredState: map[string]string{
			"conversation_closed": "true",
		},
	}

	inputs := []ReasoningHypothesis{
		{
			ID:                  "hyp-greet",
			Type:                HypothesisType("Symbolic"),
			Conclusion:          `Derived symbolic conclusion for intent "greet_user"`,
			ReasoningConfidence: 0.80,
			ProposedGoal:        goalGreet,
			SupportingPremises:  []string{"rule_match=dialogue_intent:greet_user"},
			ContributingStages:  []StageIdentifier{StageS1SymbolicFast},
		},
		{
			ID:                  "hyp-farewell",
			Type:                HypothesisType("Deliberative"),
			Conclusion:          `Deliberative conclusion for farewell`,
			ReasoningConfidence: 0.65,
			ProposedGoal:        goalFarewell,
			SupportingPremises:  []string{"provenance=LLM_FALLBACK"},
			ContributingStages:  []StageIdentifier{StageS8DeliberativeLLM},
		},
	}

	fused, err := specialist.FuseEvidence(context.Background(), inputs)
	if err != nil {
		t.Fatalf("FuseEvidence failed: %v", err)
	}

	if len(fused) != 2 {
		t.Fatalf("expected 2 distinct hypotheses after fusion for conflicting goals, got %d", len(fused))
	}
}

func TestBayesianFusionSpecialist_NilAndInvalidGoals(t *testing.T) {
	specialist := NewBayesianFusionSpecialist()

	validGoal := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "greet_user",
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
		},
	}
	invalidGoal := &SemanticGoal{
		Kind:   GoalKindCommunicative,
		Intent: "", // invalid
		Target: "user",
		DesiredState: map[string]string{
			"acknowledged": "true",
		},
	}

	inputs := []ReasoningHypothesis{
		{
			ID:                  "hyp-nil-1",
			Type:                HypothesisInference,
			Conclusion:          "Default inferred conclusion 1",
			ReasoningConfidence: 0.80,
			ProposedGoal:        nil,
		},
		{
			ID:                  "hyp-nil-2",
			Type:                HypothesisInference,
			Conclusion:          "Default inferred conclusion 2",
			ReasoningConfidence: 0.70,
			ProposedGoal:        nil,
		},
		{
			ID:                  "hyp-invalid",
			Type:                HypothesisInference,
			Conclusion:          "Invalid goal conclusion",
			ReasoningConfidence: 0.75,
			ProposedGoal:        invalidGoal,
		},
		{
			ID:                  "hyp-valid",
			Type:                HypothesisInference,
			Conclusion:          "Valid goal conclusion",
			ReasoningConfidence: 0.82,
			ProposedGoal:        validGoal,
		},
	}

	fused, err := specialist.FuseEvidence(context.Background(), inputs)
	if err != nil {
		t.Fatalf("FuseEvidence failed: %v", err)
	}
	if len(fused) != 4 {
		t.Fatalf("expected 4 distinct hypotheses (nil and invalid goals are not grouped), got %d", len(fused))
	}
}
