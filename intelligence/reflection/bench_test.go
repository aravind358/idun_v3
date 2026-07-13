package reflection

import (
	"context"
	"fmt"
	"testing"

	"idun/intelligence/communication"
)

func BenchmarkReflectEpisode(b *testing.B) {
	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	traces := []communication.Envelope{
		{ID: "e1", Source: "test", Topic: "test"},
		{ID: "e2", Source: "test", Topic: "test"},
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = srv.ReflectEpisode(ctx, fmt.Sprintf("ep-%d", i), traces)
	}
}

func BenchmarkReflectPeriodic(b *testing.B) {
	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	sum := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-bench",
		GeneratedTimestamp: 100,
		TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 100},
		EpisodeCount:       10,
		AverageScores: map[string]float64{
			"Reasoning":     0.85,
			"Communication": 0.60,
		},
		SummaryConfidence: 0.90,
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = srv.ReflectPeriodic(ctx, sum)
	}
}

func BenchmarkAnalyzeTrends(b *testing.B) {
	sum := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-bench-trend",
		GeneratedTimestamp: 100,
		TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 100},
		ImprovementRates: map[string]float64{
			"Reasoning": 0.15,
		},
		SummaryConfidence: 0.90,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AnalyzeTrends(sum)
	}
}

func BenchmarkAnalyzeCrossCognitive(b *testing.B) {
	sum := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-bench-cc",
		GeneratedTimestamp: 100,
		TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 100},
		AverageScores: map[string]float64{
			"Reasoning":     0.88,
			"Communication": 0.55,
		},
		SummaryConfidence: 0.90,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AnalyzeCrossCognitivePeriodic(sum)
	}
}

func BenchmarkEstimateGrowthPotential(b *testing.B) {
	sum := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-bench-gp",
		GeneratedTimestamp: 100,
		TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 100},
		AverageScores: map[string]float64{
			"Reasoning": 0.65,
		},
		SummaryConfidence: 0.90,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EstimateGrowthPotentialPeriodic(sum)
	}
}

func BenchmarkReflectOnReflection(b *testing.B) {
	srv := NewService()
	_ = srv.Start()
	defer srv.Close()

	sum := HistoricalSummary{
		SchemaVersion:      SchemaVersion,
		SummaryID:          "sum-bench-meta",
		GeneratedTimestamp: 100,
		TimeWindow:         TimeWindowSpec{StartTime: 0, EndTime: 100},
		AverageScores: map[string]float64{
			"Reasoning": 0.85,
		},
		SummaryConfidence: 0.90,
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = srv.ReflectOnReflection(ctx, nil, sum)
	}
}

func BenchmarkReflectionReportValidate(b *testing.B) {
	rep := ReflectionReport{
		SchemaVersion: SchemaVersion,
		ReportID:      "refl-bench",
		EpisodeID:     "ep-bench",
		Timestamp:     100,
		Mode:          ModeEpisode,
		SpecialistReports: []SpecialistReport{
			{
				SpecialistID:         "spec-1",
				TargetAbility:        "Reasoning",
				Verdict:              VerdictEvaluated,
				ReflectionConfidence: 0.95,
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rep.Validate()
	}
}

func BenchmarkReflectionReportBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewReflectionReportBuilder(fmt.Sprintf("r-%d", i), ModeEpisode, 100).
			WithEpisodeID("ep-1").
			AddSpecialistReport(SpecialistReport{
				SpecialistID:         "s1",
				TargetAbility:        "Reasoning",
				Verdict:              VerdictEvaluated,
				ReflectionConfidence: 0.90,
			}).
			Build()
	}
}
