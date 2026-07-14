package decision

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestDefaultDecisionService_EvaluateReflexive_Commit(t *testing.T) {
	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	cs := CandidateSet{
		EpisodeID: "ep-commit",
		Candidates: []Candidate{
			{
				ID:          "cand-winner",
				Description: "High utility clear winner",
				Attributes:  map[string]float64{"utility": 5.0, "safety": 2.0},
			},
			{
				ID:          "cand-loser",
				Description: "Lower score alternative",
				Attributes:  map[string]float64{"utility": 0.1, "safety": 1.0},
			},
		},
	}

	rec, err := srv.EvaluateReflexive(context.Background(), cs)
	if err != nil {
		t.Fatalf("EvaluateReflexive error: %v", err)
	}
	if rec.SelectedOutcome != OutcomeCommit {
		t.Errorf("expected SelectedOutcome COMMIT, got %s", rec.SelectedOutcome)
	}
	if rec.SelectedCandidateID != "cand-winner" {
		t.Errorf("expected selected candidate 'cand-winner', got '%s'", rec.SelectedCandidateID)
	}
}

func TestDefaultDecisionService_EvaluateReflexive_EscalateAmbiguity(t *testing.T) {
	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	// Create two candidates with near-identical utility scores (margin < 0.05)
	cs := CandidateSet{
		EpisodeID: "ep-escalate",
		Candidates: []Candidate{
			{
				ID:          "cand-a",
				Description: "Option A",
				Attributes:  map[string]float64{"utility": 2.000},
			},
			{
				ID:          "cand-b",
				Description: "Option B",
				Attributes:  map[string]float64{"utility": 1.990},
			},
		},
	}

	rec, err := srv.EvaluateReflexive(context.Background(), cs)
	if err != nil {
		t.Fatalf("EvaluateReflexive error: %v", err)
	}
	if rec.SelectedOutcome != OutcomeEscalateToDeliberative {
		t.Errorf("expected SelectedOutcome ESCALATE_TO_DELIBERATIVE, got %s", rec.SelectedOutcome)
	}
	if rec.EscalationRecommendation == nil {
		t.Fatal("expected non-nil EscalationRecommendation")
	}

	foundAmbiguity := false
	for _, dim := range rec.EscalationRecommendation.TriggeredDimensions {
		if dim == "AMBIGUITY_MARGIN" {
			foundAmbiguity = true
			break
		}
	}
	if !foundAmbiguity {
		t.Errorf("expected AMBIGUITY_MARGIN in triggered dimensions, got %v", rec.EscalationRecommendation.TriggeredDimensions)
	}
}

func TestDefaultDecisionService_EvaluateReflexive_AbstainOnConstitutionalVeto(t *testing.T) {
	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	cs := CandidateSet{
		EpisodeID: "ep-veto",
		Candidates: []Candidate{
			{
				ID:           "cand-unsafe-1",
				FlaggedRisks: []string{"SAFETY_VIOLATION"},
			},
		},
	}

	rec, err := srv.EvaluateReflexive(context.Background(), cs)
	if err != nil {
		t.Fatalf("EvaluateReflexive error: %v", err)
	}
	if rec.SelectedOutcome != OutcomeAbstain {
		t.Errorf("expected OutcomeAbstain when all candidates vetoed, got %s", rec.SelectedOutcome)
	}
}

func TestDefaultDecisionService_ConcurrentReflexiveEvaluations(t *testing.T) {
	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	const workers = 15
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cs := CandidateSet{
				EpisodeID: fmt.Sprintf("ep-concurrent-%d", id),
				Candidates: []Candidate{
					{
						ID:         fmt.Sprintf("cand-%d", id),
						Attributes: map[string]float64{"utility": float64(id + 1)},
					},
				},
			}
			_, err := srv.EvaluateReflexive(context.Background(), cs)
			if err != nil {
				t.Errorf("worker %d EvaluateReflexive error: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
}
