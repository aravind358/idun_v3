package decision

import (
	"context"
	"testing"
)

func BenchmarkEvaluateReflexive(b *testing.B) {
	service := NewService()
	ctx := context.Background()
	cset := CandidateSet{
		EpisodeID: "ep-bench",
		Candidates: []Candidate{
			{
				ID: "cand-1",
				Attributes: map[string]float64{
					"utility": 0.95,
					"safety":  1.0,
				},
			},
			{
				ID: "cand-2",
				Attributes: map[string]float64{
					"utility": 0.85,
					"safety":  0.90,
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.EvaluateReflexive(ctx, cset)
		if err != nil {
			b.Fatalf("BenchmarkEvaluateReflexive error: %v", err)
		}
	}
}

func BenchmarkCalibrateConfidence(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		CalibrateConfidence(0.90, 0.15, 1.0, 0.50)
	}
}
