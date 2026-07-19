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
func (m *mockInferenceService) ClearCache() error { return nil }

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

type mockPayloadStorer struct {
	data map[string][]byte
}

func (m *mockPayloadStorer) Store(ctx context.Context, data []byte) (string, error) {
	return "storage://mock/1", nil
}

func (m *mockPayloadStorer) Retrieve(ctx context.Context, key string) ([]byte, error) {
	if m.data != nil {
		if val, ok := m.data[key]; ok {
			return val, nil
		}
	}
	return nil, nil
}

type mockInferenceServiceCustomRef struct {
	outputRef string
}

func (m *mockInferenceServiceCustomRef) Execute(ctx context.Context, req inference.InferenceRequest) (inference.InferenceResult, error) {
	return inference.InferenceResult{
		OutputRef:         m.outputRef,
		ModelID:           "reasoning-deliberative-llm",
		BackendID:         "local-backend",
		ComputeUnits:      100,
		ExecutionDuration: 10 * time.Millisecond,
	}, nil
}

func (m *mockInferenceServiceCustomRef) ExecuteStream(ctx context.Context, req inference.InferenceRequest, stream chan<- inference.StreamChunk) error {
	return nil
}

func (m *mockInferenceServiceCustomRef) Name() string { return "mock.inference.custom" }
func (m *mockInferenceServiceCustomRef) Start() error { return nil }
func (m *mockInferenceServiceCustomRef) Close() error { return nil }
func (m *mockInferenceServiceCustomRef) ClearCache() error { return nil }

func TestDeliberativeSpecialist_StructuredJSONParsingAndValidation(t *testing.T) {
	goalJSON := []byte(`{"kind":"COMMUNICATIVE","intent":"explain_concept","target":"user","desired_state":{"explained":"true"}}`)
	storer := &mockPayloadStorer{
		data: map[string][]byte{
			"storage://reasoning/deliberative/res-valid": goalJSON,
		},
	}
	mock := &mockInferenceServiceCustomRef{outputRef: "storage://reasoning/deliberative/res-valid"}
	specialist := NewDeliberativeSpecialist(mock)

	env := communication.Envelope{ID: "env-delib-valid", PayloadRef: "storage://test"}
	hyps, err := specialist.EvaluateDeliberative(context.Background(), env, 0.40, 0.65, storer)
	if err != nil {
		t.Fatalf("EvaluateDeliberative failed: %v", err)
	}
	if len(hyps) != 1 {
		t.Fatalf("expected 1 hypothesis, got %d", len(hyps))
	}
	hyp := hyps[0]
	if hyp.ProposedGoal == nil {
		t.Fatalf("expected ProposedGoal parsed from CAS storage via storer")
	}
	if err := hyp.ProposedGoal.Validate(); err != nil {
		t.Fatalf("ProposedGoal failed validation: %v", err)
	}
	if hyp.ProposedGoal.Kind != GoalKindCommunicative || hyp.ProposedGoal.Intent != "explain_concept" {
		t.Errorf("unexpected ProposedGoal: %+v", hyp.ProposedGoal)
	}
	hasProvenance := false
	for _, p := range hyp.SupportingPremises {
		if p == "provenance=LLM_FALLBACK" {
			hasProvenance = true
			break
		}
	}
	if !hasProvenance {
		t.Errorf("expected SupportingPremises to contain provenance=LLM_FALLBACK, got %v", hyp.SupportingPremises)
	}
	if hyp.Conclusion != "Deliberative structured synthesis for env-delib-valid via storage://reasoning/deliberative/res-valid" {
		t.Errorf("expected diagnostic conclusion preserved, got %q", hyp.Conclusion)
	}
}

func TestDeliberativeSpecialist_MalformedJSONRejection(t *testing.T) {
	malformedJSON := []byte(`{"kind":"INVALID_KIND","intent":""}`)
	storer := &mockPayloadStorer{
		data: map[string][]byte{
			"storage://reasoning/deliberative/res-malformed": malformedJSON,
		},
	}
	mock := &mockInferenceServiceCustomRef{outputRef: "storage://reasoning/deliberative/res-malformed"}
	specialist := NewDeliberativeSpecialist(mock)

	env := communication.Envelope{ID: "env-delib-malformed", PayloadRef: "storage://test"}
	hyps, err := specialist.EvaluateDeliberative(context.Background(), env, 0.40, 0.65, storer)
	if err != nil {
		t.Fatalf("EvaluateDeliberative failed: %v", err)
	}
	if len(hyps) != 1 {
		t.Fatalf("expected 1 hypothesis, got %d", len(hyps))
	}
	hyp := hyps[0]
	if hyp.ProposedGoal != nil {
		t.Errorf("expected ProposedGoal == nil when JSON is malformed/invalid, got %+v", hyp.ProposedGoal)
	}
	hasRejectionStatus := false
	for _, p := range hyp.SupportingPremises {
		if p == "status=rejected_malformed_goal" {
			hasRejectionStatus = true
			break
		}
	}
	if !hasRejectionStatus {
		t.Errorf("expected SupportingPremises to contain status=rejected_malformed_goal, got %v", hyp.SupportingPremises)
	}
	if hyp.Conclusion != "Deliberative structured synthesis for env-delib-malformed via storage://reasoning/deliberative/res-malformed" {
		t.Errorf("expected diagnostic conclusion preserved, got %q", hyp.Conclusion)
	}
}
