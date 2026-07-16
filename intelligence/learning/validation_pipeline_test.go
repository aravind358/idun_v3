package learning

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestArtifactValidator(t *testing.T) {
	av := NewDefaultArtifactValidator()

	// Check supported default schema
	if !av.IsSupportedSchema("idun.reasoning.trace.v1") {
		t.Errorf("expected idun.reasoning.trace.v1 to be supported by default")
	}
	if av.IsSupportedSchema("idun.unknown.schema.v99") {
		t.Errorf("expected unknown schema to be unsupported")
	}

	// Validate valid JSON payload
	validPayload := []byte(`{"trace_id":"trace-1","metrics":{"accuracy":0.98}}`)
	if err := av.ValidatePayload("idun.reasoning.trace.v1", validPayload); err != nil {
		t.Errorf("expected valid payload to pass, got: %v", err)
	}

	// Validate invalid JSON payload
	invalidPayload := []byte(`{not valid json}`)
	if err := av.ValidatePayload("idun.reasoning.trace.v1", invalidPayload); err == nil {
		t.Errorf("expected invalid JSON payload to fail")
	}

	// Validate empty payload
	if err := av.ValidatePayload("idun.reasoning.trace.v1", []byte{}); err == nil {
		t.Errorf("expected empty payload to fail")
	}

	// Validate unsupported schema
	if err := av.ValidatePayload("idun.unknown.schema.v99", validPayload); err == nil {
		t.Errorf("expected unsupported schema validation to fail")
	}

	// Register custom schema & validator
	customErr := errors.New("custom check failed")
	av.RegisterSchema("idun.custom.schema.v1", func(b []byte) error {
		if string(b) == "fail" {
			return customErr
		}
		return nil
	})
	if !av.IsSupportedSchema("idun.custom.schema.v1") {
		t.Errorf("expected newly registered custom schema to be supported")
	}
	if err := av.ValidatePayload("idun.custom.schema.v1", []byte("fail")); !errors.Is(err, customErr) {
		t.Errorf("expected custom validation error, got: %v", err)
	}
	if err := av.ValidatePayload("idun.custom.schema.v1", []byte("pass")); err != nil {
		t.Errorf("expected custom check to pass for 'pass', got: %v", err)
	}
}

func TestValidationPipelineParameterOptimization(t *testing.T) {
	ctx := context.Background()
	pipeline := NewDefaultValidationPipeline(nil)

	summary := &AggregationSummary{
		SummaryID:              "sum-param-1",
		TimeWindowStart:        time.Now().Add(-2 * time.Hour),
		TimeWindowEnd:          time.Now(),
		TotalArtifactsIngested: 25,
		SourceArtifactHash:     "hash-param-exact-123",
		DomainSchemaIDs:        []string{"idun.reasoning.trace.v1"},
	}

	candidate := &CandidateSnapshot{
		SnapshotID: "cand-param-1",
		SemVer:     "1.0.1",
		SchemaID:   "idun.reasoning.trace.v1", // Not a structural strategy schema -> Parameter optimization path
		Lifecycle:  LifecycleDraft,
		Lineage: ReplayMetadata{
			LearningFingerprint: "fp-learn",
			PolicyFingerprint:   "fp-policy",
			SourceArtifactHash:  "hash-param-exact-123",
		},
		Payload: []byte(`{"weights":{"alpha":0.4,"beta":0.6}}`),
	}

	profile := DefaultLearningPolicyProfile()
	profile.MinimumSampleSize = 10

	results, structRes, err := pipeline.ValidateCandidate(ctx, candidate, summary, profile)
	if err != nil {
		t.Fatalf("ValidateCandidate failed: %v", err)
	}

	// Verify bifurcated path: StructuralValidationResult is nil for parameter optimization
	if structRes != nil {
		t.Errorf("expected nil structural validation result for parameter optimization path, got: %v", structRes)
	}

	if len(results) != 4 {
		t.Fatalf("expected 4 validation records, got %d", len(results))
	}

	for i, res := range results {
		if !res.Passed {
			t.Errorf("expected check [%d] (%s) to pass, reason: %s", i, res.CheckID, res.Reason)
		}
		if res.Evidence == nil {
			t.Errorf("expected factual evidence for check [%d] (%s)", i, res.CheckID)
		} else {
			if res.Evidence.SampleCount != 25 {
				t.Errorf("expected evidence sample count 25, got %d", res.Evidence.SampleCount)
			}
		}
	}
}

func TestValidationPipelineStructuralProposal(t *testing.T) {
	ctx := context.Background()
	pipeline := NewDefaultValidationPipeline(nil)

	summary := &AggregationSummary{
		SummaryID:              "sum-struct-1",
		TimeWindowStart:        time.Now().Add(-5 * time.Hour),
		TimeWindowEnd:          time.Now(),
		TotalArtifactsIngested: 50,
		SourceArtifactHash:     "hash-struct-999",
		DomainSchemaIDs:        []string{"idun.reasoning.strategy.v1"},
	}

	candidate := &CandidateSnapshot{
		SnapshotID: "cand-struct-1",
		SemVer:     "2.0.0",
		SchemaID:   "idun.reasoning.strategy.v1", // Structural strategy proposal
		Lifecycle:  LifecycleDraft,
		Lineage: ReplayMetadata{
			LearningFingerprint: "fp-learn",
			PolicyFingerprint:   "fp-policy",
			SourceArtifactHash:  "hash-struct-999",
		},
		Payload: []byte(`{"graph_steps":["Decompose","Solve","Verify"],"max_branches":4}`),
	}

	profile := DefaultLearningPolicyProfile()
	profile.MinimumSampleSize = 10

	results, structRes, err := pipeline.ValidateCandidate(ctx, candidate, summary, profile)
	if err != nil {
		t.Fatalf("ValidateCandidate failed: %v", err)
	}

	// Verify bifurcated path: StructuralValidationResult MUST be populated for structural strategy proposals
	if structRes == nil {
		t.Fatalf("expected non-nil structural validation result for structural proposal path")
	}
	if !structRes.Passed || !structRes.StaticSyntaxPassed || !structRes.ComplexityBounded || !structRes.CycleFree {
		t.Errorf("expected all structural checks passed, got: %+v", structRes)
	}
	if len(results) != 4 {
		t.Errorf("expected 4 validation records, got %d", len(results))
	}
}

func TestValidationPipelineFailures(t *testing.T) {
	ctx := context.Background()
	pipeline := NewDefaultValidationPipeline(nil)

	// 1. Sample Floor Not Met
	summaryFloor := &AggregationSummary{
		SummaryID:              "sum-fail-1",
		TimeWindowStart:        time.Now().Add(-1 * time.Hour),
		TimeWindowEnd:          time.Now(),
		TotalArtifactsIngested: 3, // Below minimum sample size 10
		SourceArtifactHash:     "hash-fail",
		DomainSchemaIDs:        []string{"idun.reasoning.trace.v1"},
	}
	candidateFloor := &CandidateSnapshot{
		SnapshotID: "cand-fail-1",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.trace.v1",
		Lifecycle:  LifecycleDraft,
		Lineage: ReplayMetadata{
			LearningFingerprint: "fp",
			PolicyFingerprint:   "fp",
			SourceArtifactHash:  "hash-fail",
		},
		Payload: []byte(`{"ok":true}`),
	}
	profile := DefaultLearningPolicyProfile()
	profile.MinimumSampleSize = 10

	resultsFloor, _, err := pipeline.ValidateCandidate(ctx, candidateFloor, summaryFloor, profile)
	if err != nil {
		t.Fatalf("ValidateCandidate floor check errored: %v", err)
	}
	var foundFloorCheck bool
	for _, res := range resultsFloor {
		if res.CheckID == "STAT_SAMPLE_FLOOR" {
			foundFloorCheck = true
			if res.Passed {
				t.Errorf("expected STAT_SAMPLE_FLOOR to fail when sample floor unmet")
			}
		}
	}
	if !foundFloorCheck {
		t.Errorf("did not find STAT_SAMPLE_FLOOR check in results")
	}

	// 2. Replay Lineage Hash Mismatch
	candidateMismatch := &CandidateSnapshot{
		SnapshotID: "cand-mismatch-1",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.trace.v1",
		Lifecycle:  LifecycleDraft,
		Lineage: ReplayMetadata{
			LearningFingerprint: "fp",
			PolicyFingerprint:   "fp",
			SourceArtifactHash:  "hash-WRONG-999", // Does not match summary.SourceArtifactHash
		},
		Payload: []byte(`{"ok":true}`),
	}
	summaryMismatch := &AggregationSummary{
		SummaryID:              "sum-mismatch",
		TimeWindowStart:        time.Now().Add(-1 * time.Hour),
		TimeWindowEnd:          time.Now(),
		TotalArtifactsIngested: 15,
		SourceArtifactHash:     "hash-RIGHT-111",
		DomainSchemaIDs:        []string{"idun.reasoning.trace.v1"},
	}
	resultsMismatch, _, err := pipeline.ValidateCandidate(ctx, candidateMismatch, summaryMismatch, profile)
	if err != nil {
		t.Fatalf("ValidateCandidate mismatch errored: %v", err)
	}
	var foundMismatchCheck bool
	for _, res := range resultsMismatch {
		if res.CheckID == "REPLAY_LINEAGE_CHECK" {
			foundMismatchCheck = true
			if res.Passed {
				t.Errorf("expected REPLAY_LINEAGE_CHECK to fail on hash mismatch")
			}
		}
	}
	if !foundMismatchCheck {
		t.Errorf("did not find REPLAY_LINEAGE_CHECK in results")
	}
}
