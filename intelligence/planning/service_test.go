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

	svc := NewService(
		WithSpecialistRegistry(reg),
		WithWorkspaceBridge(storer, pub),
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
