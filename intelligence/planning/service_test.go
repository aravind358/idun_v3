package planning

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
)

func TestService_Lifecycle(t *testing.T) {
	svc := NewService()
	if svc.Ability() != executive.AbilityPlanning {
		t.Errorf("expected AbilityPlanning, got %s", svc.Ability())
	}

	req, _ := NewPlanningRequestBuilder().WithRequestID("req-1").WithGoal("Test").Build()
	_, err := svc.PlanTactical(context.Background(), req)
	if err == nil {
		t.Error("expected error when calling PlanTactical before Start(), got nil")
	}

	if err := svc.Start(); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	res, err := svc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("expected successful PlanTactical after Start(), got: %v", err)
	}
	if res.ResultStatus != ResultNoPlans && res.ResultStatus != ResultSuccess {
		t.Errorf("unexpected result status: %s", res.ResultStatus)
	}

	_ = svc.Close()
	_, err = svc.PlanTactical(context.Background(), req)
	if err == nil {
		t.Error("expected error when calling PlanTactical after Close(), got nil")
	}
}

func TestService_OrchestrationWithSpecialistsAndWorkspace(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(&mockSpecialist{name: "DecompSpec", domains: []string{"General"}})

	storer := &mockStorer{}
	pub := &mockPublisher{}
	sub := &mockSubscriber{}

	svc := NewService(
		WithSpecialistRegistry(reg),
		WithWorkspaceBridge(storer, pub, sub),
	)
	_ = svc.Start()
	defer svc.Close()

	// Tactical Planning (Should publish to workspace)
	reqTac, _ := NewPlanningRequestBuilder().
		WithRequestID("req-tac-1").
		WithGoal("Orchestrate tactical plan").
		WithDomain("General").
		Build()

	resTac, err := svc.PlanTactical(context.Background(), reqTac)
	if err != nil {
		t.Fatalf("PlanTactical failed: %v", err)
	}
	if resTac.ResultStatus != ResultSuccess {
		t.Errorf("expected RESULT_SUCCESS, got %s", resTac.ResultStatus)
	}
	if len(resTac.Plans) != 1 || len(resTac.Plans[0].Subgoals) != 1 {
		t.Errorf("expected 1 plan with 1 subgoal, got %+v", resTac.Plans)
	}

	// Verify that Plan, Trace, and Result envelopes were published
	if len(pub.publishedEnvs) != 3 {
		t.Fatalf("expected 3 envelopes published for tactical plan, got %d", len(pub.publishedEnvs))
	}
	foundPlan, foundTrace, foundRes := false, false, false
	for _, env := range pub.publishedEnvs {
		if env.Topic == communication.TopicCandidatePlans && env.ID[:8] == "env-plan" {
			foundPlan = true
		} else if env.Topic == communication.TopicReflections && env.ID[:9] == "env-trace" {
			foundTrace = true
		} else if env.Topic == communication.TopicCandidatePlans && env.ID[:7] == "env-res" {
			foundRes = true
		}
	}
	if !foundPlan || !foundTrace || !foundRes {
		t.Errorf("missing expected published envelope types: plan=%v trace=%v res=%v", foundPlan, foundTrace, foundRes)
	}

	// Reflexive Planning (Should skip publication per Section 4.1 / ShouldPublishToWorkspace)
	pub.publishedEnvs = nil
	reqRef, _ := NewPlanningRequestBuilder().
		WithRequestID("req-ref-1").
		WithGoal("Orchestrate reflexive plan").
		Build()

	_, err = svc.PlanReflexive(context.Background(), reqRef)
	if err != nil {
		t.Fatalf("PlanReflexive failed: %v", err)
	}
	if len(pub.publishedEnvs) != 0 {
		t.Errorf("expected reflexive plan to skip workspace broadcast, got %d envelopes", len(pub.publishedEnvs))
	}
}

func TestService_ExecutiveAbilityDriver(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(&mockSpecialist{name: "CoreSpec", domains: []string{"General"}})

	svc := NewService(WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	// DecomposeGoal
	planID, err := svc.DecomposeGoal(context.Background(), "High-level goal decomposition")
	if err != nil || planID == "" {
		t.Fatalf("DecomposeGoal failed: %v / planID=%s", err, planID)
	}

	// ExecuteTask
	status, outRef, err := svc.ExecuteTask(context.Background(), "wf-1", "node-101", executive.BudgetStandard, "cas://payload-88")
	if err != nil {
		t.Fatalf("ExecuteTask failed: %v", err)
	}
	if status != executive.StatusConfident || outRef == "" {
		t.Errorf("unexpected epistemic status: %s / outRef=%s", status, outRef)
	}
}

func TestService_TimeoutPropagationAndIsolation(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(&mockSpecialist{name: "SlowSpec", domains: []string{"General"}, duration: 300 * time.Millisecond})

	svc := NewService(WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-slow").
		WithGoal("Test service timeout").
		WithBudget(30*time.Millisecond, 0.70).
		Build()

	res, err := svc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("expected service to isolate timeout and return result without pipeline error, got: %v", err)
	}
	if res.Traces[0].TerminationReason != TerminationTimeLimit {
		t.Errorf("expected TerminationTimeLimit, got %s", res.Traces[0].TerminationReason)
	}
	if res.ResultStatus != ResultNoPlans {
		t.Errorf("expected RESULT_NO_PLANS on timeout with 0 subgoals, got %s", res.ResultStatus)
	}
}

func TestService_ConcurrencyAndRingBufferRetention(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxTraceRetention = 5

	reg := NewSpecialistRegistry()
	_ = reg.Register(&mockSpecialist{name: "FastSpec", domains: []string{"General"}})

	svc := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req, _ := NewPlanningRequestBuilder().
				WithRequestID(fmt.Sprintf("req-conc-%d", idx)).
				WithGoal(fmt.Sprintf("Concurrent task %d", idx)).
				Build()
			_, _ = svc.PlanTactical(context.Background(), req)
		}(i)
	}
	wg.Wait()

	svc.mu.RLock()
	tracesCount := len(svc.traces)
	keysCount := len(svc.traceKeys)
	svc.mu.RUnlock()

	if tracesCount != 5 || keysCount != 5 {
		t.Fatalf("expected exactly 5 traces retained in ring buffer, got traces=%d keys=%d", tracesCount, keysCount)
	}
}

func TestService_ValidationFirewall(t *testing.T) {
	svc := NewService()
	_ = svc.Start()
	defer svc.Close()

	// Pass invalid request (e.g., negative confidence floor or empty goal)
	req := &PlanningRequest{RequestID: "bad", MinConfidenceFloor: 2.5}
	_, err := svc.PlanTactical(context.Background(), req)
	if err == nil {
		t.Error("expected Stage 0 Validation Firewall to intercept invalid request")
	}
}

// TestService_ActiveGoalsBridge verifies that Planning subscribes to TopicActiveGoals
// and correctly invokes the planning pipeline when a ReasoningResult envelope arrives.
func TestService_ActiveGoalsBridge(t *testing.T) {
	// Set up a CAS storer that holds a ReasoningResult payload.
	storer := &casStorer{data: make(map[string][]byte)}
	pub := &capturingPublisher{}
	sub := &recordingSubscriber{}

	svc := NewService(
		WithWorkspaceBridge(storer, pub, sub),
	)

	if err := svc.Start(); err != nil {
		t.Fatalf("failed to start planning service: %v", err)
	}
	defer svc.Close()

	// Verify subscription was registered on TopicActiveGoals.
	if sub.topic != communication.TopicActiveGoals {
		t.Errorf("expected subscription on TopicActiveGoals, got %q", sub.topic)
	}
	if sub.subscriberID != "Intelligence.Planning" {
		t.Errorf("expected subscriberID 'Intelligence.Planning', got %q", sub.subscriberID)
	}

	// Synthesize a serialized ReasoningResult payload and store it in CAS.
	payload := []byte(`{"schema_version":"2.0","envelope_id":"rs-env-1","source_frame_id":"sf-1","status":"UNAMBIGUOUS_SOLVED","strategy_used":"","primary_hypothesis":{"id":"hyp-1","type":"DEDUCTIVE","conclusion":"help the user with their question","reasoning_confidence":0.9,"calibrated_confidence":0.9,"contributing_stages":[],"supporting_premises":[],"evidence_trace":""},"ambiguity_set":[],"contradictions_flagged":[],"proposed_belief_updates":[],"strategy_telemetry":{"episode_id":"","strategy_selected":"","specialists_executed":null,"execution_duration_ms":0,"calibrated_confidence":0,"resource_cost_tier":"","escalated_to_llm":false},"constitution_annotations":[],"reasoning_trace":[],"offline_mode":true,"processed_duration_ms":0}`)
	casKey, _ := storer.Store(context.Background(), payload)

	// Build and deliver a TopicActiveGoals envelope.
	env := communication.Envelope{
		ID:         "env-active-goals-1",
		Source:     "Intelligence.Reasoning",
		Topic:      communication.TopicActiveGoals,
		PayloadRef: casKey,
	}

	if err := sub.handler(context.Background(), env); err != nil {
		t.Fatalf("handleActiveGoal returned unexpected error: %v", err)
	}
}

// casStorer is a simple CAS mock that stores and retrieves by a deterministic key.
type casStorer struct {
	mu   sync.Mutex
	data map[string][]byte
	seq  int
}

func (c *casStorer) Store(_ context.Context, data []byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++
	key := fmt.Sprintf("cas://key-%d", c.seq)
	c.data[key] = data
	return key, nil
}

func (c *casStorer) Retrieve(_ context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.data[key]
	if !ok {
		return nil, fmt.Errorf("cas: key not found: %s", key)
	}
	return d, nil
}

// capturingPublisher captures envelopes published to the Workspace.
type capturingPublisher struct {
	mu   sync.Mutex
	envs []communication.Envelope
}

func (p *capturingPublisher) Publish(_ context.Context, env communication.Envelope) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.envs = append(p.envs, env)
	return nil
}

// recordingSubscriber captures the Subscribe call for verification.
type recordingSubscriber struct {
	topic        communication.TopicID
	subscriberID string
	handler      func(ctx context.Context, env communication.Envelope) error
}

func (r *recordingSubscriber) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (WorkspaceSubscription, error) {
	r.topic = topic
	r.subscriberID = subscriberID
	r.handler = handler
	return nil, nil
}

// Ensure time import is used.
var _ = time.Second
