package learning

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type mockLearner struct {
	id       string
	consumes []string
	produces []string
}

func (m *mockLearner) LearnerID() string { return m.id }
func (m *mockLearner) LearnerVersion() string { return "1.0.0" }
func (m *mockLearner) LearnerFingerprint() string { return "fp-mock" }
func (m *mockLearner) Consumes() []string { return m.consumes }
func (m *mockLearner) Produces() []string { return m.produces }
func (m *mockLearner) Generate(ctx context.Context, summary *AggregationSummary) ([]*CandidateSnapshot, error) {
	return nil, errors.New("phase 1 mock generator placeholder")
}

func TestLearnerRegistry(t *testing.T) {
	reg := NewLearnerRegistry()

	l1 := &mockLearner{
		id:       "learner-reasoning",
		consumes: []string{"idun.reasoning.trace.v1"},
		produces: []string{"idun.reasoning.strategy.v1"},
	}
	l2 := &mockLearner{
		id:       "learner-planning",
		consumes: []string{"idun.planning.trace.v1"},
		produces: []string{"idun.planning.htn.v1"},
	}

	if err := reg.Register(l1); err != nil {
		t.Fatalf("failed to register l1: %v", err)
	}
	if err := reg.Register(l2); err != nil {
		t.Fatalf("failed to register l2: %v", err)
	}

	// Duplicate registration error
	if err := reg.Register(l1); !errors.Is(err, ErrLearnerAlreadyRegistered) {
		t.Errorf("expected ErrLearnerAlreadyRegistered, got %v", err)
	}

	// Lookup by consumes
	consumers := reg.LookupByConsumes("idun.reasoning.trace.v1")
	if len(consumers) != 1 || consumers[0].LearnerID() != "learner-reasoning" {
		t.Errorf("unexpected consumers: %v", consumers)
	}

	// Lookup by produces
	producers := reg.LookupByProduces("idun.planning.htn.v1")
	if len(producers) != 1 || producers[0].LearnerID() != "learner-planning" {
		t.Errorf("unexpected producers: %v", producers)
	}

	// List
	list := reg.ListLearners()
	if len(list) != 2 {
		t.Errorf("expected 2 learners, got %d", len(list))
	}
}

func TestDefaultSnapshotRegistry(t *testing.T) {
	ctx := context.Background()
	reg := NewDefaultSnapshotRegistry()

	cand1 := &CandidateSnapshot{
		SnapshotID: "snap-1.0.0",
		SemVer:     "1.0.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  LifecycleValidated,
		Lineage: ReplayMetadata{
			LearningFingerprint: "fp-learn",
			PolicyFingerprint:   "fp-pol",
			SourceArtifactHash:  "sha-src",
		},
		Payload: []byte(`{"version":1}`),
	}

	if err := reg.Publish(ctx, cand1); err != nil {
		t.Fatalf("publish cand1 failed: %v", err)
	}

	active, err := reg.GetActive(ctx, "idun.reasoning.strategy.v1")
	if err != nil || active.SnapshotID != "snap-1.0.0" {
		t.Fatalf("unexpected active snapshot: %v, %v", active, err)
	}

	// Publish second snapshot
	cand2 := &CandidateSnapshot{
		SnapshotID: "snap-1.1.0",
		SemVer:     "1.1.0",
		SchemaID:   "idun.reasoning.strategy.v1",
		Lifecycle:  LifecycleValidated,
		Lineage: ReplayMetadata{
			LearningFingerprint: "fp-learn",
			PolicyFingerprint:   "fp-pol",
			SourceArtifactHash:  "sha-src",
		},
		Payload: []byte(`{"version":2}`),
	}
	if err := reg.Publish(ctx, cand2); err != nil {
		t.Fatalf("publish cand2 failed: %v", err)
	}

	if current, _ := reg.GetActive(ctx, "idun.reasoning.strategy.v1"); current.SnapshotID != "snap-1.1.0" {
		t.Errorf("expected snap-1.1.0 active, got %q", current.SnapshotID)
	}

	// Rollback to 1.0.0
	if err := reg.Rollback(ctx, "idun.reasoning.strategy.v1", "1.0.0"); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	if current, _ := reg.GetActive(ctx, "idun.reasoning.strategy.v1"); current.SnapshotID != "snap-1.0.0" {
		t.Errorf("expected rolled back snap-1.0.0, got %q", current.SnapshotID)
	}
}

func TestSnapshotRegistryConcurrentSafety(t *testing.T) {
	ctx := context.Background()
	reg := NewDefaultSnapshotRegistry()

	var wg sync.WaitGroup
	workers := 10
	iterations := 50

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				c := &CandidateSnapshot{
					SnapshotID: "snap-conc",
					SemVer:     "1.0.0",
					SchemaID:   "idun.conc.v1",
					Lifecycle:  LifecycleValidated,
					Lineage: ReplayMetadata{
						LearningFingerprint: "fp-l",
						PolicyFingerprint:   "fp-p",
						SourceArtifactHash:  "sha-s",
					},
					Payload: []byte("{}"),
				}
				_ = reg.Publish(ctx, c)
				_, _ = reg.GetActive(ctx, "idun.conc.v1")
			}
		}(i)
	}
	wg.Wait()
}

func TestNewDefaultLearnerRegistry(t *testing.T) {
	reg := NewDefaultLearnerRegistry()
	learners := reg.ListLearners()
	if len(learners) != 9 {
		t.Fatalf("expected 9 default learners registered, got %d", len(learners))
	}
	expectedIDs := []string{
		"learner-reasoning-heuristics-v1",
		"learner-planning-specialist-v1",
		"learner-decision-weights-v1",
		"learner-threshold-opt-v1",
		"learner-weight-opt-v1",
		"learner-calibration-opt-v1",
		"learner-confidence-opt-v1",
		"learner-preference-opt-v1",
		"learner-cross-domain-v1",
	}
	for _, id := range expectedIDs {
		if _, ok := reg.Get(id); !ok {
			t.Errorf("expected default learner %q to be registered", id)
		}
	}
}
