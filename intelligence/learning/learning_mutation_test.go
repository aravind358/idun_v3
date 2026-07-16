package learning

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// MUTATION TESTING SUITE — idun/intelligence/learning
// ============================================================================
// Each test verifies a specific behavioral invariant. A code mutation that
// inverts a comparison, removes a guard, or changes a threshold MUST cause
// at least one test in this file to fail — establishing mutation coverage.
// ============================================================================

// MUTATION 1: invert sampleFloor comparison (>= → <)
// Confirms: validation rejects candidates when sample floor is not met.
func TestMutation_SampleFloorGate(t *testing.T) {
	vp := NewDefaultValidationPipeline(nil)
	ctx := context.Background()
	profile := DefaultLearningPolicyProfile()
	// Summary below minimum sample size
	summary := &AggregationSummary{
		TotalArtifactsIngested: 1, // far below MinimumSampleSize
		SourceArtifactHash:     "hash-mut-1",
	}
	cand := &CandidateSnapshot{
		SnapshotID: "cand-mut-1",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  LifecycleDraft,
		Payload:    []byte(`{"weights":{"a":1}}`),
		Lineage: ReplayMetadata{
			PolicyFingerprint:  profile.PolicyFingerprint,
			SourceArtifactHash: "hash-mut-1",
		},
	}
	results, _, err := vp.ValidateCandidate(ctx, cand, summary, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sampleCheck *ValidationResult
	for i := range results {
		if results[i].CheckID == "STAT_SAMPLE_FLOOR" {
			sampleCheck = &results[i]
		}
	}
	if sampleCheck == nil {
		t.Fatal("STAT_SAMPLE_FLOOR check missing from results")
	}
	if sampleCheck.Passed {
		t.Errorf("expected STAT_SAMPLE_FLOOR to FAIL with %d < %d samples", summary.TotalArtifactsIngested, profile.MinimumSampleSize)
	}
}

// MUTATION 2: invert replay hash check (== → !=)
// Confirms: replay lineage mismatch is detected.
func TestMutation_ReplayLineageMismatchDetected(t *testing.T) {
	vp := NewDefaultValidationPipeline(nil)
	ctx := context.Background()
	profile := DefaultLearningPolicyProfile()
	summary := &AggregationSummary{
		TotalArtifactsIngested: int(profile.MinimumSampleSize) + 10,
		SourceArtifactHash:     "correct-hash",
	}
	cand := &CandidateSnapshot{
		SnapshotID: "cand-mut-2",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  LifecycleDraft,
		Payload:    []byte(`{"weights":{"a":1}}`),
		Lineage: ReplayMetadata{
			PolicyFingerprint:  profile.PolicyFingerprint,
			SourceArtifactHash: "WRONG-hash", // intentionally mismatched
		},
	}
	results, _, err := vp.ValidateCandidate(ctx, cand, summary, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var replayCheck *ValidationResult
	for i := range results {
		if results[i].CheckID == "REPLAY_LINEAGE_CHECK" {
			replayCheck = &results[i]
		}
	}
	if replayCheck == nil {
		t.Fatal("REPLAY_LINEAGE_CHECK check missing from results")
	}
	if replayCheck.Passed {
		t.Error("expected REPLAY_LINEAGE_CHECK to FAIL on hash mismatch")
	}
}

// MUTATION 3: invert payload-size check (<= → >)
// Confirms: oversized payloads are rejected by constitutional bounds.
func TestMutation_PayloadSizeBoundEnforced(t *testing.T) {
	vp := NewDefaultValidationPipeline(nil)
	ctx := context.Background()
	profile := DefaultLearningPolicyProfile()
	// Build a payload exceeding MaxPayloadBytes
	oversized := make([]byte, MaxPayloadBytes+1)
	for i := range oversized {
		oversized[i] = 'x'
	}
	summary := &AggregationSummary{
		TotalArtifactsIngested: int(profile.MinimumSampleSize) + 10,
		SourceArtifactHash:     "hash-mut-3",
	}
	cand := &CandidateSnapshot{
		SnapshotID: "cand-mut-3",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  LifecycleDraft,
		Payload:    oversized,
		Lineage: ReplayMetadata{
			PolicyFingerprint:  profile.PolicyFingerprint,
			SourceArtifactHash: "hash-mut-3",
		},
	}
	_, _, err := vp.ValidateCandidate(ctx, cand, summary, profile)
	// Payload validation fires at schema level before pipeline
	if err == nil {
		// Check that CONST_SAFETY_BOUNDS failed
		results, _, _ := vp.ValidateCandidate(ctx, cand, summary, profile)
		var constCheck *ValidationResult
		for i := range results {
			if results[i].CheckID == "CONST_SAFETY_BOUNDS" {
				constCheck = &results[i]
			}
		}
		if constCheck != nil && constCheck.Passed {
			t.Error("expected CONST_SAFETY_BOUNDS to FAIL for oversized payload")
		}
	}
	// Either error or failed check is acceptable — both prove the gate fires.
}

// MUTATION 4: modify SuccessRatio formula (AcceptedCandidates / total → RejectedCandidates / total)
// Confirms: SuccessRatio reflects accepted, not rejected.
func TestMutation_SuccessRatioCorrectness(t *testing.T) {
	usage := LearnerUsage{
		LearnerID:          "mut-learner",
		CandidatesProduced: 10,
		CandidatesAccepted: 7,
		ContributionScore:  0.8,
		ExecutionTime:      5 * time.Millisecond,
	}
	summary := UpdateLearnerPerformanceSummary(nil, usage, "1.0.0", "fp")
	expected := float64(7) / float64(10)
	if summary.SuccessRatio != expected {
		t.Errorf("expected SuccessRatio=%.2f, got %.2f", expected, summary.SuccessRatio)
	}
	if summary.AcceptedCandidates != 7 {
		t.Errorf("expected AcceptedCandidates=7, got %d", summary.AcceptedCandidates)
	}
	if summary.RejectedCandidates != 3 {
		t.Errorf("expected RejectedCandidates=3, got %d", summary.RejectedCandidates)
	}
}

// MUTATION 5: disable capability guard (SupportsOfflineLearning check removed)
// Confirms: service abstains immediately when offline learning capability is false.
func TestMutation_CapabilityGuardEnforced(t *testing.T) {
	caps := DefaultLearningCapabilities()
	caps.SupportsOfflineLearning = false
	snap := &LearningStrategySnapshot{
		SnapshotID:        "snap-mut-cap",
		SchemaVersion:     SchemaVersion,
		ActiveProfile:     DefaultLearningPolicyProfile(),
		Capabilities:      caps,
		AggregationPolicy: DefaultAggregationPolicy(),
	}
	sp := &staticStrategyProvider{snap: snap}
	s, err := NewService(
		WithStrategyProvider(sp),
		WithLearnerRegistry(NewDefaultLearnerRegistry()),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	_ = s.Start()
	defer s.Close()

	req := &LearningRequest{
		RequestID:         "req-cap-mut",
		DomainSchemaID:    "idun.reasoning.trace.v1",
		PolicyFingerprint: DefaultLearningPolicyProfile().PolicyFingerprint,
		TimeWindowStart:   time.Now().Add(-1 * time.Hour),
		TimeWindowEnd:     time.Now().Add(1 * time.Hour),
	}
	res, err := s.RunCycle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != StatusAbstained {
		t.Errorf("expected ABSTAINED status when SupportsOfflineLearning=false, got %v", res.Status)
	}
	if res.TerminationReason != ReasonCapabilityUnavailable {
		t.Errorf("expected CAPABILITY_UNAVAILABLE reason, got %v", res.TerminationReason)
	}
}

// MUTATION 6: remove nil-guard in UpdateLearnerPerformanceSummary
// Confirms: function handles nil existing without panic.
func TestMutation_UpdateLearnerNilGuard(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UpdateLearnerPerformanceSummary panicked on nil existing: %v", r)
		}
	}()
	usage := LearnerUsage{
		LearnerID:          "nil-guard-learner",
		CandidatesProduced: 3,
		CandidatesAccepted: 2,
		ContributionScore:  0.7,
		ExecutionTime:      1 * time.Millisecond,
	}
	result := UpdateLearnerPerformanceSummary(nil, usage, "1.0.0", "fp-ng")
	if result == nil {
		t.Error("expected non-nil result from nil existing")
	}
	if result.LearnerID != "nil-guard-learner" {
		t.Errorf("expected LearnerID=nil-guard-learner, got %q", result.LearnerID)
	}
}

// MUTATION 7: invert draft lifecycle guard in CandidateSnapshot.Validate()
// Confirms: invalid lifecycle is rejected.
func TestMutation_InvalidLifecycleRejected(t *testing.T) {
	cand := &CandidateSnapshot{
		SnapshotID: "snap-mut-lc",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  CandidateLifecycle("INVALID_LC"),
		Payload:    []byte(`{"weights":{"a":1}}`),
	}
	if err := cand.Validate(); err == nil {
		t.Error("expected validation error for invalid lifecycle, got nil")
	}
}

// MUTATION 8: alter rejection summary routing (STAT_SAMPLE_FLOOR maps to StructuralValidationFailures)
// Confirms: rejection category counters are correct.
func TestMutation_RejectionSummaryCategorization(t *testing.T) {
	summary := ComputeCandidateRejectionSummary([]string{
		"STAT_SAMPLE_FLOOR",
		"SCHEMA_PAYLOAD_CHECK",
		"REPLAY_LINEAGE_CHECK",
		"CAPABILITY_CHECK",
		"CONST_SAFETY_BOUNDS",
		"DUPLICATE_CHECK",
		"UNKNOWN_CHECK",
	})
	if summary.StatisticalValidationFailures != 1 {
		t.Errorf("expected StatisticalValidationFailures=1, got %d", summary.StatisticalValidationFailures)
	}
	if summary.StructuralValidationFailures != 1 {
		t.Errorf("expected StructuralValidationFailures=1, got %d", summary.StructuralValidationFailures)
	}
	if summary.ReplayValidationFailures != 1 {
		t.Errorf("expected ReplayValidationFailures=1, got %d", summary.ReplayValidationFailures)
	}
	if summary.CapabilityValidationFailures != 1 {
		t.Errorf("expected CapabilityValidationFailures=1, got %d", summary.CapabilityValidationFailures)
	}
	if summary.ConstitutionalValidationFailures != 1 {
		t.Errorf("expected ConstitutionalValidationFailures=1, got %d", summary.ConstitutionalValidationFailures)
	}
	if summary.DuplicateCandidateFailures != 1 {
		t.Errorf("expected DuplicateCandidateFailures=1, got %d", summary.DuplicateCandidateFailures)
	}
	if summary.OtherFailures != 1 {
		t.Errorf("expected OtherFailures=1, got %d", summary.OtherFailures)
	}
	if summary.TotalRejected != 7 {
		t.Errorf("expected TotalRejected=7, got %d", summary.TotalRejected)
	}
}

// MUTATION 9: alter running average weight computation (prevWeight / currWeight)
// Confirms: AverageValidationScore converges correctly over multiple cycles.
func TestMutation_RunningAverageCorrectness(t *testing.T) {
	var existing *LearnerPerformanceSummary
	scores := []float64{0.8, 0.9, 0.7}
	for _, score := range scores {
		usage := LearnerUsage{
			LearnerID:          "avg-learner",
			CandidatesProduced: 1,
			CandidatesAccepted: 1,
			ContributionScore:  score,
			ExecutionTime:      1 * time.Millisecond,
		}
		existing = UpdateLearnerPerformanceSummary(existing, usage, "1.0.0", "fp")
	}
	expectedAvg := (0.8 + 0.9 + 0.7) / 3.0
	if abs64(existing.AverageValidationScore-expectedAvg) > 1e-9 {
		t.Errorf("expected AverageValidationScore=%.4f, got %.4f", expectedAvg, existing.AverageValidationScore)
	}
	if existing.Executions != 3 {
		t.Errorf("expected Executions=3, got %d", existing.Executions)
	}
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// MUTATION 10: remove CandidateSnapshot.Validate() call in SnapshotRegistry.Publish()
// Confirms: invalid snapshots cannot be published to the registry.
func TestMutation_SnapshotRegistryFirewall(t *testing.T) {
	reg := NewDefaultSnapshotRegistry()
	ctx := context.Background()

	// Missing SnapshotID — structurally invalid
	invalid := &CandidateSnapshot{
		SnapshotID: "",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  LifecycleValidated,
		Payload:    []byte(`{"weights":{"a":1}}`),
	}
	err := reg.Publish(ctx, invalid)
	if err == nil {
		t.Error("expected publish to reject invalid CandidateSnapshot with empty SnapshotID")
	}
}

// staticStrategyProvider is a test helper for deterministic strategy injection.
type staticStrategyProvider struct {
	snap *LearningStrategySnapshot
}

func (p *staticStrategyProvider) ActiveSnapshot() *LearningStrategySnapshot {
	return p.snap
}
