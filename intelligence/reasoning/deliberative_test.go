package reasoning

import (
	"context"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/inference"
)

type mockInferenceService struct {
	executed bool
}

func (m *mockInferenceService) Execute(ctx context.Context, req inference.InferenceRequest) (inference.InferenceResult, error) {
	m.executed = true
	return inference.InferenceResult{
		OutputRef:         "storage://reasoning/deliberative/res-1",
		ModelID:           "reasoning-deliberative-llm",
		BackendID:         "local-backend",
		ComputeUnits:      100,
		ExecutionDuration: 10 * time.Millisecond,
	}, nil
}

func (m *mockInferenceService) ExecuteStream(ctx context.Context, req inference.InferenceRequest, stream chan<- inference.StreamChunk) error {
	return nil
}

func (m *mockInferenceService) Name() string {
	return "mock.inference"
}

func (m *mockInferenceService) Start() error { return nil }
func (m *mockInferenceService) Close() error { return nil }

func TestDeliberativeSpecialist_NoEscalationWhenConfident(t *testing.T) {
	mock := &mockInferenceService{}
	specialist := NewDeliberativeSpecialist(mock)

	env := communication.Envelope{ID: "env-1", PayloadRef: "storage://test"}
	hyps, err := specialist.EvaluateDeliberative(context.Background(), env, 0.85, 0.65)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hyps) != 0 {
		t.Errorf("expected 0 escalated hypotheses when confident, got %d", len(hyps))
	}
	if mock.executed {
		t.Errorf("expected inference service not to be called when confident")
	}
}

func TestDeliberativeSpecialist_EscalatesWhenLowConfidence(t *testing.T) {
	mock := &mockInferenceService{}
	specialist := NewDeliberativeSpecialist(mock)

	env := communication.Envelope{ID: "env-low", PayloadRef: "storage://test"}
	hyps, err := specialist.EvaluateDeliberative(context.Background(), env, 0.45, 0.65)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hyps) != 1 {
		t.Fatalf("expected 1 escalated hypothesis, got %d", len(hyps))
	}
	if !mock.executed {
		t.Errorf("expected inference service to execute")
	}
	if hyps[0].Type != HypothesisDeliberative {
		t.Errorf("expected hypothesis type Deliberative, got %s", hyps[0].Type)
	}
}
