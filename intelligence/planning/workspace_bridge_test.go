package planning

import (
	"context"
	"testing"
	"time"

	"idun/intelligence/communication"
)

type mockStorer struct {
	storedData []byte
	err        error
}

func (m *mockStorer) Store(ctx context.Context, data []byte) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	m.storedData = data
	return "storage://cas/payload-8888", nil
}

func (m *mockStorer) Retrieve(ctx context.Context, key string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.storedData, nil
}

type mockSubscriber struct{}
func (m *mockSubscriber) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error) {
	return nil, nil
}

type mockPublisher struct {
	publishedEnvs []communication.Envelope
	err           error
}

func (m *mockPublisher) Publish(ctx context.Context, env communication.Envelope) error {
	if m.err != nil {
		return m.err
	}
	m.publishedEnvs = append(m.publishedEnvs, env)
	return nil
}

func TestShouldPublishToWorkspace(t *testing.T) {
	if ShouldPublishToWorkspace(DepthReflexive) {
		t.Error("expected reflexive plans to skip workspace broadcast")
	}
	if !ShouldPublishToWorkspace(DepthTactical) || !ShouldPublishToWorkspace(DepthStrategic) {
		t.Error("expected tactical and strategic plans to publish to workspace")
	}
}

func TestWorkspaceBridge_PublishAndEnvelopes(t *testing.T) {
	plan := &CandidatePlan{
		PlanID:             "plan-pub-1",
		SchemaVersion:      SchemaVersion2_0_0,
		CreatedAt:          time.Now().UTC(),
		StrategySnapshotID: "snap-pub",
		PlanFingerprint:    "hash123",
		Domain:             "General",
		Goal:               "Test publication",
		ConfidenceProfile: ConfidenceProfile{
			GoalConfidence: 0.9, PreconditionConfidence: 0.9, DependencyConfidence: 0.9,
			ResourceConfidence: 0.9, TimingConfidence: 0.9, ConstraintConfidence: 0.9,
			OverallConfidence: 0.9,
		},
		ReplayMetadata: ReplayMetadata{StrategySnapshotID: "snap-pub"},
		PlanStatus:     PlanStatusComplete,
		TraceID:        "trace-pub-1",
	}

	storer := &mockStorer{}
	pub := &mockPublisher{}

	env, err := PublishPlan(context.Background(), plan, storer, pub)
	if err != nil {
		t.Fatalf("failed to publish plan: %v", err)
	}
	if env.Topic != communication.TopicCandidatePlans {
		t.Errorf("expected topic candidate-plans, got %s", env.Topic)
	}
	if env.PayloadRef != "storage://cas/payload-8888" || env.RawConfidence != 0.9 {
		t.Errorf("unexpected envelope fields: %+v", env)
	}
	if len(pub.publishedEnvs) != 1 {
		t.Fatalf("expected 1 envelope published, got %d", len(pub.publishedEnvs))
	}

	// Test Trace publication
	trace := &PlanningTrace{
		TraceID:            "trace-pub-1",
		PlanID:             "plan-pub-1",
		SchemaVersion:      SchemaVersion2_0_0,
		StrategySnapshotID: "snap-pub",
		TerminationReason:  TerminationGoalFound,
		ConfidenceProfile:  plan.ConfidenceProfile,
	}

	traceEnv, err := PublishPlanningTrace(context.Background(), trace, storer, pub)
	if err != nil {
		t.Fatalf("failed to publish trace: %v", err)
	}
	if traceEnv.Topic != communication.TopicReflections {
		t.Errorf("expected trace envelope on reflections topic, got %s", traceEnv.Topic)
	}

	// Test validation firewall interception
	invalidPlan := &CandidatePlan{PlanID: "bad"} // Missing SchemaVersion, Goal, etc.
	_, err = PublishPlan(context.Background(), invalidPlan, storer, pub)
	if err == nil {
		t.Error("expected validation firewall error on invalid plan publication, got nil")
	}
}

func TestMarshalUnmarshalRoundTrips(t *testing.T) {
	plan := &CandidatePlan{PlanID: "p-round", SchemaVersion: SchemaVersion2_0_0, Goal: "Roundtrip", TraceID: "t-1"}
	data, err := MarshalPlan(plan)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	p2, err := UnmarshalPlan(data)
	if err != nil || p2.PlanID != "p-round" {
		t.Fatalf("unmarshal mismatch: %v / %+v", err, p2)
	}

	trace := &PlanningTrace{TraceID: "t-round", PlanID: "p-round", SchemaVersion: SchemaVersion2_0_0, TerminationReason: TerminationGoalFound}
	tData, err := MarshalPlanningTrace(trace)
	if err != nil {
		t.Fatalf("trace marshal failed: %v", err)
	}
	t2, err := UnmarshalPlanningTrace(tData)
	if err != nil || t2.TraceID != "t-round" {
		t.Fatalf("trace unmarshal mismatch: %v / %+v", err, t2)
	}

	res := &PlanningResult{ResultID: "r-round", RequestID: "req-1", Plans: []*CandidatePlan{plan}, Traces: []*PlanningTrace{trace}}
	rData, err := MarshalPlanningResult(res)
	if err != nil {
		t.Fatalf("result marshal failed: %v", err)
	}
	r2, err := UnmarshalPlanningResult(rData)
	if err != nil || r2.ResultID != "r-round" {
		t.Fatalf("result unmarshal mismatch: %v / %+v", err, r2)
	}
}

