package learning

import (
	"math"
	"testing"
	"time"
)

func TestComputeTraceStatisticalSummary(t *testing.T) {
	summary := &AggregationSummary{
		SummaryID:              "sum-test",
		TimeWindowStart:        time.Now().Add(-1 * time.Hour),
		TimeWindowEnd:          time.Now(),
		TotalArtifactsIngested: 100,
		SourceArtifactHash:     "hash-src-123",
		DomainSchemaIDs:        []string{"schema-1", "schema-2"},
	}

	c1 := &CandidateSnapshot{
		SnapshotID: "cand-1",
		SchemaID:   "schema-1",
		Lineage: ReplayMetadata{
			SourceArtifactHash: "hash-src-123",
		},
		ValidationRecords: []ValidationResult{
			{CheckID: "C1", Passed: true, Score: 0.90},
			{CheckID: "C2", Passed: true, Score: 0.80},
		},
	}
	c2 := &CandidateSnapshot{
		SnapshotID: "cand-2",
		SchemaID:   "schema-3",
		Lineage: ReplayMetadata{
			SourceArtifactHash: "hash-src-123",
		},
		ValidationRecords: []ValidationResult{
			{CheckID: "C1", Passed: true, Score: 0.70},
		},
	}

	stats := ComputeTraceStatisticalSummary(summary, []*CandidateSnapshot{c1, c2})

	if stats.TotalArtifactsAnalyzed != 100 {
		t.Errorf("expected 100 artifacts analyzed, got %d", stats.TotalArtifactsAnalyzed)
	}
	if math.Abs(stats.MeanValidationScore-0.80) > 1e-6 {
		t.Errorf("expected mean validation score ~0.80, got %f", stats.MeanValidationScore)
	}
	if stats.MinValidationScore != 0.70 {
		t.Errorf("expected min validation score 0.70, got %f", stats.MinValidationScore)
	}
	if stats.MaxValidationScore != 0.90 {
		t.Errorf("expected max validation score 0.90, got %f", stats.MaxValidationScore)
	}
	if stats.ReplayFidelityRatio != 1.0 {
		t.Errorf("expected replay fidelity ratio 1.0, got %f", stats.ReplayFidelityRatio)
	}
	if math.Abs(stats.DomainCoverageRatio-0.20) > 1e-6 {
		t.Errorf("expected domain coverage ratio ~0.20, got %f", stats.DomainCoverageRatio)
	}

	if err := stats.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestTraceStatisticalSummaryValidationErrors(t *testing.T) {
	stats := &TraceStatisticalSummary{
		TotalArtifactsAnalyzed: -1,
		MeanValidationScore:    0.5,
		EstimatedDriftScore:    0.1,
		ReplayFidelityRatio:    1.0,
	}
	if err := stats.Validate(); err == nil {
		t.Errorf("expected error for negative artifacts analyzed")
	}

	stats = &TraceStatisticalSummary{
		TotalArtifactsAnalyzed: 10,
		MeanValidationScore:    math.NaN(),
		EstimatedDriftScore:    0.1,
		ReplayFidelityRatio:    1.0,
	}
	if err := stats.Validate(); err == nil {
		t.Errorf("expected error for NaN score")
	}
}
