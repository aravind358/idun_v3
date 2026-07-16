package learning

import (
	"strings"
	"testing"
	"time"
)

func TestReplayMetadataValidation(t *testing.T) {
	rm := ReplayMetadata{
		LearningFingerprint: "fp-learn-1",
		PolicyFingerprint:   "fp-policy-1",
		SourceArtifactHash:  "sha256-src-1",
		ReplaySeed:          42,
		ExperimentID:        "exp-1",
		ParentSnapshot:      "parent-1",
	}
	if err := rm.Validate(); err != nil {
		t.Fatalf("expected valid ReplayMetadata, got %v", err)
	}

	rmInvalid := rm
	rmInvalid.LearningFingerprint = ""
	if err := rmInvalid.Validate(); err == nil {
		t.Error("expected error for missing learning_fingerprint")
	}

	rmInvalid = rm
	rmInvalid.ExperimentID = strings.Repeat("a", MaxStringLength+1)
	if err := rmInvalid.Validate(); err == nil {
		t.Error("expected error for string length exceeded")
	}
}

func TestLearningPolicyProfileValidation(t *testing.T) {
	prof := DefaultLearningPolicyProfile()
	if err := prof.Validate(); err != nil {
		t.Fatalf("default profile validation failed: %v", err)
	}

	invalidAuthor := *prof
	invalidAuthor.Author = "Learning"
	if err := invalidAuthor.Validate(); err == nil {
		t.Error("expected error for non-Executive author")
	}

	invalidRate := *prof
	invalidRate.LearningRate = 1.5
	if err := invalidRate.Validate(); err == nil {
		t.Error("expected error for learning_rate out of bounds")
	}

	invalidSample := *prof
	invalidSample.MinimumSampleSize = 0
	if err := invalidSample.Validate(); err == nil {
		t.Error("expected error for zero minimum_sample_size")
	}
}

func TestLearnerUsageValidation(t *testing.T) {
	lu := LearnerUsage{
		LearnerID:          "learner-1",
		DomainSchemaID:     "idun.reasoning.strategy.v1",
		Invoked:            true,
		CandidatesProduced: 10,
		CandidatesAccepted: 8,
		ExecutionTime:      100 * time.Millisecond,
		ContributionScore:  0.8,
	}
	if err := lu.Validate(); err != nil {
		t.Fatalf("expected valid LearnerUsage, got %v", err)
	}

	invalidAccepted := lu
	invalidAccepted.CandidatesAccepted = 12
	if err := invalidAccepted.Validate(); err == nil {
		t.Error("expected error when candidates_accepted > candidates_produced")
	}

	invalidScore := lu
	invalidScore.ContributionScore = -0.1
	if err := invalidScore.Validate(); err == nil {
		t.Error("expected error for negative contribution_score")
	}
}

func TestCandidateSnapshotValidation(t *testing.T) {
	cs := CandidateSnapshot{
		SnapshotID: "snap-1",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  LifecycleDraft,
		Lineage: ReplayMetadata{
			LearningFingerprint: "fp-learn",
			PolicyFingerprint:   "fp-policy",
			SourceArtifactHash:  "sha-src",
		},
		Payload: []byte(`{"weights":[1,2,3]}`),
	}
	if err := cs.Validate(); err != nil {
		t.Fatalf("expected valid CandidateSnapshot, got %v", err)
	}

	invalidLifecycle := cs
	invalidLifecycle.Lifecycle = CandidateLifecycle("ILLEGAL_STATE")
	if err := invalidLifecycle.Validate(); err == nil {
		t.Error("expected error for invalid lifecycle")
	}

	invalidPayload := cs
	invalidPayload.Payload = make([]byte, MaxPayloadBytes+1)
	if err := invalidPayload.Validate(); err == nil {
		t.Error("expected error for oversized payload")
	}
}

func TestLearningTraceAndResultValidation(t *testing.T) {
	trace := &LearningTrace{
		TraceID:           "trace-1",
		RequestID:         "req-1",
		DomainSchemaID:    "idun.reasoning.strategy.v1",
		PolicyFingerprint: "fp-policy",
		Aggregation: AggregationSummary{
			SummaryID:          "sum-1",
			TimeWindowStart:    time.Now().Add(-1 * time.Hour),
			TimeWindowEnd:      time.Now(),
			SourceArtifactHash: "sha256-hash",
		},
		Status:            StatusPublished,
		TerminationReason: ReasonSuccess,
		TotalDuration:     50 * time.Millisecond,
		TraceTimestamp:    time.Now(),
	}
	if err := trace.Validate(); err != nil {
		t.Fatalf("expected valid LearningTrace, got %v", err)
	}

	res := &LearningResult{
		ResultID:          "res-1",
		RequestID:         "req-1",
		Status:            StatusPublished,
		TerminationReason: ReasonSuccess,
		Candidates: []*CandidateSnapshot{
			{
				SnapshotID: "snap-1",
				SemVer:     "1.0.0",
				SchemaID:   "idun.reasoning.strategy.v1",
				Lifecycle:  LifecycleValidated,
				Lineage: ReplayMetadata{
					LearningFingerprint: "fp-learn",
					PolicyFingerprint:   "fp-policy",
					SourceArtifactHash:  "sha-src",
				},
				Payload: []byte("{}"),
			},
		},
		Traces:        []*LearningTrace{trace},
		TotalDuration: 50 * time.Millisecond,
	}
	if err := res.Validate(); err != nil {
		t.Fatalf("expected valid LearningResult, got %v", err)
	}

	resInvalid := *res
	resInvalid.Status = LearningResultStatus("ILLEGAL")
	if err := resInvalid.Validate(); err == nil {
		t.Error("expected error for invalid result status")
	}
}

func TestLearningStrategySnapshotValidation(t *testing.T) {
	snap := &LearningStrategySnapshot{
		SnapshotID:    "snap-test",
		SchemaVersion: SchemaVersion,
		ActiveProfile: DefaultLearningPolicyProfile(),
		Capabilities:  DefaultLearningCapabilities(),
		CreatedAt:     time.Now(),
	}
	if err := snap.Validate(); err != nil {
		t.Fatalf("expected valid strategy snapshot, got %v", err)
	}

	snapInvalidVer := *snap
	snapInvalidVer.SchemaVersion = "1.0.0-OLD"
	if err := snapInvalidVer.Validate(); err == nil {
		t.Error("expected error on schema version mismatch")
	}
}
