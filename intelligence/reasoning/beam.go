package reasoning

import (
	"errors"
	"sort"
)

var (
	ErrEmptyHypotheses = errors.New("reasoning: cannot select beam from empty hypothesis list")
)

// BeamSelectionSpecialist implements Stage S6 Multi-Hypothesis Beam Selection.
//
// MANDATORY INVARIANTS:
// 1. Sorts candidate hypotheses descending by ReasoningConfidence.
// 2. Preserves up to MaxBeamWidth (default 3) total candidates (1 primary + up to 2 ambiguity runners-up).
// 3. Preserves close runner-up hypotheses within ambiguity threshold rather than collapsing into a single winner.
type BeamSelectionSpecialist struct{}

// NewBeamSelectionSpecialist returns an initialized BeamSelectionSpecialist.
func NewBeamSelectionSpecialist() *BeamSelectionSpecialist {
	return &BeamSelectionSpecialist{}
}

// ID returns StageS6BeamSelection.
func (s *BeamSelectionSpecialist) ID() StageIdentifier {
	return StageS6BeamSelection
}

// SelectBeam sorts candidate hypotheses by ReasoningConfidence and returns the primary
// winner alongside an intentional ambiguity set of close runner-ups up to maxBeamWidth.
func (s *BeamSelectionSpecialist) SelectBeam(
	hypotheses []ReasoningHypothesis,
	maxBeamWidth int,
	ambiguityThreshold float64,
) (ReasoningHypothesis, []ReasoningHypothesis, error) {
	if len(hypotheses) == 0 {
		return ReasoningHypothesis{}, nil, ErrEmptyHypotheses
	}
	if maxBeamWidth <= 0 || maxBeamWidth > MaxBeamWidth {
		maxBeamWidth = MaxBeamWidth
	}
	if ambiguityThreshold <= 0.0 {
		ambiguityThreshold = 0.25 // Default ambiguity margin
	}

	sorted := make([]ReasoningHypothesis, len(hypotheses))
	copy(sorted, hypotheses)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ReasoningConfidence > sorted[j].ReasoningConfidence
	})

	primary := sorted[0]
	beam := make([]ReasoningHypothesis, 0, maxBeamWidth-1)

	for i := 1; i < len(sorted) && len(beam) < maxBeamWidth-1; i++ {
		diff := primary.ReasoningConfidence - sorted[i].ReasoningConfidence
		if diff <= ambiguityThreshold {
			beam = append(beam, sorted[i])
		}
	}

	return primary, beam, nil
}
