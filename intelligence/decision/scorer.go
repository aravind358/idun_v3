package decision

import (
	"context"
	"fmt"
	"math"
	"sort"
)

// DefaultObjectiveScorer implements Tier2ObjectiveScorer, providing fast-path linear utility
// scoring for Reflexive mode and Multi-Criteria Decision Analysis (MCDA) trade-off matrix for Deliberative mode.
type DefaultObjectiveScorer struct{}

// NewDefaultObjectiveScorer constructs the Tier 2 objective utility scorer.
func NewDefaultObjectiveScorer() *DefaultObjectiveScorer {
	return &DefaultObjectiveScorer{}
}

// ScoreReflexive computes fast-path linear dot-product utility scores U(c_i) = w^T * x_i.
// It returns CandidateScore records sorted in descending order of score.
func (s *DefaultObjectiveScorer) ScoreReflexive(candidates []Candidate, snapshot *DecisionStrategySnapshot) ([]CandidateScore, error) {
	if snapshot == nil {
		return nil, ErrInvalidStrategySnapshot
	}
	if len(candidates) == 0 {
		return []CandidateScore{}, nil
	}

	scores := make([]CandidateScore, 0, len(candidates))
	weights := snapshot.FeatureWeights

	for _, cand := range candidates {
		var rawScore float64
		// Compute dot-product w^T * x_i
		for attrName, attrVal := range cand.Attributes {
			w, ok := weights[attrName]
			if !ok {
				w = 1.0 // default weight for unconfigured feature
			}
			rawScore += w * attrVal
		}

		// Incorporate explicit benefit vs cost differential
		rawScore += (cand.EstimatedBenefit - cand.EstimatedCost)

		// Compute bounded confidence in [0.0, 1.0] using sigmoid squash over raw score and attribute density
		conf := 1.0 / (1.0 + math.Exp(-0.5*rawScore))
		if conf < 0.05 {
			conf = 0.05
		} else if conf > 0.9999 {
			conf = 0.9999
		}

		scores = append(scores, CandidateScore{
			CandidateID: cand.ID,
			Score:       rawScore,
			Confidence:  conf,
			Rationale:   fmt.Sprintf("linear utility score %.3f (confidence %.3f)", rawScore, conf),
		})
	}

	// Sort descending by utility score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores, nil
}

// ScoreDeliberative computes rigorous Multi-Criteria Decision Analysis (MCDA) and generates
// a pairwise trade-off matrix comparing score differentials across all surviving candidate pairs.
func (s *DefaultObjectiveScorer) ScoreDeliberative(ctx context.Context, candidates []Candidate, snapshot *DecisionStrategySnapshot) ([]CandidateScore, map[string]map[string]float64, error) {
	scores, err := s.ScoreReflexive(candidates, snapshot)
	if err != nil {
		return nil, nil, err
	}

	tradeoffMatrix := make(map[string]map[string]float64, len(scores))
	scoreMap := make(map[string]float64, len(scores))
	for _, sc := range scores {
		scoreMap[sc.CandidateID] = sc.Score
	}

	for _, scA := range scores {
		matrixRow := make(map[string]float64, len(scores))
		for _, scB := range scores {
			if scA.CandidateID == scB.CandidateID {
				matrixRow[scB.CandidateID] = 0.0
			} else {
				matrixRow[scB.CandidateID] = scA.Score - scB.Score
			}
		}
		tradeoffMatrix[scA.CandidateID] = matrixRow
	}

	return scores, tradeoffMatrix, nil
}
