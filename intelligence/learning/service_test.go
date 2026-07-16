package learning

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/core/memory"
	"idun/intelligence/communication"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

type mockGeneratingLearner struct {
	id       string
	consumes []string
	produces []string
}

func (m *mockGeneratingLearner) LearnerID() string      { return m.id }
func (m *mockGeneratingLearner) LearnerVersion() string   { return "1.0.0" }
func (m *mockGeneratingLearner) LearnerFingerprint() string { return "fp-mock-gen" }
func (m *mockGeneratingLearner) Consumes() []string     { return m.consumes }
func (m *mockGeneratingLearner) Produces() []string { return m.produces }
func (m *mockGeneratingLearner) Generate(ctx context.Context, summary *AggregationSummary) ([]*CandidateSnapshot, error) {
	c := &CandidateSnapshot{
		SnapshotID: "snap-gen-1",
		SemVer:     "1.0.1",
		SchemaID:   m.produces[0],
		Lifecycle:  LifecycleDraft,
		Lineage: ReplayMetadata{
			LearningFingerprint: "fp-gen",
			PolicyFingerprint:   "fp-pol",
			SourceArtifactHash:  summary.SourceArtifactHash,
		},
		Payload: []byte(`{"mock":true}`),
	}
	return []*CandidateSnapshot{c}, nil
}

func TestServiceLifecycle(t *testing.T) {
	svc, err := NewService()
	if err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	if svc.Ability() != executive.AbilityLearning {
		t.Errorf("expected AbilityLearning, got %s", svc.Ability())
	}

	if err := svc.Start(); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	if err := svc.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}
}

func TestServiceCapabilityFailFast(t *testing.T) {
	ctx := context.Background()
	svc, _ := NewService()
	_ = svc.Start()

	// Modify active snapshot to declare NO offline learning support
	snap := svc.strategyProv.ActiveSnapshot()
	modifiedCaps := *snap.Capabilities
	modifiedCaps.SupportsOfflineLearning = false

	modifiedSnap := &LearningStrategySnapshot{
		SnapshotID:    "snap-no-offline",
		SchemaVersion: SchemaVersion,
		ActiveProfile: snap.ActiveProfile,
		Capabilities:  &modifiedCaps,
		CreatedAt:     time.Now(),
	}
	_ = svc.strategyProv.(*DefaultStrategyProvider).SwapSnapshot(modifiedSnap)

	req, _ := NewLearningRequestBuilder().
		WithRequestID("req-failfast").
		WithDomainSchemaID("idun.reasoning.strategy.v1").
		WithPolicyFingerprint(snap.ActiveProfile.PolicyFingerprint).
		Build()

	res, err := svc.RunCycle(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error during failfast check: %v", err)
	}
	if res.Status != StatusAbstained || res.TerminationReason != ReasonCapabilityUnavailable {
		t.Errorf("expected ABSTAINED / CAPABILITY_UNAVAILABLE, got %s / %s", res.Status, res.TerminationReason)
	}
}

func TestServiceRunCycle(t *testing.T) {
	ctx := context.Background()
	svc, _ := NewService()
	_ = svc.Start()

	learner := &mockGeneratingLearner{
		id:       "learner-mock-gen",
		consumes: []string{"idun.reasoning.trace.v1"},
		produces: []string{"idun.reasoning.strategy.v1"},
	}
	if err := svc.RegisterLearner(learner); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	req, _ := NewLearningRequestBuilder().
		WithRequestID("req-cycle-1").
		WithDomainSchemaID("idun.reasoning.strategy.v1").
		WithPolicyFingerprint(svc.config.PolicyProfile.PolicyFingerprint).
		Build()

	res, err := svc.RunCycle(ctx, req)
	if err != nil {
		t.Fatalf("RunCycle failed: %v", err)
	}
	if res.Status != StatusPublished || len(res.Candidates) != 1 {
		t.Errorf("expected PUBLISHED with 1 candidate, got %s with %d candidates", res.Status, len(res.Candidates))
	}
	if len(res.Traces) != 1 || res.Traces[0].CandidateCount != 1 {
		t.Errorf("expected 1 trace with candidate count 1")
	}
}

func TestServiceDriverMethods(t *testing.T) {
	ctx := context.Background()
	svc, _ := NewService()
	_ = svc.Start()

	status, resID, err := svc.ExecuteTask(ctx, "wf-1", "node-1", executive.BudgetStandard, "cas://trace-101")
	if err != nil {
		t.Fatalf("ExecuteTask failed: %v", err)
	}
	if status != executive.StatusConfident || resID == "" {
		t.Errorf("unexpected status/resID: %v, %s", status, resID)
	}

	consolidateID, err := svc.ConsolidateExperience(ctx, "cas://episodic-summary")
	if err != nil {
		t.Fatalf("ConsolidateExperience failed: %v", err)
	}
	if consolidateID == "" {
		t.Error("expected valid consolidation result ID")
	}
}

func TestServiceConcurrentSafety(t *testing.T) {
	ctx := context.Background()
	svc, _ := NewService()
	_ = svc.Start()

	var wg sync.WaitGroup
	workers := 8
	iterations := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				req, _ := NewLearningRequestBuilder().
					WithRequestID(fmt.Sprintf("req-conc-%d-%d", id, j)).
					WithDomainSchemaID("idun.reasoning.strategy.v1").
					WithPolicyFingerprint(svc.config.PolicyProfile.PolicyFingerprint).
					Build()
				_, _ = svc.RunCycle(ctx, req)
			}
		}(i)
	}
	wg.Wait()
}

func TestServiceRunCycleEndToEnd(t *testing.T) {
	ctx := context.Background()
	store := newMockMemory()
	agg := NewDefaultAggregator(store)
	pipe := NewDefaultValidationPipeline(nil)
	exp := NewDefaultExperimentManager(nil)

	svc, err := NewService(
		WithAggregator(agg),
		WithValidationPipeline(pipe),
		WithExperimentManager(exp),
	)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// Populate mock store with 105 reasoning traces to satisfy minimum sample floor (100)
	now := time.Now()
	store.mu.Lock()
	for i := 0; i < 105; i++ {
		store.records = append(store.records, memory.Record{
			ID:        fmt.Sprintf("trace-%d", i),
			Type:      "idun.reasoning.trace.v1",
			Payload:   []byte(`{"step":"mock"}`),
			CreatedAt: now.Add(-time.Duration(i) * time.Minute),
		})
	}
	store.mu.Unlock()

	// Start service before registering learners
	_ = svc.Start()
	if err := svc.RegisterLearner(NewReasoningLearner()); err != nil {
		t.Fatalf("failed to register reasoning learner: %v", err)
	}

	req, _ := NewLearningRequestBuilder().
		WithRequestID("req-e2e-1").
		WithDomainSchemaID("idun.reasoning.strategy.v1").
		WithTimeWindow(now.Add(-2*time.Hour), now.Add(1*time.Hour)).
		WithPolicyFingerprint(svc.config.PolicyProfile.PolicyFingerprint).
		Build()

	res, err := svc.RunCycle(ctx, req)
	if err != nil {
		t.Fatalf("RunCycle failed: %v", err)
	}

	if res.Status != StatusPublished {
		t.Errorf("expected PUBLISHED status, got %s (reason: %s)", res.Status, res.TerminationReason)
	}
	if len(res.Candidates) != 1 {
		t.Fatalf("expected 1 accepted candidate, got %d", len(res.Candidates))
	}

	cand := res.Candidates[0]
	// Verify Phase 2 promotion boundary: candidates strictly transition Draft -> Validated
	if cand.Lifecycle != LifecycleValidated {
		t.Errorf("expected candidate lifecycle to transition to %s, got %s", LifecycleValidated, cand.Lifecycle)
	}
	if len(cand.ValidationRecords) == 0 || cand.StructuralValidation == nil {
		t.Errorf("expected populated validation records and structural check for strategy proposal")
	}

	// Verify snapshot registry published the validated candidate
	pub, err := svc.GetActiveSnapshot(ctx, "idun.reasoning.strategy.v1")
	if err != nil || pub.SnapshotID != cand.SnapshotID {
		t.Errorf("snapshot not published properly: err=%v, pub=%+v", err, pub)
	}

	// Verify trace usage and contribution score
	if len(res.Traces) != 1 || len(res.Traces[0].LearnerUsages) != 1 {
		t.Fatalf("unexpected trace or usage counts")
	}
	u := res.Traces[0].LearnerUsages[0]
	if u.CandidatesProduced != 1 || u.CandidatesAccepted != 1 || u.ContributionScore != 1.0 {
		t.Errorf("unexpected usage metrics: %+v", u)
	}
}

func TestServiceRunCycleValidationFail(t *testing.T) {
	ctx := context.Background()
	store := newMockMemory()
	agg := NewDefaultAggregator(store)
	pipe := NewDefaultValidationPipeline(nil)

	svc, _ := NewService(
		WithAggregator(agg),
		WithValidationPipeline(pipe),
	)
	_ = svc.Start()
	_ = svc.RegisterLearner(NewReasoningLearner())

	// Only add 2 records -> below minimum sample floor (10)
	now := time.Now()
	store.mu.Lock()
	for i := 0; i < 2; i++ {
		store.records = append(store.records, memory.Record{
			ID:        fmt.Sprintf("t-%d", i),
			Type:      "idun.reasoning.trace.v1",
			Payload:   []byte(`{"step":"mock"}`),
			CreatedAt: now,
		})
	}
	store.mu.Unlock()

	req, _ := NewLearningRequestBuilder().
		WithRequestID("req-fail-floor").
		WithDomainSchemaID("idun.reasoning.strategy.v1").
		WithTimeWindow(now.Add(-2*time.Hour), now.Add(1*time.Hour)).
		WithPolicyFingerprint(svc.config.PolicyProfile.PolicyFingerprint).
		Build()

	res, err := svc.RunCycle(ctx, req)
	if err != nil {
		t.Fatalf("RunCycle failed: %v", err)
	}

	if res.Status != StatusValidationFail || res.TerminationReason != ReasonSampleFloorNotMet {
		t.Errorf("expected VALIDATION_FAILED / SAMPLE_FLOOR_NOT_MET, got %s / %s", res.Status, res.TerminationReason)
	}
	if len(res.Candidates) != 0 {
		t.Errorf("expected 0 accepted candidates when validation fails, got %d", len(res.Candidates))
	}

	// Verify snapshot registry remains clean (no candidate published)
	pub, err := svc.GetActiveSnapshot(ctx, "idun.reasoning.strategy.v1")
	if err == nil && pub != nil && pub.SnapshotID != "" {
		t.Errorf("expected no published snapshot when validation failed, got: %+v", pub)
	}
}

type mockGovernanceBridge struct {
	mu            sync.Mutex
	evaluatedRefs []string
}

func (m *mockGovernanceBridge) EvaluateDiagnostics(ctx context.Context, diagnosticsRef string) (*HealthRecommendation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evaluatedRefs = append(m.evaluatedRefs, diagnosticsRef)
	return &HealthRecommendation{
		RecommendationID: "rec-mock-1",
		Action:           ActionContinueRollout,
		Confidence:       1.0,
		Rationale:        "All healthy",
	}, nil
}

type mockRolloutExecutor struct {
	mu         sync.Mutex
	promotions []string
}

func (m *mockRolloutExecutor) PromoteCandidate(ctx context.Context, snapshotID string, targetLifecycle CandidateLifecycle) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.promotions = append(m.promotions, fmt.Sprintf("%s:%s", snapshotID, targetLifecycle))
	return nil
}

func (m *mockRolloutExecutor) GetStatus(ctx context.Context, snapshotID string) (CandidateLifecycle, error) {
	return LifecycleValidated, nil
}

func TestServicePhase3LineageAndSnapshotRegistry(t *testing.T) {
	ctx := context.Background()
	reg := NewLearnerRegistry()
	reg.Register(&mockGeneratingLearner{
		id:       "learner-reasoning-heuristics-v1",
		consumes: []string{"idun.reasoning.trace.v1", "idun.reflection.report.v1"},
		produces: []string{"idun.reasoning.strategy.v1"},
	})

	store := newMockMemory()
	now := time.Now()
	for i := 0; i < 80; i++ {
		_ = store.CreateRecord(memory.Record{
			ID:        fmt.Sprintf("t-%d", i),
			Type:      "idun.reasoning.trace.v1",
			Payload:   []byte(`{"step":"mock"}`),
			CreatedAt: now,
		})
	}
	for i := 0; i < 25; i++ {
		_ = store.CreateRecord(memory.Record{
			ID:        fmt.Sprintf("r-%d", i),
			Type:      "idun.reflection.report.v1",
			Payload:   []byte(`{"report":"mock"}`),
			CreatedAt: now,
		})
	}

	agg := NewDefaultAggregator(store)
	val := NewDefaultValidationPipeline(nil)

	snapReg := NewDefaultSnapshotRegistry()
	// Pre-seed an active parent snapshot in snapReg
	parentSnap := &CandidateSnapshot{
		SnapshotID: "snap-parent-v0",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  LifecycleValidated,
		Lineage: ReplayMetadata{
			LearningFingerprint: "fp-learn-parent",
			PolicyFingerprint:   "fp-pol-parent",
			LearnerFingerprint:  "fp-mock-gen",
			SourceArtifactHash:  "hash-parent",
			ReplaySeed:          12345,
			ParentSnapshot:      "snap-root",
			AncestorSnapshot:    "snap-root",
			GenerationDepth:     1,
		},
		Payload: []byte(`{"parent":true}`),
	}
	if err := snapReg.Publish(ctx, parentSnap); err != nil {
		t.Fatalf("parent publish failed: %v", err)
	}

	svc, err := NewService(
		WithAggregator(agg),
		WithValidationPipeline(val),
		WithSnapshotRegistry(snapReg),
	)
	if err != nil {
		t.Fatalf("service init failed: %v", err)
	}
	svc.learnerReg = reg
	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer svc.Close()

	req, _ := NewLearningRequestBuilder().
		WithRequestID("req-phase3-lineage").
		WithDomainSchemaID("idun.reasoning.strategy.v1").
		WithTimeWindow(now.Add(-2*time.Hour), now.Add(1*time.Hour)).
		WithPolicyFingerprint(svc.config.PolicyProfile.PolicyFingerprint).
		Build()

	res, err := svc.RunCycle(ctx, req)
	if err != nil {
		t.Fatalf("RunCycle failed: %v", err)
	}

	if len(res.Candidates) != 1 {
		t.Fatalf("expected 1 accepted candidate, got %d", len(res.Candidates))
	}

	cand := res.Candidates[0]
	if cand.Provenance == nil {
		t.Fatalf("expected Provenance to be populated on candidate")
	}
	if cand.Provenance.ParentSnapshot != "snap-parent-v0" {
		t.Errorf("expected parent_snapshot snap-parent-v0, got %s", cand.Provenance.ParentSnapshot)
	}
	if cand.Provenance.AncestorSnapshot != "snap-root" {
		t.Errorf("expected ancestor_snapshot snap-root, got %s", cand.Provenance.AncestorSnapshot)
	}
	if cand.Provenance.GenerationDepth != 2 {
		t.Errorf("expected generation_depth 2, got %d", cand.Provenance.GenerationDepth)
	}
	if cand.Lineage.ParentSnapshot != "snap-parent-v0" || cand.Lineage.GenerationDepth != 2 {
		t.Errorf("expected lineage fields synced, got parent=%s depth=%d", cand.Lineage.ParentSnapshot, cand.Lineage.GenerationDepth)
	}
	if cand.Lineage.LearnerFingerprint != "fp-mock-gen" {
		t.Errorf("expected learner fingerprint fp-mock-gen, got %s", cand.Lineage.LearnerFingerprint)
	}

	// Verify trace contains full deterministic replay lineage fields
	if len(res.Traces) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(res.Traces))
	}
	tr := res.Traces[0]
	if tr.Lineage.ParentSnapshot != "snap-parent-v0" || tr.Lineage.GenerationDepth != 2 || tr.Lineage.LearnerFingerprint != "fp-mock-gen" {
		t.Errorf("unexpected trace lineage metadata: %+v", tr.Lineage)
	}
}

func TestServicePhase3GovernanceAndWorkspace(t *testing.T) {
	ctx := context.Background()
	reg := NewLearnerRegistry()
	reg.Register(&mockGeneratingLearner{
		id:       "learner-reasoning-heuristics-v1",
		consumes: []string{"idun.reasoning.trace.v1", "idun.reflection.report.v1"},
		produces: []string{"idun.reasoning.strategy.v1"},
	})

	store := newMockMemory()
	now := time.Now()
	for i := 0; i < 80; i++ {
		_ = store.CreateRecord(memory.Record{
			ID:        fmt.Sprintf("t-%d", i),
			Type:      "idun.reasoning.trace.v1",
			Payload:   []byte(`{"step":"mock"}`),
			CreatedAt: now,
		})
	}
	for i := 0; i < 25; i++ {
		_ = store.CreateRecord(memory.Record{
			ID:        fmt.Sprintf("r-%d", i),
			Type:      "idun.reflection.report.v1",
			Payload:   []byte(`{"report":"mock"}`),
			CreatedAt: now,
		})
	}

	agg := NewDefaultAggregator(store)
	val := NewDefaultValidationPipeline(nil)
	gov := &mockGovernanceBridge{}
	rollout := &mockRolloutExecutor{}
	ws := workspace.NewEngine()
	_ = ws.Start()
	defer ws.Close()

	svc, err := NewService(
		WithAggregator(agg),
		WithValidationPipeline(val),
		WithGovernanceBridge(gov),
		WithRolloutExecutor(rollout),
		WithWorkspace(ws),
	)
	if err != nil {
		t.Fatalf("service init failed: %v", err)
	}
	svc.learnerReg = reg
	_ = svc.Start()
	defer svc.Close()

	req, _ := NewLearningRequestBuilder().
		WithRequestID("req-phase3-gov-ws").
		WithDomainSchemaID("idun.reasoning.strategy.v1").
		WithTimeWindow(now.Add(-2*time.Hour), now.Add(1*time.Hour)).
		WithPolicyFingerprint(svc.config.PolicyProfile.PolicyFingerprint).
		Build()

	res, err := svc.RunCycle(ctx, req)
	if err != nil {
		t.Fatalf("RunCycle failed: %v", err)
	}

	gov.mu.Lock()
	if len(gov.evaluatedRefs) != 1 || gov.evaluatedRefs[0] != res.Traces[0].TraceID {
		t.Errorf("expected governance bridge evaluated diagnostics with trace ID %s, got %v", res.Traces[0].TraceID, gov.evaluatedRefs)
	}
	gov.mu.Unlock()

	rollout.mu.Lock()
	if len(rollout.promotions) != 1 {
		t.Errorf("expected rollout executor promotion called once for validated candidate, got %v", rollout.promotions)
	}
	rollout.mu.Unlock()

	// Verify envelope published to TopicReflections
	envelopes := ws.ListTopicEnvelopes(communication.TopicReflections, 10)
	if len(envelopes) == 0 {
		t.Errorf("expected learning result envelope published to TopicReflections")
	} else {
		found := false
		for _, e := range envelopes {
			if e.PayloadRef == res.ResultID && e.Source == "idun.intelligence.learning" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("did not find envelope matching result ID %s across %d published envelopes", res.ResultID, len(envelopes))
		}
	}
}
