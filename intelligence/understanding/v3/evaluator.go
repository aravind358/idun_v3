package v3

import (
	"sort"
)

const (
	MaxBeamWidth      = 3
	MaxDelta          = 0.15
	ImpasseThreshold  = 0.40
)

// EvaluateHypotheses processes a raw list of hypotheses, sorts them by confidence,
// determines the primary hypothesis, constructs the ambiguity set based on the MaxDelta,
// and resolves the final InterpretationStatus.
func EvaluateHypotheses(hyps []Hypothesis) (Primary Hypothesis, AmbiguitySet []Hypothesis, Status InterpretationStatus) {
	if len(hyps) == 0 {
		return createImpasse(), nil, StatusFailed
	}

	// 1. Sort by confidence descending
	sort.Slice(hyps, func(i, j int) bool {
		return hyps[i].Confidence() > hyps[j].Confidence()
	})

	primary := hyps[0]
	clampedConf := clamp(primary.Confidence())

	// 2. Impasse Check
	if clampedConf < ImpasseThreshold {
		return createImpasse(), nil, StatusFailed
	}

	// 3. Build Ambiguity Set
	var ambSet []Hypothesis
	for i := 1; i < len(hyps); i++ {
		if len(ambSet) >= MaxBeamWidth-1 {
			break
		}

		c := clamp(hyps[i].Confidence())
		delta := clampedConf - c

		if delta <= MaxDelta {
			// Create a copy with the calculated delta
			h := NewHypothesis(
				hyps[i].Intent(),
				c,
				delta,
				hyps[i].SourceLayer(),
				hyps[i].Slots(),
			)
			ambSet = append(ambSet, h)
		}
	}

	// Update primary with clamped confidence and 0.0 delta
	primary = NewHypothesis(
		primary.Intent(),
		clampedConf,
		0.0,
		primary.SourceLayer(),
		primary.Slots(),
	)

	// 4. Resolve Status
	if len(ambSet) > 0 {
		return primary, ambSet, StatusAmbiguous
	}

	return primary, nil, StatusUnambiguous
}

func clamp(v float64) float64 {
	if v < 0.0 {
		return 0.0
	}
	if v > 1.0 {
		return 1.0
	}
	return v
}

func createImpasse() Hypothesis {
	return NewHypothesis("unresolved_intent", 0.0, 0.0, LayerReflexiveGrammar, nil)
}
