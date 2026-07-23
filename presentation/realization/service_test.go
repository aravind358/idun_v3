package realization_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/inference"
	"idun/presentation/realization"
)

type mockSub struct{}

func (m *mockSub) Cancel() error { return nil }

type mockWorkspace struct {
	mu        sync.Mutex
	handlers  map[communication.TopicID][]func(context.Context, communication.Envelope) error
	published []communication.Envelope
}

func newMockWorkspace() *mockWorkspace {
	return &mockWorkspace{
		handlers: make(map[communication.TopicID][]func(context.Context, communication.Envelope) error),
	}
}

func (m *mockWorkspace) Subscribe(topic communication.TopicID, subscriberID string, handler func(context.Context, communication.Envelope) error) (realization.WorkspaceSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[topic] = append(m.handlers[topic], handler)
	return &mockSub{}, nil
}

func (m *mockWorkspace) Publish(ctx context.Context, env communication.Envelope) error {
	m.mu.Lock()
	m.published = append(m.published, env)
	handlers := append([]func(context.Context, communication.Envelope) error(nil), m.handlers[env.Topic]...)
	m.mu.Unlock()

	for _, h := range handlers {
		_ = h(ctx, env)
	}
	return nil
}

type mockStorer struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func newMockStorer() *mockStorer {
	return &mockStorer{data: make(map[string][]byte)}
}

func (m *mockStorer) Store(ctx context.Context, data []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := "store://" + string(data)[:min(len(data), 10)]
	cp := make([]byte, len(data))
	copy(cp, data)
	m.data[key] = cp
	return key, nil
}

func (m *mockStorer) Retrieve(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := make([]byte, len(val))
	copy(cp, val)
	return cp, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type mockInference struct {
	executeFunc func(ctx context.Context, req inference.InferenceRequest) (inference.InferenceResult, error)
}

func (m *mockInference) Execute(ctx context.Context, req inference.InferenceRequest) (inference.InferenceResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return inference.InferenceResult{
		ModelID:           req.ModelID,
		OutputRef:         "realized: output text",
		ExecutionDuration: 10 * time.Millisecond,
	}, nil
}

func (m *mockInference) ExecuteStream(ctx context.Context, req inference.InferenceRequest, stream chan<- inference.StreamChunk) error {
	return nil
}

func (m *mockInference) Name() string { return "MockInference" }
func (m *mockInference) Start() error { return nil }
func (m *mockInference) Close() error { return nil }
func (m *mockInference) ClearCache() error { return nil }

func TestBuildRealizationPrompt(t *testing.T) {
	resp := realization.ExecutionResponse{
		ResponseID:       "resp-1",
		ParentRef:        "parent-1",
		FinalizedContent: "Action: Deploy to staging.",
		Tone:             realization.ToneConversational,
		Language:         "en-US",
	}

	prompt := realization.BuildRealizationPromptFromLegacy(resp)
	if !strings.Contains(prompt, "Action: Deploy to staging.") {
		t.Fatalf("prompt missing content")
	}
	if !strings.Contains(prompt, "conversational") {
		t.Fatalf("prompt missing tone")
	}
}

func TestServiceLifecycle(t *testing.T) {
	ws := newMockWorkspace()
	storer := newMockStorer()
	inf := &mockInference{}

	svc, err := realization.NewServiceBuilder().
		WithWorkspace(ws).
		WithInference(inf).
		WithStorage(storer).
		WithConfig(realization.DefaultConfig()).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if svc.Name() != "Presentation.LanguageRealization" {
		t.Fatalf("unexpected name: %s", svc.Name())
	}

	if err := svc.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := svc.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestServiceRealizationPass(t *testing.T) {
	ws := newMockWorkspace()
	storer := newMockStorer()
	inf := &mockInference{
		executeFunc: func(ctx context.Context, req inference.InferenceRequest) (inference.InferenceResult, error) {
			key, _ := storer.Store(ctx, []byte("Hello, the action has been scheduled successfully."))
			return inference.InferenceResult{
				OutputRef: key,
			}, nil
		},
	}

	svc, _ := realization.NewService(ws, inf, storer, realization.DefaultConfig())
	_ = svc.Start(context.Background())
	defer svc.Close()

	resp := realization.ExecutionResponse{
		ResponseID:       "resp-42",
		ParentRef:        "parent-42",
		FinalizedContent: "Scheduled action successfully.",
		Tone:             realization.ToneProfessional,
		Language:         "en-US",
	}
	respBytes, _ := json.Marshal(resp)
	respRef, _ := storer.Store(context.Background(), respBytes)

	env, _ := communication.NewEnvelopeBuilder().
		WithSource("Intelligence.Executive").
		WithTopic(communication.TopicActionExecution).
		WithParentRef("parent-42").
		WithPayloadRef(respRef).
		Build()

	_ = ws.Publish(context.Background(), env)

	ws.mu.Lock()
	published := ws.published
	ws.mu.Unlock()

	var found bool
	for _, p := range published {
		if p.Source == "Presentation.LanguageRealization" {
			found = true
			if p.ParentRef != "parent-42" {
				t.Fatalf("expected ParentRef parent-42, got %s", p.ParentRef)
			}
			outBytes, _ := storer.Retrieve(context.Background(), p.PayloadRef)
			var rOut realization.RealizedOutput
			if err := json.Unmarshal(outBytes, &rOut); err != nil {
				t.Fatalf("failed to unmarshal RealizedOutput: %v", err)
			}
			if rOut.RealizedText != "Hello, the action has been scheduled successfully." {
				t.Fatalf("unexpected text: %s", rOut.RealizedText)
			}
		}
	}
	if !found {
		t.Fatalf("expected realization output envelope published to workspace")
	}
}
