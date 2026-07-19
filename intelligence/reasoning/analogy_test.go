package reasoning

import (
	"context"
	"testing"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/embedding"
	"idun/intelligence/understanding"
)

type mockEmbeddingService struct {
	similarity float64
}

func (m *mockEmbeddingService) Embed(ctx context.Context, req embedding.EmbeddingRequest) (embedding.EmbeddingResult, error) {
	return embedding.EmbeddingResult{}, nil
}

func (m *mockEmbeddingService) EmbedBatch(ctx context.Context, reqs []embedding.EmbeddingRequest) ([]embedding.EmbeddingResult, error) {
	return nil, nil
}

func (m *mockEmbeddingService) Similarity(ctx context.Context, vectorRefA, vectorRefB string) (float64, error) {
	return m.similarity, nil
}

func (m *mockEmbeddingService) Name() string { return "mock.embedding" }
func (m *mockEmbeddingService) Start() error { return nil }
func (m *mockEmbeddingService) Close() error { return nil }

func TestCaseAnalogySpecialist_EvaluateAnalogy(t *testing.T) {
	embedder := &mockEmbeddingService{similarity: 0.88}
	mem := &mockMemoryProvider{
		records: map[string][]memory.Record{
			"case": {
				{ID: "case/previous-incident", Type: "case"},
			},
		},
	}

	specialist := NewCaseAnalogySpecialist(embedder, mem)
	if specialist.ID() != StageS5CaseAnalogy {
		t.Errorf("expected stage ID S5, got %s", specialist.ID())
	}

	env := communication.Envelope{ID: "env-analogy", PayloadRef: "storage://query/1"}

	hyps, err := specialist.EvaluateAnalogy(context.Background(), env, nil, nil)
	if err != nil {
		t.Fatalf("expected analogy evaluation to succeed, got %v", err)
	}
	if len(hyps) != 1 {
		t.Fatalf("expected 1 analogical hypothesis, got %d", len(hyps))
	}

	if hyps[0].Type != HypothesisAnalogy {
		t.Errorf("expected hypothesis type Analogy, got %s", hyps[0].Type)
	}
	if hyps[0].ReasoningConfidence != 0.88 {
		t.Errorf("expected similarity 0.88, got %f", hyps[0].ReasoningConfidence)
	}
	if hyps[0].CalibratedConfidence != 0.0 {
		t.Errorf("expected CalibratedConfidence == 0 prior to Stage S7, got %f", hyps[0].CalibratedConfidence)
	}
	if err := hyps[0].Validate(); err != nil {
		t.Fatalf("generated analogical hypothesis failed validation: %v", err)
	}
}

func TestCaseAnalogySpecialist_EvaluateAnalogyWithPayloadAndAdaptation(t *testing.T) {
	embedder := &mockEmbeddingService{similarity: 0.85}
	payloadJSON := []byte(`{
		"intent": "deploy_service",
		"resolved_goal": {
			"kind": "TOOL_EXECUTION",
			"intent": "deploy_service",
			"target": "k8s_cluster",
			"desired_state": {"version": "v1.0.0", "region": "us-east-1"}
		},
		"parameter_slots": ["version"]
	}`)

	mem := &mockMemoryProvider{
		records: map[string][]memory.Record{
			"case": {
				{ID: "case/deploy-past", Type: "case", Payload: payloadJSON},
			},
		},
	}

	specialist := NewCaseAnalogySpecialist(embedder, mem)
	env := communication.Envelope{ID: "env-analogy-adapt", PayloadRef: "storage://query/deploy"}
	frame := &understanding.SemanticFrame{
		PrimaryHypothesis: understanding.Hypothesis{
			Intent: "deploy_service",
			Slots: []understanding.Slot{
				{Name: "version", Value: "v2.0.0"},
				{Name: "region", Value: "eu-west-1"},
			},
		},
	}

	hyps, err := specialist.EvaluateAnalogy(context.Background(), env, frame, nil)
	if err != nil {
		t.Fatalf("EvaluateAnalogy failed: %v", err)
	}
	if len(hyps) != 1 {
		t.Fatalf("expected 1 hypothesis, got %d", len(hyps))
	}
	hyp := hyps[0]
	if hyp.ProposedGoal == nil {
		t.Fatalf("expected ProposedGoal recovered from historical case payload")
	}
	if err := hyp.ProposedGoal.Validate(); err != nil {
		t.Fatalf("ProposedGoal validation failed: %v", err)
	}
	// Verify version adapted because it was in parameter_slots
	if hyp.ProposedGoal.DesiredState["version"] != "v2.0.0" {
		t.Errorf("expected adapted version v2.0.0, got %s", hyp.ProposedGoal.DesiredState["version"])
	}
	// Verify region NOT adapted because it was NOT in parameter_slots
	if hyp.ProposedGoal.DesiredState["region"] != "us-east-1" {
		t.Errorf("expected unadapted region us-east-1, got %s", hyp.ProposedGoal.DesiredState["region"])
	}
	// Verify conclusion remains diagnostic
	wantConclusionPrefix := "Analogical match for intent \"deploy_service\" from case case/deploy-past"
	if len(hyp.Conclusion) < len(wantConclusionPrefix) || hyp.Conclusion[:len(wantConclusionPrefix)] != wantConclusionPrefix {
		t.Errorf("expected diagnostic conclusion starting with %q, got %q", wantConclusionPrefix, hyp.Conclusion)
	}
}

func TestCaseAnalogySpecialist_EvaluateAnalogyMissingPayload(t *testing.T) {
	embedder := &mockEmbeddingService{similarity: 0.85}
	mem := &mockMemoryProvider{
		records: map[string][]memory.Record{
			"case": {
				{ID: "case/empty-payload", Type: "case"},
			},
		},
	}

	specialist := NewCaseAnalogySpecialist(embedder, mem)
	env := communication.Envelope{ID: "env-analogy-empty", PayloadRef: "storage://query/empty"}

	hyps, err := specialist.EvaluateAnalogy(context.Background(), env, nil, nil)
	if err != nil {
		t.Fatalf("EvaluateAnalogy failed: %v", err)
	}
	if len(hyps) != 1 {
		t.Fatalf("expected 1 hypothesis, got %d", len(hyps))
	}
	hyp := hyps[0]
	if hyp.ProposedGoal != nil {
		t.Errorf("expected nil ProposedGoal when payload empty, got %+v", hyp.ProposedGoal)
	}
	hasMissingArtifact := false
	for _, p := range hyp.SupportingPremises {
		if len(p) > len("missing_learning_artifact=") && p[:len("missing_learning_artifact=")] == "missing_learning_artifact=" {
			hasMissingArtifact = true
			break
		}
	}
	if !hasMissingArtifact {
		t.Errorf("expected SupportingPremises to contain missing_learning_artifact, got %v", hyp.SupportingPremises)
	}
}
