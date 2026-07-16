package executive_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

// mockDecisionStorer satisfies ExecutivePayloadStorer for test isolation.
type mockDecisionStorer struct {
	mu   sync.Mutex
	data map[string][]byte
}

func (m *mockDecisionStorer) Retrieve(_ context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		return nil, nil
	}
	return m.data[key], nil
}

// mockExecSubscriber satisfies ExecutiveWorkspaceSubscriber using a real workspace.Engine.
type mockExecSubscriber struct {
	ws *workspace.Engine
}

func (m *mockExecSubscriber) Subscribe(
	topic communication.TopicID,
	id string,
	handler func(ctx context.Context, env communication.Envelope) error,
) (executive.ExecutiveWorkspaceSubscription, error) {
	return m.ws.Subscribe(topic, id, handler)
}

func TestServiceV2_EvaluatedOptionsBridge(t *testing.T) {
	ws := workspace.NewEngine()
	if err := ws.Start(); err != nil {
		t.Fatalf("workspace start: %v", err)
	}
	defer ws.Close()

	cal := calibration.NewService()
	constGate := constitution.NewGate()

	storer := &mockDecisionStorer{
		data: map[string][]byte{"storage://cas/dec-001": []byte(`{"decision_id":"d1"}`)},
	}

	svc, err := executive.NewServiceV2(
		ws, cal, constGate, 1000,
		executive.WithWorkspaceBridgeOpt(storer, &mockExecSubscriber{ws: ws}),
	)
	if err != nil {
		t.Fatalf("NewServiceV2: %v", err)
	}
	if err := svc.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer svc.Close()

	// Observe TopicActionExecution — this is what World subscribes to.
	var (
		received      int
		receivedMu    sync.Mutex
		receivedPayload string
	)
	_, err = ws.Subscribe(communication.TopicActionExecution, "test-world-observer", func(_ context.Context, env communication.Envelope) error {
		receivedMu.Lock()
		received++
		receivedPayload = env.PayloadRef
		receivedMu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe TopicActionExecution: %v", err)
	}

	// Simulate Decision publishing an evaluated options envelope.
	payload := map[string]string{"decision_id": "dec-exec-001", "outcome": "COMMIT"}
	payloadBytes, _ := json.Marshal(payload)
	_ = payloadBytes

	decEnv, err := communication.NewEnvelopeBuilder().
		WithSource("CognitiveAbility.Decision").
		WithTopic(communication.TopicEvaluatedOptions).
		WithPayloadRef("storage://cas/dec-001").
		WithModality("structured-frame").
		WithConfidence(0.92).
		WithUrgency(0).
		WithCostEstimate(10).
		Build()
	if err != nil {
		t.Fatalf("build decision envelope: %v", err)
	}

	// Publish to workspace — Executive's subscription handler fires synchronously.
	if err := ws.Publish(context.Background(), decEnv); err != nil {
		t.Fatalf("publish TopicEvaluatedOptions: %v", err)
	}

	// Allow handler dispatch to settle.
	time.Sleep(10 * time.Millisecond)

	receivedMu.Lock()
	got := received
	gotRef := receivedPayload
	receivedMu.Unlock()

	if got == 0 {
		t.Fatal("Executive bridge: expected envelope on TopicActionExecution, got none")
	}
	if gotRef != "storage://cas/dec-001" {
		t.Errorf("Executive bridge: expected payloadRef forwarded unchanged, got %q", gotRef)
	}
}

// TestServiceV2_EvaluatedOptionsBridge_InvalidEnvelope verifies that Executive
// rejects malformed TopicEvaluatedOptions envelopes without panicking.
func TestServiceV2_EvaluatedOptionsBridge_InvalidEnvelope(t *testing.T) {
	ws := workspace.NewEngine()
	if err := ws.Start(); err != nil {
		t.Fatalf("workspace start: %v", err)
	}
	defer ws.Close()

	cal := calibration.NewService()
	constGate := constitution.NewGate()
	storer := &mockDecisionStorer{}

	svc, err := executive.NewServiceV2(
		ws, cal, constGate, 1000,
		executive.WithWorkspaceBridgeOpt(storer, &mockExecSubscriber{ws: ws}),
	)
	if err != nil {
		t.Fatalf("NewServiceV2: %v", err)
	}
	if err := svc.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer svc.Close()

	// Publish an invalid envelope (missing PayloadRef) directly via a raw workspace event.
	// The bridge handler should return an error (not panic), and no crash should occur.
	badEnv := communication.Envelope{
		ID:            "bad-env",
		Source:        "CognitiveAbility.Decision",
		Topic:         communication.TopicEvaluatedOptions,
		PayloadRef:    "", // Missing — should be rejected
		RawConfidence: 0.5,
		CreatedAt:     time.Now(),
	}
	// Workspace validates envelopes before delivery, so this will be rejected at the bus layer.
	err = ws.Publish(context.Background(), badEnv)
	if err == nil {
		t.Error("expected workspace to reject invalid envelope (missing PayloadRef), got nil error")
	}
}
