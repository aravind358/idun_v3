package decision

import (
	"context"
	"testing"
)

func TestDecisionBuilder_BuildAndEvaluate(t *testing.T) {
	builder := NewDecisionBuilder().
		WithConfig(DefaultDecisionConfig()).
		WithConstitutionalGate(NewDefaultConstitutionalGate()).
		WithObjectiveScorer(NewDefaultObjectiveScorer())

	service, err := builder.Build()
	if err != nil {
		t.Fatalf("DecisionBuilder.Build error: %v", err)
	}

	cset := CandidateSet{
		EpisodeID: "ep-builder-test",
		Candidates: []Candidate{
			{
				ID: "c1",
				Attributes: map[string]float64{
					"utility": 0.9,
					"safety":  1.0,
				},
			},
		},
	}

	rec, err := service.EvaluateReflexive(context.Background(), cset)
	if err != nil {
		t.Fatalf("EvaluateReflexive error: %v", err)
	}

	if rec.SelectedOutcome != OutcomeCommit {
		t.Errorf("expected OutcomeCommit, got %v", rec.SelectedOutcome)
	}
}
