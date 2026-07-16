package decision

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"idun/intelligence/communication"
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
	if len(rec.RejectedCandidates) != 0 {
		t.Errorf("expected 0 rejected, got %d", len(rec.RejectedCandidates))
	}
}

type mockSub struct{}
func (m *mockSub) Cancel() error { return nil }

type mockSubscriber struct {
	handler func(ctx context.Context, env communication.Envelope) error
}
func (m *mockSubscriber) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error) {
	m.handler = handler
	return &mockSub{}, nil
}

type mockPayloadStorerBridge struct {
	data map[string][]byte
}
func (m *mockPayloadStorerBridge) Store(ctx context.Context, data []byte) (string, error) {
	return "storage://cas/test-key", nil
}
func (m *mockPayloadStorerBridge) Retrieve(ctx context.Context, key string) ([]byte, error) {
	if m.data == nil {
		return nil, fmt.Errorf("not found")
	}
	d, ok := m.data[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return d, nil
}

func TestService_CandidatePlansBridge(t *testing.T) {
	storer := &mockPayloadStorerBridge{data: make(map[string][]byte)}
	pub := &mockPublisher{}
	sub := &mockSubscriber{}

	srv := NewService(WithWorkspaceBridge(storer, pub, sub))
	_ = srv.Start()
	defer srv.Close()

	if sub.handler == nil {
		t.Fatal("expected handler to be registered")
	}

	payload := planningResultPayload{
		ResultID: "res-123",
		Plans: []*planPayload{
			{
				PlanID:        "plan-1",
				Goal:          "Test goal",
				Domain:        "Test domain",
				EstimatedCost: 10.0,
			},
		},
	}
	data, _ := json.Marshal(payload)
	storer.data["storage://cas/plan-ref"] = data

	env := communication.Envelope{
		ID:         "env-123",
		Topic:      communication.TopicCandidatePlans,
		PayloadRef: "storage://cas/plan-ref",
	}

	err := sub.handler(context.Background(), env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pub.mu.Lock()
	defer pub.mu.Unlock()
	if len(pub.envelopes) == 0 {
		t.Fatal("expected decision to be published to workspace")
	}

	outEnv := pub.envelopes[0]
	if outEnv.Topic != communication.TopicEvaluatedOptions {
		t.Errorf("expected output topic %s, got %s", communication.TopicEvaluatedOptions, outEnv.Topic)
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
