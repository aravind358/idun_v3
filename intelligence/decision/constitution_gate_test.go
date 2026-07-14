package decision

import (
	"context"
	"testing"
)

func TestDefaultConstitutionalGate_Filter(t *testing.T) {
	gate := NewDefaultConstitutionalGate()
	ctx := context.Background()

	cs := CandidateSet{
		EpisodeID: "ep-gate-test",
		Candidates: []Candidate{
			{
				ID:          "cand-clean",
				Description: "Safe path",
				Attributes:  map[string]float64{"utility": 10.0},
			},
			{
				ID:           "cand-violation-tag",
				Description:  "Unsafe path",
				FlaggedRisks: []string{"SAFETY_VIOLATION"},
				Attributes:   map[string]float64{"utility": 99.0},
			},
			{
				ID:          "cand-violation-attr",
				Description: "Unsafe attribute",
				Attributes:  map[string]float64{"utility": 50.0, "constitutional_safety": -1.0},
			},
		},
	}

	surviving, rejected, err := gate.Filter(ctx, cs)
	if err != nil {
		t.Fatalf("unexpected Filter() error: %v", err)
	}

	if len(surviving) != 1 {
		t.Fatalf("expected exactly 1 surviving candidate, got %d", len(surviving))
	}
	if surviving[0].ID != "cand-clean" {
		t.Errorf("expected surviving candidate to be 'cand-clean', got '%s'", surviving[0].ID)
	}

	if len(rejected) != 2 {
		t.Fatalf("expected 2 rejected candidates, got %d", len(rejected))
	}
	for _, rej := range rejected {
		if rej.RejectionStage != "TIER_1_CONSTITUTION" {
			t.Errorf("expected RejectionStage 'TIER_1_CONSTITUTION', got '%s'", rej.RejectionStage)
		}
	}
}
