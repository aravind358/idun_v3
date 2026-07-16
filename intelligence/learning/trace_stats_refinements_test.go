package learning

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"
)

func TestLearnerPerformanceSummaryValidationAndBuilder(t *testing.T) {
	summary, err := NewLearnerPerformanceSummaryBuilder("learner-reasoning-v1", "1.0.0", "fp-123").
		WithCounters(10, 8, 2).
		WithMetrics(0.85, 1200, 0.80).
		Build()
	if err != nil {
		t.Fatalf("expected valid build, got: %v", err)
	}
	if summary.LearnerID != "learner-reasoning-v1" || summary.Executions != 10 || summary.SuccessRatio != 0.80 {
		t.Errorf("unexpected summary fields: %+v", summary)
	}

	invalidMissingID := &LearnerPerformanceSummary{
		LearnerVersion: "1.0.0",
		SuccessRatio:   0.5,
	}
	if err := invalidMissingID.Validate(); err == nil {
		t.Error("expected validation error for missing learner_id")
	}

	invalidNaN := &LearnerPerformanceSummary{
		LearnerID:              "learner-test",
		AverageValidationScore: math.NaN(),
		SuccessRatio:           0.5,
	}
	if err := invalidNaN.Validate(); err == nil {
		t.Error("expected validation error for NaN score")
	}

	invalidRatio := &LearnerPerformanceSummary{
		LearnerID:    "learner-test",
		SuccessRatio: 1.5,
	}
	if err := invalidRatio.Validate(); err == nil {
		t.Error("expected validation error for out-of-bounds success_ratio")
	}
}

func TestCandidateRejectionSummaryValidationAndBuilder(t *testing.T) {
	summary, err := NewCandidateRejectionSummaryBuilder().
		WithCounts(15, 5, 3, 2, 1, 2, 1, 1).
		Build()
	if err != nil {
		t.Fatalf("expected valid rejection summary build, got: %v", err)
	}
	if summary.TotalRejected != 15 || summary.StatisticalValidationFailures != 5 {
		t.Errorf("unexpected rejection summary fields: %+v", summary)
	}

	invalidSum := &CandidateRejectionSummary{
		TotalRejected:                 2,
		StatisticalValidationFailures: 5,
	}
	if err := invalidSum.Validate(); err == nil {
		t.Error("expected validation error for sum exceeding total_rejected")
	}
}

func TestUpdateLearnerPerformanceSummaryHelper(t *testing.T) {
	var lps *LearnerPerformanceSummary

	u1 := LearnerUsage{
		LearnerID:          "learner-opt-v1",
		CandidatesProduced: 5,
		CandidatesAccepted: 4,
		ExecutionTime:      1000 * time.Microsecond,
		ContributionScore:  0.8,
	}
	lps = UpdateLearnerPerformanceSummary(lps, u1, "1.0.0", "fp-1")
	if lps.Executions != 1 || lps.AcceptedCandidates != 4 || lps.RejectedCandidates != 1 || lps.SuccessRatio != 0.8 {
		t.Errorf("unexpected update state after 1st cycle: %+v", lps)
	}

	u2 := LearnerUsage{
		LearnerID:          "learner-opt-v1",
		CandidatesProduced: 5,
		CandidatesAccepted: 5,
		ExecutionTime:      2000 * time.Microsecond,
		ContributionScore:  1.0,
	}
	lps = UpdateLearnerPerformanceSummary(lps, u2, "1.0.0", "fp-1")
	if lps.Executions != 2 || lps.AcceptedCandidates != 9 || lps.RejectedCandidates != 1 || lps.AverageExecutionTimeUs != 1500 {
		t.Errorf("unexpected update state after 2nd cycle: %+v", lps)
	}
	if math.Abs(lps.AverageValidationScore-0.9) > 1e-6 {
		t.Errorf("unexpected average validation score: %f", lps.AverageValidationScore)
	}
}

func TestComputeCandidateRejectionSummaryHelper(t *testing.T) {
	failedChecks := []string{
		"STAT_SAMPLE_FLOOR",
		"SCHEMA_PAYLOAD_CHECK",
		"STRUCTURAL_CHECK",
		"REPLAY_LINEAGE_CHECK",
		"CAPABILITY_CHECK",
		"CONST_SAFETY_BOUNDS",
		"DUPLICATE_CHECK",
		"UNKNOWN_CUSTOM_CHECK",
	}
	summary := ComputeCandidateRejectionSummary(failedChecks)
	if summary.TotalRejected != 8 || summary.StatisticalValidationFailures != 1 ||
		summary.StructuralValidationFailures != 2 || summary.ReplayValidationFailures != 1 ||
		summary.CapabilityValidationFailures != 1 || summary.ConstitutionalValidationFailures != 1 ||
		summary.DuplicateCandidateFailures != 1 || summary.OtherFailures != 1 {
		t.Errorf("unexpected categorization counts: %+v", summary)
	}
	if err := summary.Validate(); err != nil {
		t.Errorf("unexpected validation failure on computed summary: %v", err)
	}
}

func TestServiceLearnerPerformanceConcurrentSafety(t *testing.T) {
	s, err := NewService()
	if err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	learnerIDs := []string{
		"learner-reasoning-heuristics-v1",
		"learner-planning-specialist-v1",
		"learner-decision-weights-v1",
	}
	for _, lid := range learnerIDs {
		s.mu.Lock()
		s.learnerPerformance[lid] = &LearnerPerformanceSummary{
			LearnerID:      lid,
			LearnerVersion: "1.0.0",
			Executions:     1,
		}
		s.mu.Unlock()
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	const numGoroutines = 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			lid := learnerIDs[idx%len(learnerIDs)]
			_, _ = s.GetLearnerPerformanceSummary(ctx, lid)
			_ = s.ListLearnerPerformanceSummaries(ctx)
		}(i)
	}

	wg.Wait()
}
