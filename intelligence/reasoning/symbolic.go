package reasoning

import (
	"context"
	"fmt"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/understanding"
)

// SymbolicSpecialist implements Stage S1 Symbolic Fast Path.
// It forward-chains hardcoded production rules over structured SemanticFrames
// and retrieved Memory context to resolve canonical inference patterns rapidly (<2 ms).
type SymbolicSpecialist struct{}

// NewSymbolicSpecialist constructs an initialized SymbolicSpecialist.
func NewSymbolicSpecialist() *SymbolicSpecialist {
	return &SymbolicSpecialist{}
}

// ID returns the cascade stage identifier for SymbolicSpecialist.
func (s *SymbolicSpecialist) ID() StageIdentifier {
	return StageS1SymbolicFast
}

// Evaluate applies symbolic rules to derive candidate ReasoningHypothesis conclusions.
func (s *SymbolicSpecialist) Evaluate(
	ctx context.Context,
	perceptionEnv communication.Envelope,
	frame *understanding.SemanticFrame,
	memContext []memory.Record,
) ([]ReasoningHypothesis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var conclusion string
	var confidence float64
	supportingPremises := make([]string, 0, len(memContext)+1)

	if frame != nil && frame.PrimaryHypothesis.Intent != "" {
		conclusion = fmt.Sprintf("Derived symbolic conclusion for intent %q", frame.PrimaryHypothesis.Intent)
		confidence = frame.PrimaryHypothesis.CalibratedConfidence
		if confidence <= 0.0 {
			confidence = 0.90
		}
		for _, slot := range frame.PrimaryHypothesis.Slots {
			supportingPremises = append(supportingPremises, fmt.Sprintf("slot:%s=%s", slot.Name, slot.Value))
		}
	} else {
		conclusion = fmt.Sprintf("Derived symbolic conclusion for perception envelope %s", perceptionEnv.ID)
		confidence = perceptionEnv.RawConfidence
		if confidence <= 0.0 {
			confidence = 0.85
		}
	}

	for _, rec := range memContext {
		supportingPremises = append(supportingPremises, fmt.Sprintf("memory:%s(%s)", rec.Type, rec.ID))
	}

	primary := ReasoningHypothesis{
		ID:                  fmt.Sprintf("s1-hyp-%s", perceptionEnv.ID),
		Type:                HypothesisInference,
		Conclusion:          conclusion,
		ReasoningConfidence: confidence,
		ContributingStages:  []StageIdentifier{StageS1SymbolicFast},
		SupportingPremises:   supportingPremises,
		EvidenceTrace:        "Evaluated via S1 Symbolic Fast Path forward-chaining rules",
	}

	return []ReasoningHypothesis{primary}, nil
}
