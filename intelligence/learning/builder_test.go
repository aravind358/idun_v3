package learning

import (
	"testing"
	"time"
)

func TestLearningRequestBuilder(t *testing.T) {
	req, err := NewLearningRequestBuilder().
		WithRequestID("req-build-1").
		WithDomainSchemaID("idun.reasoning.strategy.v1").
		WithTimeWindow(time.Now().Add(-1*time.Hour), time.Now()).
		WithPolicyFingerprint("fp-policy-sha256").
		Build()
	if err != nil {
		t.Fatalf("builder failed: %v", err)
	}
	if req.RequestID != "req-build-1" {
		t.Errorf("expected request id req-build-1, got %q", req.RequestID)
	}

	_, err = NewLearningRequestBuilder().
		WithRequestID("req-build-2").
		// Missing domain schema and fingerprint
		Build()
	if err == nil {
		t.Error("expected build validation error when required fields missing")
	}
}

func TestLearningTraceBuilder(t *testing.T) {
	trace, err := NewLearningTraceBuilder("trace-b-1", "req-b-1", "schema-1", "fp-1").
		WithAggregation(AggregationSummary{
			SummaryID:          "sum-b-1",
			TimeWindowStart:    time.Now().Add(-2 * time.Hour),
			TimeWindowEnd:      time.Now(),
			SourceArtifactHash: "sha-src-b",
		}).
		WithCandidateCount(5).
		WithStatus(StatusPublished).
		WithTerminationReason(ReasonSuccess).
		Build()
	if err != nil {
		t.Fatalf("trace builder failed: %v", err)
	}
	if trace.CandidateCount != 5 {
		t.Errorf("expected candidate count 5, got %d", trace.CandidateCount)
	}
}

func TestCandidateSnapshotBuilder(t *testing.T) {
	cs, err := NewCandidateSnapshotBuilder("snap-b-1", "1.0.1", "schema-b-1", ReplayMetadata{
		LearningFingerprint: "fp-learn",
		PolicyFingerprint:   "fp-pol",
		SourceArtifactHash:  "sha-src",
	}).
		WithLifecycle(LifecycleValidated).
		WithPayload([]byte("test payload")).
		Build()
	if err != nil {
		t.Fatalf("candidate builder failed: %v", err)
	}
	if string(cs.Payload) != "test payload" {
		t.Errorf("expected 'test payload', got %q", string(cs.Payload))
	}
}

func TestExperimentProfileBuilder(t *testing.T) {
	ep, err := NewExperimentProfileBuilder("exp-b-1", "schema-exp", "snap-target").
		WithRatios(0.20, 0.05).
		WithMaxDuration(12 * time.Hour).
		WithReplaySeed(999).
		Build()
	if err != nil {
		t.Fatalf("experiment builder failed: %v", err)
	}
	if ep.ShadowRatio != 0.20 || ep.ReplaySeed != 999 {
		t.Errorf("unexpected parameters: ratio=%f seed=%d", ep.ShadowRatio, ep.ReplaySeed)
	}
}
