package reasoning

import (
	"context"
	"testing"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/infrastructure/embedding"
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
