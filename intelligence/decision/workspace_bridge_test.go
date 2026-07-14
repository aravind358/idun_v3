package decision

import (
	"context"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
)

type mockStorer struct {
	storedBytes []byte
}

func (m *mockStorer) Store(ctx context.Context, data []byte) (string, error) {
	m.storedBytes = data
	return "storage://cas/test-dec-ref", nil
}

type mockPublisher struct {
	mu        sync.Mutex
	envelopes []communication.Envelope
}

func (m *mockPublisher) Publish(ctx context.Context, env communication.Envelope) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.envelopes = append(m.envelopes, env)
	return nil
}

func TestShouldPublishToWorkspace(t *testing.T) {
	if ShouldPublishToWorkspace(DepthReflexive) {
		t.Error("expected ShouldPublishToWorkspace(DepthReflexive) to be false")
	}
	if !ShouldPublishToWorkspace(DepthDeliberative) {
		t.Error("expected ShouldPublishToWorkspace(DepthDeliberative) to be true")
	}
}

func TestPublishDeliberativeDecision_SuccessAndRejection(t *testing.T) {
	ctx := context.Background()
	storer := &mockStorer{}
	publisher := &mockPublisher{}

	// 1. Verify Reflexive record is rejected from workspace publishing
	reflexiveRec := &DecisionRecord{
		DecisionID:        "dec-ref",
		DeliberationDepth: DepthReflexive,
	}
	_, err := PublishDeliberativeDecision(ctx, reflexiveRec, storer, publisher)
	if err == nil {
		t.Error("expected error when publishing reflexive record to workspace, got nil")
	}

	// 2. Verify Deliberative record is published successfully
	deliberativeRec := &DecisionRecord{
		DecisionID:          "dec-delib",
		EpisodeID:           "ep-delib",
		SchemaVersion:       "2.0.0-FROZEN",
		Timestamp:           time.Now(),
		StrategyVersion:     "v2.0.0",
		DeliberationDepth:   DepthDeliberative,
		SelectedOutcome:     OutcomeCommit,
		SelectedCandidateID: "cand-1",
		Confidence:          0.95,
	}

	env, err := PublishDeliberativeDecision(ctx, deliberativeRec, storer, publisher)
	if err != nil {
		t.Fatalf("PublishDeliberativeDecision error: %v", err)
	}

	if env.PayloadRef != "storage://cas/test-dec-ref" {
		t.Errorf("expected payload ref 'storage://cas/test-dec-ref', got '%s'", env.PayloadRef)
	}
	if len(publisher.envelopes) != 1 {
		t.Fatalf("expected exactly 1 published envelope, got %d", len(publisher.envelopes))
	}
}
