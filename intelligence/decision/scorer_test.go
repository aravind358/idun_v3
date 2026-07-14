package decision

import (
	"context"
	"math"
	"testing"
)

func TestDefaultObjectiveScorer_ReflexiveAndDeliberative(t *testing.T) {
	scorer := NewDefaultObjectiveScorer()
	snap := NewDefaultStrategySnapshot("v2.0.0-test")

	candidates := []Candidate{
		{
			ID:          "c1",
			Description: "High utility option",
			Attributes:  map[string]float64{"utility": 2.0, "safety": 1.0},
		},
		{
			ID:          "c2",
			Description: "Low utility option",
			Attributes:  map[string]float64{"utility": 0.5, "safety": 1.0},
		},
	}

	scores, err := scorer.ScoreReflexive(candidates, snap)
	if err != nil {
		t.Fatalf("ScoreReflexive error: %v", err)
	}
	if len(scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(scores))
	}
	if scores[0].CandidateID != "c1" {
		t.Errorf("expected c1 to rank first, got %s", scores[0].CandidateID)
	}
	if scores[0].Score <= scores[1].Score {
		t.Errorf("expected score c1 > score c2, got %.3f <= %.3f", scores[0].Score, scores[1].Score)
	}

	// Verify Deliberative MCDA trade-off matrix
	ctx := context.Background()
	_, matrix, err := scorer.ScoreDeliberative(ctx, candidates, snap)
	if err != nil {
		t.Fatalf("ScoreDeliberative error: %v", err)
	}

	expectedDiff := scores[0].Score - scores[1].Score
	actualDiff := matrix["c1"]["c2"]
	if math.Abs(expectedDiff-actualDiff) > 1e-6 {
		t.Errorf("expected tradeoff diff %.6f, got %.6f", expectedDiff, actualDiff)
	}
}
