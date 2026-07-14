package decision

import (
	"testing"
)

func TestParetoDominates_AndFindParetoFrontier(t *testing.T) {
	cands := []Candidate{
		{
			ID: "cand-A", // Dominates C
			Attributes: map[string]float64{
				"utility": 0.90,
				"safety":  0.95,
			},
		},
		{
			ID: "cand-B", // Pareto-efficient trade-off against A
			Attributes: map[string]float64{
				"utility": 0.95,
				"safety":  0.85,
			},
		},
		{
			ID: "cand-C", // Dominated by A
			Attributes: map[string]float64{
				"utility": 0.80,
				"safety":  0.90,
			},
		},
	}

	efficient, dominated, err := FindParetoFrontier(cands, []string{"utility", "safety"})
	if err != nil {
		t.Fatalf("FindParetoFrontier error: %v", err)
	}

	if len(efficient) != 2 {
		t.Errorf("expected 2 efficient candidates (A, B), got %d", len(efficient))
	}
	if len(dominated) != 1 {
		t.Errorf("expected 1 dominated candidate (C), got %d", len(dominated))
	}
	if len(dominated) > 0 && dominated[0].ID != "cand-C" {
		t.Errorf("expected cand-C to be dominated, got %s", dominated[0].ID)
	}
}

func TestComputeTradeoffDistance(t *testing.T) {
	attrsA := map[string]float64{"x": 1.0, "y": 2.0}
	attrsB := map[string]float64{"x": 4.0, "y": 6.0}
	weights := map[string]float64{"x": 1.0, "y": 1.0}

	dist := ComputeTradeoffDistance(attrsA, attrsB, weights)
	// sqrt((1-4)^2 + (2-6)^2) = sqrt(9 + 16) = 5.0
	if dist != 5.0 {
		t.Errorf("expected tradeoff distance 5.0, got %f", dist)
	}
}
