package understanding_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/inference"
	"idun/intelligence/understanding"
	"idun/intelligence/workspace"
)

type mockInferenceService struct {
	mu        sync.Mutex
	outputRef string
	err       error
	delay     time.Duration
}

func (m *mockInferenceService) Execute(ctx context.Context, req inference.InferenceRequest) (inference.InferenceResult, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return inference.InferenceResult{}, ctx.Err()
		}
	}
	if m.err != nil {
		return inference.InferenceResult{}, m.err
	}
	return inference.InferenceResult{
		OutputRef: m.outputRef,
		ModelID:   req.ModelID,
	}, nil
}

func (m *mockInferenceService) ExecuteStream(ctx context.Context, req inference.InferenceRequest, stream chan<- inference.StreamChunk) error {
	return nil
}

func (m *mockInferenceService) Name() string { return "MockInference" }
func (m *mockInferenceService) Start() error { return nil }
func (m *mockInferenceService) Close() error { return nil }
func (m *mockInferenceService) ClearCache() error { return nil }

type mockWorkspaceSubscription struct {
	id string
}

func (s *mockWorkspaceSubscription) ID() workspace.SubscriptionID     { return workspace.SubscriptionID(s.id) }
func (s *mockWorkspaceSubscription) Topic() communication.TopicID     { return communication.TopicImpasses }
func (s *mockWorkspaceSubscription) Subscriber() string               { return "Understanding.DeliberativeWorker" }
func (s *mockWorkspaceSubscription) Cancel() error                    { return nil }

type mockDeliberativeWorkspace struct {
	mu            sync.Mutex
	published     []communication.Envelope
	impasseHandler workspace.EnvelopeHandler
}

func (m *mockDeliberativeWorkspace) Publish(ctx context.Context, env communication.Envelope, opts ...workspace.PublishOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, env)
	return nil
}

func (m *mockDeliberativeWorkspace) Subscribe(topic communication.TopicID, subscriberID string, handler workspace.EnvelopeHandler) (workspace.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if topic == communication.TopicImpasses {
		m.impasseHandler = handler
	}
	return &mockWorkspaceSubscription{id: "sub-1"}, nil
}

func (m *mockDeliberativeWorkspace) SubscribeAll(subscriberID string, handler workspace.EnvelopeHandler) ([]workspace.Subscription, error) {
	return nil, nil
}

func (m *mockDeliberativeWorkspace) Unsubscribe(id workspace.SubscriptionID) error { return nil }
func (m *mockDeliberativeWorkspace) GetEnvelope(id string) (communication.Envelope, bool) {
	return communication.Envelope{}, false
}
func (m *mockDeliberativeWorkspace) ListTopicEnvelopes(topic communication.TopicID, limit int) []communication.Envelope {
	return nil
}
func (m *mockDeliberativeWorkspace) StorePendingCandidate(ctx context.Context, topic communication.TopicID, candidate workspace.PendingCandidate) error {
	return nil
}
func (m *mockDeliberativeWorkspace) GetPendingCandidates(topic communication.TopicID) []workspace.PendingCandidate {
	return nil
}
func (m *mockDeliberativeWorkspace) RemovePendingCandidate(topic communication.TopicID, envelopeID string) bool {
	return false
}
func (m *mockDeliberativeWorkspace) RegisterEpisodeDependencies(ctx context.Context, epID string, dependsOn []string) error {
	return nil
}
func (m *mockDeliberativeWorkspace) IsEpisodeReady(ctx context.Context, epID string) (bool, error) {
	return true, nil
}
func (m *mockDeliberativeWorkspace) ResolveDependencies(ctx context.Context, epID string) error {
	return nil
}
func (m *mockDeliberativeWorkspace) NotifyDependencyComplete(ctx context.Context, depID string) error {
	return nil
}
func (m *mockDeliberativeWorkspace) RegisterEpisodeChild(ctx context.Context, parentID string, childID string) error {
	return nil
}
func (m *mockDeliberativeWorkspace) GetEpisodeChildren(ctx context.Context, parentID string) ([]string, error) {
	return nil, nil
}
func (m *mockDeliberativeWorkspace) Name() string { return "MockDeliberativeWorkspace" }
func (m *mockDeliberativeWorkspace) Start() error { return nil }
func (m *mockDeliberativeWorkspace) Close() error { return nil }

func makeValidDeliberativeJSON(intent string, conf float64) string {
	frame := understanding.SemanticFrame{
		FrameVersion: "2.0",
		EnvelopeID:   "env-delib",
		Status:       understanding.StatusUnambiguous,
		PrimaryHypothesis: understanding.Hypothesis{
			Intent:               intent,
			CalibratedConfidence: conf,
			SourceLayer:          understanding.LayerDeliberativeLLM,
			Slots: []understanding.Slot{
				{Name: "topic", Value: "deep_semantics", Confidence: 0.90},
			},
		},
		AmbiguitySet:        []understanding.Hypothesis{},
		ProcessedDurationMs: 12.5,
	}
	bytes, _ := json.Marshal(frame)
	return string(bytes)
}

func TestDeliberativeWorker_SuccessfulInterpretation(t *testing.T) {
	infSvc := &mockInferenceService{
		outputRef: makeValidDeliberativeJSON("synthesize_research_plan", 0.91),
	}
	worker := understanding.NewDeliberativeWorker(infSvc, nil, 2*time.Second)

	frame, err := worker.InterpretDeliberative(context.Background(), "env-1", "complex input needing LLM", "")
	if err != nil {
		t.Fatalf("InterpretDeliberative failed: %v", err)
	}

	if frame.PrimaryHypothesis.Intent != "synthesize_research_plan" {
		t.Fatalf("expected intent synthesize_research_plan, got %s", frame.PrimaryHypothesis.Intent)
	}
	if frame.PrimaryHypothesis.SourceLayer != understanding.LayerDeliberativeLLM {
		t.Fatalf("expected source layer LayerDeliberativeLLM, got %s", frame.PrimaryHypothesis.SourceLayer)
	}
}

func TestDeliberativeWorker_TimeoutAndCancellation(t *testing.T) {
	// 1. Timeout handling
	infSvcSlow := &mockInferenceService{
		delay: 100 * time.Millisecond,
	}
	workerTimeout := understanding.NewDeliberativeWorker(infSvcSlow, nil, 10*time.Millisecond)

	_, err := workerTimeout.InterpretDeliberative(context.Background(), "env-slow", "input", "")
	if !errors.Is(err, understanding.ErrInferenceTimeout) {
		t.Fatalf("expected ErrInferenceTimeout, got: %v", err)
	}

	// 2. Explicit cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = workerTimeout.InterpretDeliberative(ctx, "env-cancel", "input", "")
	if !errors.Is(err, understanding.ErrDeliberativeCancelled) {
		t.Fatalf("expected ErrDeliberativeCancelled, got: %v", err)
	}
}

func TestDeliberativeWorker_MalformedResponses(t *testing.T) {
	testCases := []struct {
		name      string
		outputRef string
	}{
		{
			name:      "free_form_text",
			outputRef: "Sure, here is your interpretation: book_flight",
		},
		{
			name: "missing_required_intent",
			outputRef: `{"FrameVersion":"2.0","EnvelopeID":"env-bad","Status":"UNAMBIGUOUS","PrimaryHypothesis":{"Intent":"","CalibratedConfidence":0.8,"SourceLayer":"DELIBERATIVE_LLM"}}`,
		},
		{
			name: "out_of_bounds_confidence_high",
			outputRef: `{"FrameVersion":"2.0","EnvelopeID":"env-bad","Status":"UNAMBIGUOUS","PrimaryHypothesis":{"Intent":"book_flight","CalibratedConfidence":1.5,"SourceLayer":"DELIBERATIVE_LLM"}}`,
		},
		{
			name: "out_of_bounds_confidence_negative",
			outputRef: `{"FrameVersion":"2.0","EnvelopeID":"env-bad","Status":"UNAMBIGUOUS","PrimaryHypothesis":{"Intent":"book_flight","CalibratedConfidence":-0.2,"SourceLayer":"DELIBERATIVE_LLM"}}`,
		},
		{
			name: "unknown_hallucinated_fields",
			outputRef: `{"FrameVersion":"2.0","EnvelopeID":"env-bad","Status":"UNAMBIGUOUS","UnknownField":"Hallucinated","PrimaryHypothesis":{"Intent":"book_flight","CalibratedConfidence":0.8,"SourceLayer":"DELIBERATIVE_LLM"}}`,
		},
		{
			name: "multiple_trailing_objects",
			outputRef: `{"FrameVersion":"2.0","EnvelopeID":"env-bad","Status":"UNAMBIGUOUS","PrimaryHypothesis":{"Intent":"book_flight","CalibratedConfidence":0.8,"SourceLayer":"DELIBERATIVE_LLM"}}{"Extra":"object"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			infSvc := &mockInferenceService{outputRef: tc.outputRef}
			worker := understanding.NewDeliberativeWorker(infSvc, nil, 1*time.Second)
			_, err := worker.InterpretDeliberative(context.Background(), "env-bad", "input", "")
			if !errors.Is(err, understanding.ErrMalformedInferenceResponse) {
				t.Fatalf("case %s: expected ErrMalformedInferenceResponse, got %v", tc.name, err)
			}
		})
	}
}

func TestService_EscalatesToDeliberativeWorker(t *testing.T) {
	infSvc := &mockInferenceService{
		outputRef: makeValidDeliberativeJSON("deliberative_fallback_intent", 0.89),
	}
	worker := understanding.NewDeliberativeWorker(infSvc, nil, 2*time.Second)

	svc := understanding.NewService(
		understanding.WithConfigOptions(),
		nil,
		understanding.WithDeliberativeWorker(worker),
	)

	// An utterance that does not match grammar or local neural patterns -> triggers escalation
	frame, err := svc.InterpretEnvelope(context.Background(), communication.Envelope{
		ID:         "env-escalate",
		PayloadRef: "utterance that no local specialist recognizes",
	})
	if err != nil {
		t.Fatalf("InterpretEnvelope failed: %v", err)
	}

	if frame.PrimaryHypothesis.Intent != "deliberative_fallback_intent" {
		t.Fatalf("expected fallback intent from deliberative worker, got %s", frame.PrimaryHypothesis.Intent)
	}
	if frame.PrimaryHypothesis.SourceLayer != understanding.LayerDeliberativeLLM {
		t.Fatalf("expected source layer LayerDeliberativeLLM, got %s", frame.PrimaryHypothesis.SourceLayer)
	}
}

func TestDeliberativeWorker_WorkspaceSubscription(t *testing.T) {
	ws := &mockDeliberativeWorkspace{}
	infSvc := &mockInferenceService{
		outputRef: makeValidDeliberativeJSON("impasse_resolved_intent", 0.94),
	}
	worker := understanding.NewDeliberativeWorker(infSvc, ws, 2*time.Second)

	if err := worker.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer worker.Close()

	ws.mu.Lock()
	handler := ws.impasseHandler
	ws.mu.Unlock()

	if handler == nil {
		t.Fatalf("expected handler subscribed to TopicImpasses")
	}

	// Trigger impasse envelope delivery
	impasseEnv := communication.Envelope{
		ID:         "impasse-100",
		PayloadRef: "complex impasse input",
	}
	if err := handler(context.Background(), impasseEnv); err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	ws.mu.Lock()
	pubCount := len(ws.published)
	ws.mu.Unlock()
	if pubCount != 1 {
		t.Fatalf("expected 1 published envelope to TopicUserIntent, got %d", pubCount)
	}
}
