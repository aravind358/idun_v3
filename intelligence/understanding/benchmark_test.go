package understanding_test

import (
	"context"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/understanding"
)

func BenchmarkNormalizer(b *testing.B) {
	norm := understanding.NewDefaultNormalizer()
	input := "   Can we please RESCHEDULE the meeting with her tomorrow afternoon?   "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = norm.Normalize(input)
	}
}

func BenchmarkGrammarSpecialist(b *testing.B) {
	grammar := understanding.NewDefaultGrammarSpecialist()
	norm := understanding.NewDefaultNormalizer().Normalize("set alarm for 07:00")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = grammar.Evaluate(norm)
	}
}

func BenchmarkNeuralSpecialist(b *testing.B) {
	neural := understanding.NewDefaultNeuralSpecialist()
	norm := understanding.NewDefaultNormalizer().Normalize("can we move the meeting to next week?")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = neural.Evaluate(norm)
	}
}

func BenchmarkMergeHypothesesByIntent(b *testing.B) {
	candidates := []understanding.Hypothesis{
		{
			Intent:               "book_flight",
			CalibratedConfidence: 0.90,
			SourceLayer:          understanding.LayerReflexiveGrammar,
			Slots: []understanding.Slot{
				{Name: "destination", Value: "SEA", Confidence: 0.95},
			},
		},
		{
			Intent:               "book_flight",
			CalibratedConfidence: 0.88,
			SourceLayer:          understanding.LayerNeuralClassifier,
			Slots: []understanding.Slot{
				{Name: "airline", Value: "Alaska", Confidence: 0.92},
				{Name: "date", Value: "tomorrow", Confidence: 0.80},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = understanding.MergeHypothesesByIntent(candidates)
	}
}

func BenchmarkSpeculativeEvaluator(b *testing.B) {
	evaluator := understanding.NewSpeculativeEvaluator(0.15)
	grammar := understanding.NewDefaultGrammarSpecialist()
	neural := understanding.NewDefaultNeuralSpecialist()
	norm := understanding.NewDefaultNormalizer().Normalize("Can we reschedule the meeting?")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = evaluator.EvaluateParallel(ctx, norm, grammar, neural, nil)
	}
}

func BenchmarkSemanticFrameBuilderAndValidate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame, err := understanding.NewSemanticFrameBuilder("env-bench").
			WithPrimaryHypothesis("query_status", 0.95, understanding.LayerReflexiveGrammar).
			WithStatus(understanding.StatusUnambiguous).
			WithProcessedDuration(0.5).
			Build()
		if err != nil {
			b.Fatalf("build: %v", err)
		}
		_ = frame.Validate()
	}
}

func BenchmarkServiceInterpretEnvelope(b *testing.B) {
	svc := understanding.NewService(understanding.WithConfigOptions(), nil)
	env := communication.Envelope{ID: "bench-env", PayloadRef: "status"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.InterpretEnvelope(ctx, env)
	}
}

func BenchmarkDeliberativeWorkerMocked(b *testing.B) {
	infSvc := &mockInferenceService{
		outputRef: makeValidDeliberativeJSON("synthesize_plan", 0.92),
	}
	worker := understanding.NewDeliberativeWorker(infSvc, nil, 2*time.Second)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = worker.InterpretDeliberative(ctx, "bench-delib", "complex input", "")
	}
}
