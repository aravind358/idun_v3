package reasoning

import (
	"errors"
	"testing"
)

func TestReasoningResultBuilder_Success(t *testing.T) {
	primary := ReasoningHypothesis{
		ID:                   "hyp-1",
		Type:                 HypothesisInference,
		Conclusion:           "Subject is admin",
		CalibratedConfidence: 0.96,
	}

	ambiguity := ReasoningHypothesis{
		ID:                   "hyp-2",
		Type:                 HypothesisRelation,
		Conclusion:           "Subject reports to admin",
		CalibratedConfidence: 0.70,
	}

	result, err := NewReasoningResultBuilder("env-builder", "frame-builder").
		WithStatus(StatusUnambiguousSolved).
		WithStrategyUsed(StrategySymbolicFast).
		WithPrimaryHypothesis(primary).
		AddAmbiguityHypothesis(ambiguity).
		AddContradictionFlag(ContradictionFlag{
			BeliefID:         "bel-99",
			ConflictingClaim: "Subject is guest",
			Confidence:       0.85,
			DetectedAtStage:  StageS3CSPCheck,
		}).
		AddBeliefUpdateProposal(BeliefUpdateProposal{
			Subject:            "user",
			Predicate:          "role",
			Object:             "admin",
			ProposedConfidence: 0.96,
			SourceHypothesisID: "hyp-1",
		}).
		WithCompilationCandidate(&CompilationCandidate{
			EligibleForLearning:  true,
			SourceStage:          StageS1SymbolicFast,
			StrategyUsed:         StrategySymbolicFast,
			CalibratedConfidence: 0.96,
			Antecedents: []CompilationCondition{
				{Slot: "role", Operator: "EQ", Value: "admin"},
			},
			DerivedConsequent: CompilationConsequent{
				Intent:     "allow",
				RuleAction: "FAST_ALLOW",
			},
		}).
		WithStrategyTelemetry(StrategyTelemetry{
			EpisodeID:            "ep-builder",
			StrategySelected:     StrategySymbolicFast,
			CalibratedConfidence: 0.96,
		}).
		WithOfflineMode(true).
		WithProcessedDurationMs(1.1).
		Build()

	if err != nil {
		t.Fatalf("expected builder to succeed, got %v", err)
	}

	if result.EnvelopeID != "env-builder" || result.PrimaryHypothesis.ID != "hyp-1" {
		t.Fatalf("builder output mismatch: %+v", result)
	}
	if len(result.AmbiguitySet) != 1 || len(result.ContradictionsFlagged) != 1 || len(result.ProposedBeliefUpdates) != 1 {
		t.Fatalf("expected slices to have length 1, got %+v", result)
	}
}

func TestReasoningResultBuilder_MissingPrimaryHypothesis(t *testing.T) {
	_, err := NewReasoningResultBuilder("env-err", "frame-err").Build()
	if err == nil {
		t.Fatalf("expected error when building without primary hypothesis")
	}
}

func TestReasoningResultBuilder_MissingIDs(t *testing.T) {
	_, err := NewReasoningResultBuilder("", "frame-err").Build()
	if !errors.Is(err, ErrMissingEnvelopeID) {
		t.Fatalf("expected ErrMissingEnvelopeID, got %v", err)
	}

	_, err = NewReasoningResultBuilder("env-err", "").Build()
	if !errors.Is(err, ErrMissingSourceFrameID) {
		t.Fatalf("expected ErrMissingSourceFrameID, got %v", err)
	}
}

func TestStrategySpecBuilder_SuccessAndValidation(t *testing.T) {
	spec, err := NewStrategySpecBuilder(StrategyGraphDeliberative).
		EnableStage(StageS1SymbolicFast).
		EnableStage(StageS2RelationalGraph).
		WithPriorityOrder([]StageIdentifier{StageS1SymbolicFast, StageS2RelationalGraph}).
		WithGraphBounds(100, 500, 2).
		Build()

	if err != nil {
		t.Fatalf("expected strategy spec builder to succeed, got %v", err)
	}
	if spec.MaxGraphNodes != 100 || spec.MaxGraphEdges != 500 || spec.MaxGraphDepth != 2 {
		t.Fatalf("expected custom graph bounds, got %+v", spec)
	}
	if !spec.IsStageEnabled(StageS2RelationalGraph) {
		t.Errorf("expected stage S2 to be enabled")
	}
}
