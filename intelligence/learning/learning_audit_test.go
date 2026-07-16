package learning

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"idun/core/memory"
)

func TestAuditPrivacyNoRawPayloadInTelemetry(t *testing.T) {
	telemetryTypes := []interface{}{
		LearningTrace{},
		LearnerPerformanceSummary{},
		CandidateRejectionSummary{},
		TraceStatisticalSummary{},
		LearningCampaignSummary{},
		LearnerUsage{},
	}

	for _, proto := range telemetryTypes {
		typ := reflect.TypeOf(proto)
		assertNoRawPayloadFields(t, typ)
	}
}

func assertNoRawPayloadFields(t *testing.T, typ reflect.Type) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// Check for forbidden raw payload/data types
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Uint8 {
			t.Errorf("Privacy violation: telemetry struct %s contains raw byte slice field %s", typ.Name(), field.Name)
		}
		if field.Type.Kind() == reflect.Map && field.Type.Elem().Kind() == reflect.Interface {
			t.Errorf("Privacy violation: telemetry struct %s contains unbounded interface map field %s", typ.Name(), field.Name)
		}
		// Recursively check struct fields
		if field.Type.Kind() == reflect.Struct && field.Type.PkgPath() == "idun/intelligence/learning" {
			assertNoRawPayloadFields(t, field.Type)
		}
	}
}

func TestAuditReplayDeterminismMetadata(t *testing.T) {
	store := newMockMemory()
	agg := NewDefaultAggregator(store)
	s, err := NewService(
		WithAggregator(agg),
		WithLearnerRegistry(NewDefaultLearnerRegistry()),
	)
	if err != nil {
		t.Fatalf("failed to init service: %v", err)
	}
	_ = s.Start()
	defer s.Close()

	ctx := context.Background()
	_ = store.CreateRecord(memory.Record{
		ID:        "rec-audit-1",
		Type:      "idun.reasoning.trace.v1",
		Payload:   []byte(`{"category":"COGNITIVE_PERFORMANCE","score":0.90}`),
		CreatedAt: time.Now().Add(-30 * time.Minute),
	})

	req := &LearningRequest{
		RequestID:         "req-audit-1",
		DomainSchemaID:    "idun.reasoning.trace.v1",
		PolicyFingerprint: DefaultLearningPolicyProfile().PolicyFingerprint,
		TimeWindowStart:   time.Now().Add(-1 * time.Hour),
		TimeWindowEnd:     time.Now().Add(1 * time.Hour),
	}

	res, err := s.RunCycle(ctx, req)
	if err != nil {
		t.Fatalf("RunCycle failed: %v", err)
	}

	for _, cand := range res.Candidates {
		if cand.Lineage.PolicyFingerprint == "" {
			t.Errorf("candidate %s missing PolicyFingerprint in lineage", cand.SnapshotID)
		}
		if cand.Lineage.SourceArtifactHash == "" {
			t.Errorf("candidate %s missing SourceArtifactHash in lineage", cand.SnapshotID)
		}
	}

	for _, trace := range res.Traces {
		if trace.PolicyFingerprint == "" {
			t.Errorf("trace %s missing PolicyFingerprint", trace.TraceID)
		}
	}
}

func TestAuditResponsibilityBoundaryNoLearnerRanking(t *testing.T) {
	ranking := NewDefaultCandidateRankingEngine()
	ctx := context.Background()
	profile := DefaultLearningPolicyProfile()
	summary := &AggregationSummary{}

	c1 := &CandidateSnapshot{
		SnapshotID: "c1",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lineage: ReplayMetadata{
			LearnerFingerprint: "learner-a",
		},
		ValidationRecords: []ValidationResult{
			{CheckID: "SCHEMA_PAYLOAD_CHECK", Passed: true, Score: 0.95},
		},
	}
	c2 := &CandidateSnapshot{
		SnapshotID: "c2",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lineage: ReplayMetadata{
			LearnerFingerprint: "learner-b",
		},
		ValidationRecords: []ValidationResult{
			{CheckID: "SCHEMA_PAYLOAD_CHECK", Passed: true, Score: 0.80},
		},
	}

	ranked, err := ranking.RankCandidates(ctx, []*CandidateSnapshot{c2, c1}, summary, profile)
	if err != nil {
		t.Fatalf("RankCandidates failed: %v", err)
	}
	if len(ranked) != 2 || ranked[0].SnapshotID != "c1" {
		t.Errorf("expected candidate c1 ranked first based purely on Pareto validation score, got %v", ranked)
	}
}

func TestAuditTelemetryImmutabilityAndIsolation(t *testing.T) {
	s, err := NewService(
		WithAggregator(NewDefaultAggregator(newMockMemory())),
		WithLearnerRegistry(NewDefaultLearnerRegistry()),
	)
	if err != nil {
		t.Fatalf("failed to init service: %v", err)
	}
	_ = s.Start()
	defer s.Close()

	ctx := context.Background()
	req := &LearningRequest{
		RequestID:         "req-iso-1",
		DomainSchemaID:    "idun.reasoning.trace.v1",
		PolicyFingerprint: DefaultLearningPolicyProfile().PolicyFingerprint,
		TimeWindowStart:   time.Now().Add(-1 * time.Hour),
		TimeWindowEnd:     time.Now().Add(1 * time.Hour),
	}
	_, _ = s.RunCycle(ctx, req)

	// Fetch summary and mutate it externally
	sum, err := s.GetLearnerPerformanceSummary(ctx, "learner-reasoning-heuristics-v1")
	if err == nil && sum != nil {
		sum.Executions = 999999
		sum.LearnerFingerprint = "COMPROMISED"
	}

	// Verify internal state remains completely unmutated and isolated
	sumAfter, err := s.GetLearnerPerformanceSummary(ctx, "learner-reasoning-heuristics-v1")
	if err == nil && sumAfter != nil {
		if sumAfter.Executions == 999999 || strings.Contains(sumAfter.LearnerFingerprint, "COMPROMISED") {
			t.Errorf("Audit failure: internal learner performance state was mutated via external reference!")
		}
	}
}
