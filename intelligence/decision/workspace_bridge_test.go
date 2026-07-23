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

func (m *mockStorer) Retrieve(ctx context.Context, key string) ([]byte, error) {
	return m.storedBytes, nil
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
	_, err := PublishDeliberativeDecision(ctx, reflexiveRec, nil, storer, publisher)
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

	env, err := PublishDeliberativeDecision(ctx, deliberativeRec, nil, storer, publisher)
	if err != nil {
		t.Fatalf("PublishDeliberativeDecision error: %v", err)
	}

	if env.PayloadRef != "storage://cas/test-dec-ref" {
		t.Errorf("expected payload ref 'storage://cas/test-dec-ref', got '%s'", env.PayloadRef)
	}
	if len(publisher.envelopes) != 1 {
		t.Fatalf("expected exactly 1 published envelope, got %d", len(publisher.envelopes))
	}

	// 3. Verify Deliberative record with selectedDesc stores boundary.CommunicationMessage
	envWithDesc, err := PublishDeliberativeDecision(ctx, deliberativeRec, &Candidate{Description: "Hello, IDUN!"}, storer, publisher, "env-parent-1")
	if err != nil {
		t.Fatalf("PublishDeliberativeDecision with selectedDesc error: %v", err)
	}
	if envWithDesc.PayloadRef != "storage://cas/test-dec-ref" {
		t.Errorf("expected payload ref 'storage://cas/test-dec-ref', got '%s'", envWithDesc.PayloadRef)
	}
	if envWithDesc.PayloadModality != "communication-message" {
		t.Errorf("expected payload modality 'communication-message', got '%s'", envWithDesc.PayloadModality)
	}
	if len(publisher.envelopes) != 2 {
		t.Fatalf("expected 2 published envelopes, got %d", len(publisher.envelopes))
	}
}

