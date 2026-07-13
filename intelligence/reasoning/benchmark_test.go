package reasoning

import (
	"context"
	"testing"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/constitution"
)

func benchMemoryRecords() []memory.Record {
	return []memory.Record{
		{ID: "bel-role-admin", Type: "belief"},
		{ID: "case-prev-1", Type: "case"},
	}
}

func BenchmarkSymbolicReasoning(b *testing.B) {
	specialist := NewSymbolicSpecialist()
	ctx := context.Background()
	env := communication.Envelope{ID: "bench-env", Source: "bench"}
	records := benchMemoryRecords()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = specialist.Evaluate(ctx, env, nil, records)
	}
}

func BenchmarkSessionGraphOperations(b *testing.B) {
	spec := SelectStrategyForID(StrategyGraphDeliberative)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		graph := NewSessionGraph(spec.MaxGraphNodes, spec.MaxGraphEdges, spec.MaxGraphDepth)
		_ = graph.AddNode("A", "concept", "node A")
		_ = graph.AddNode("B", "concept", "node B")
		_ = graph.AddNode("C", "concept", "node C")
		_ = graph.AddEdge("A", "B", "leads_to", 0.9)
		_ = graph.AddEdge("B", "C", "implies", 0.85)
		_, _ = graph.TraverseBounded("A", 3)
		graph.Clear()
	}
}

func BenchmarkBayesianEvidenceFusion(b *testing.B) {
	specialist := NewBayesianFusionSpecialist()
	ctx := context.Background()
	hyps := []ReasoningHypothesis{
		{ID: "h1", ReasoningConfidence: 0.8},
		{ID: "h2", ReasoningConfidence: 0.6},
		{ID: "h3", ReasoningConfidence: 0.9},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = specialist.FuseEvidence(ctx, hyps)
	}
}

func BenchmarkCaseAnalogy(b *testing.B) {
	specialist := NewCaseAnalogySpecialist(nil, nil)
	ctx := context.Background()
	env := communication.Envelope{ID: "bench-analogy"}
	records := benchMemoryRecords()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = specialist.EvaluateAnalogy(ctx, env, nil, records)
	}
}

func BenchmarkBeamSelection(b *testing.B) {
	specialist := NewBeamSelectionSpecialist()
	hyps := []ReasoningHypothesis{
		{ID: "h1", ReasoningConfidence: 0.92},
		{ID: "h2", ReasoningConfidence: 0.89},
		{ID: "h3", ReasoningConfidence: 0.85},
		{ID: "h4", ReasoningConfidence: 0.40},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = specialist.SelectBeam(hyps, 3, 0.25)
	}
}

func BenchmarkCalibrationIntegration(b *testing.B) {
	mockCal := &mockCalibService{multiplier: 0.95}
	specialist := NewCalibrationSpecialist(mockCal)
	ctx := context.Background()
	primary := ReasoningHypothesis{ID: "p1", ReasoningConfidence: 0.9}
	beam := []ReasoningHypothesis{{ID: "b1", ReasoningConfidence: 0.85}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = specialist.CalibrateHypotheses(ctx, "bench", communication.TopicActiveGoals, primary, beam)
	}
}

func BenchmarkDeliberativeWorkerMocked(b *testing.B) {
	mockInf := &mockInferenceService{}
	specialist := NewDeliberativeSpecialist(mockInf)
	ctx := context.Background()
	env := communication.Envelope{ID: "bench-delib", PayloadRef: "storage://test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = specialist.EvaluateDeliberative(ctx, env, 0.50, 0.65)
	}
}

func BenchmarkConstitutionIntegrationMocked(b *testing.B) {
	mockGate := &mockActionGate{verdict: constitution.VerdictApproved, sig: "SIG-OK"}
	specialist := NewConstitutionSpecialist(mockGate)
	ctx := context.Background()
	res := &ReasoningResult{
		EnvelopeID: "res-bench",
		PrimaryHypothesis: ReasoningHypothesis{
			ID:                   "p1",
			CalibratedConfidence: 0.9,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = specialist.EvaluateResult(ctx, res)
	}
}

func BenchmarkReasonEnvelope(b *testing.B) {
	srv := NewService(DefaultConfig(), nil, nil)
	ctx := context.Background()
	env := communication.Envelope{ID: "bench-env-full", Source: "bench", RawConfidence: 0.9}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = srv.ReasonEnvelope(ctx, env, StrategySpec{})
	}
}
