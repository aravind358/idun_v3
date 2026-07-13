package reasoning

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestReasoningResultValidation_Success(t *testing.T) {
	result := ReasoningResult{
		SchemaVersion: SchemaVersion,
		EnvelopeID:    "env-001",
		SourceFrameID: "frame-001",
		Status:        StatusUnambiguousSolved,
		StrategyUsed:  StrategySymbolicFast,
		PrimaryHypothesis: ReasoningHypothesis{
			ID:                   "hyp-001",
			Type:                 HypothesisInference,
			Conclusion:           "Subject is authorized",
			CalibratedConfidence: 0.95,
		},
		AmbiguitySet: []ReasoningHypothesis{
			{
				ID:                   "hyp-002",
				Type:                 HypothesisInference,
				Conclusion:           "Subject requires elevation",
				CalibratedConfidence: 0.82,
			},
		},
		ContradictionsFlagged: []ContradictionFlag{
			{
				BeliefID:         "bel-101",
				ConflictingClaim: "Subject was revoked",
				Confidence:       0.90,
				DetectedAtStage:  StageS3CSPCheck,
			},
		},
		ProposedBeliefUpdates: []BeliefUpdateProposal{
			{
				Subject:            "user:42",
				Predicate:          "status",
				Object:             "authorized",
				ProposedConfidence: 0.95,
				SourceHypothesisID: "hyp-001",
			},
		},
		CompilationCandidate: &CompilationCandidate{
			EligibleForLearning:  true,
			SourceStage:          StageS8DeliberativeLLM,
			StrategyUsed:         StrategyGraphDeliberative,
			CalibratedConfidence: 0.94,
			Antecedents: []CompilationCondition{
				{Slot: "role", Operator: "EQ", Value: "admin"},
			},
			DerivedConsequent: CompilationConsequent{
				Intent:     "grant_access",
				RuleAction: "FAST_ALLOW",
			},
		},
		StrategyTelemetry: StrategyTelemetry{
			EpisodeID:            "ep-001",
			StrategySelected:     StrategySymbolicFast,
			SpecialistsExecuted:  []StageIdentifier{StageS0ContextAssembly, StageS1SymbolicFast},
			ExecutionDurationMs:  0.8,
			CalibratedConfidence: 0.95,
			ResourceCostTier:     "LOCAL_FAST",
			EscalatedToLLM:       false,
			OutcomeStatus:        StatusUnambiguousSolved,
		},
	}

	if err := result.Validate(); err != nil {
		t.Fatalf("expected valid result, got %v", err)
	}
}

func TestReasoningResultValidation_Failures(t *testing.T) {
	validHyp := ReasoningHypothesis{
		ID:                   "hyp-1",
		Type:                 HypothesisInference,
		Conclusion:           "conclusion",
		CalibratedConfidence: 0.8,
	}

	tests := []struct {
		name    string
		mutate  func(*ReasoningResult)
		wantErr error
	}{
		{
			name: "invalid schema version",
			mutate: func(r *ReasoningResult) {
				r.SchemaVersion = "1.0"
			},
			wantErr: ErrInvalidSchemaVersion,
		},
		{
			name: "missing envelope ID",
			mutate: func(r *ReasoningResult) {
				r.EnvelopeID = ""
			},
			wantErr: ErrMissingEnvelopeID,
		},
		{
			name: "missing source frame ID",
			mutate: func(r *ReasoningResult) {
				r.SourceFrameID = ""
			},
			wantErr: ErrMissingSourceFrameID,
		},
		{
			name: "beam overflow (> MaxBeamWidth)",
			mutate: func(r *ReasoningResult) {
				r.AmbiguitySet = []ReasoningHypothesis{validHyp, validHyp, validHyp} // 1 primary + 3 runners-up = 4 > 3
			},
			wantErr: ErrBeamOverflow,
		},
		{
			name: "invalid confidence out of bounds",
			mutate: func(r *ReasoningResult) {
				r.PrimaryHypothesis.CalibratedConfidence = 1.5
			},
			wantErr: ErrInvalidConfidence,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := ReasoningResult{
				SchemaVersion:     SchemaVersion,
				EnvelopeID:        "env-1",
				SourceFrameID:     "frame-1",
				Status:            StatusUnambiguousSolved,
				PrimaryHypothesis: validHyp,
				StrategyTelemetry: StrategyTelemetry{
					EpisodeID:            "ep-1",
					CalibratedConfidence: 0.8,
				},
			}
			tc.mutate(&r)
			if err := r.Validate(); !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestReasoningResult_CloneAndImmutability(t *testing.T) {
	orig := ReasoningResult{
		SchemaVersion: SchemaVersion,
		EnvelopeID:    "env-orig",
		SourceFrameID: "frame-orig",
		Status:        StatusUnambiguousSolved,
		PrimaryHypothesis: ReasoningHypothesis{
			ID:                 "hyp-orig",
			Type:               HypothesisInference,
			Conclusion:         "Original",
			ContributingStages: []StageIdentifier{StageS1SymbolicFast},
		},
		AmbiguitySet: []ReasoningHypothesis{
			{
				ID:         "hyp-amb",
				Conclusion: "Ambiguous",
			},
		},
		CompilationCandidate: &CompilationCandidate{
			EligibleForLearning: true,
			Antecedents: []CompilationCondition{
				{Slot: "A", Operator: "EQ", Value: "1"},
			},
		},
	}

	clone := orig.Clone()
	clone.EnvelopeID = "env-mutated"
	clone.PrimaryHypothesis.Conclusion = "Mutated"
	clone.PrimaryHypothesis.ContributingStages[0] = StageS8DeliberativeLLM
	clone.CompilationCandidate.Antecedents[0].Value = "999"

	if orig.EnvelopeID == clone.EnvelopeID {
		t.Errorf("EnvelopeID was mutated on original")
	}
	if orig.PrimaryHypothesis.Conclusion == clone.PrimaryHypothesis.Conclusion {
		t.Errorf("PrimaryHypothesis.Conclusion was mutated on original")
	}
	if orig.PrimaryHypothesis.ContributingStages[0] == clone.PrimaryHypothesis.ContributingStages[0] {
		t.Errorf("ContributingStages slice was mutated on original")
	}
	if orig.CompilationCandidate.Antecedents[0].Value == "999" {
		t.Errorf("CompilationCandidate.Antecedents slice was mutated on original")
	}
}

func TestReasoningResult_JSONSerialization(t *testing.T) {
	orig := ReasoningResult{
		SchemaVersion: SchemaVersion,
		EnvelopeID:    "env-json",
		SourceFrameID: "frame-json",
		Status:        StatusUnambiguousSolved,
		StrategyUsed:  StrategySymbolicFast,
		PrimaryHypothesis: ReasoningHypothesis{
			ID:                   "hyp-1",
			Type:                 HypothesisInference,
			Conclusion:           "test conclusion",
			CalibratedConfidence: 0.9,
		},
		StrategyTelemetry: StrategyTelemetry{
			EpisodeID:            "ep-json",
			CalibratedConfidence: 0.9,
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("failed to marshal ReasoningResult: %v", err)
	}

	var decoded ReasoningResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ReasoningResult: %v", err)
	}

	if err := decoded.Validate(); err != nil {
		t.Fatalf("decoded object failed validation: %v", err)
	}
	if decoded.EnvelopeID != orig.EnvelopeID || decoded.PrimaryHypothesis.Conclusion != orig.PrimaryHypothesis.Conclusion {
		t.Errorf("decoded content mismatch: got %+v", decoded)
	}
}
