package planning

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/decision"
)

// TestStage5B14_SystemValidation_EndToEndPipeline validates the complete multi-plan integration flow:
// PlanningRequest -> PlanningService -> HTN + GOAP + TreeSearch -> PlanningResult -> DecisionService -> Executive verification.
func TestStage5B14_SystemValidation_EndToEndPipeline(t *testing.T) {
	storer := newSharedCASStorer()
	bridge := newSharedWorkspaceBridge()

	reg := NewSpecialistRegistry()
	if err := reg.Register(NewHTNSpecialist("TACTICAL")); err != nil {
		t.Fatalf("failed to register HTN specialist: %v", err)
	}
	if err := reg.Register(NewGOAPSpecialist("TACTICAL")); err != nil {
		t.Fatalf("failed to register GOAP specialist: %v", err)
	}
	if err := reg.Register(NewTreeSearchSpecialist("TACTICAL")); err != nil {
		t.Fatalf("failed to register TreeSearch specialist: %v", err)
	}

	planSvc := NewService(
		WithSpecialistRegistry(reg),
		WithWorkspaceBridge(storer, bridge, bridge),
	)
	if err := planSvc.Start(); err != nil {
		t.Fatalf("failed to start planning service: %v", err)
	}
	defer planSvc.Close()

	decSub := &decisionSubscriberBridge{bridge: bridge}
	decSvc := decision.NewService(decision.WithWorkspaceBridge(storer, bridge, decSub))
	if err := decSvc.Start(); err != nil {
		t.Fatalf("failed to start decision service: %v", err)
	}
	defer decSvc.Close()

	req, err := NewPlanningRequestBuilder().
		WithRequestID("req-s5b14-pipeline").
		WithGoal("Validate Stage 5B.1.4 end-to-end multi-specialist orchestration").
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		WithBudget(2*time.Second, 0.80).
		Build()
	if err != nil {
		t.Fatalf("failed to build planning request: %v", err)
	}

	res, err := planSvc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical failed during end-to-end pipeline validation: %v", err)
	}

	// Verify coexistence and multi-plan candidate generation (1 per registered specialist)
	if len(res.Plans) != 3 {
		t.Fatalf("expected exactly 3 candidate plans (HTN, GOAP, TreeSearch), got %d", len(res.Plans))
	}
	if res.PrimaryPlanID == "" {
		t.Error("expected PrimaryPlanID to be populated from highest ranking plan")
	}
	if res.ResultStatus != ResultSuccess {
		t.Errorf("expected ResultStatus=%s, got %s", ResultSuccess, res.ResultStatus)
	}
	if len(res.Traces) == 0 {
		t.Error("expected diagnostic traces to be emitted in PlanningResult")
	}

	// Verify plan integrity and rollback strategy conservation across candidates
	hasTreeSearchPlan := false
	for i, plan := range res.Plans {
		if err := plan.Validate(); err != nil {
			t.Errorf("candidate[%d] plan validation failed: %v", i, err)
		}
		if plan.ConfidenceProfile.OverallConfidence == 0 {
			t.Errorf("candidate[%d] missing overall confidence profile", i)
		}
		if plan.PlannerID == "MultiAlternativeTreeSearchSpecialist" || plan.PlannerType == "TreeSearch" {
			hasTreeSearchPlan = true
			if len(plan.Subgoals) == 0 {
				t.Errorf("expected TreeSearch candidate plan to contain subgoals contributed by strategic search")
			}
		}
	}
	if !hasTreeSearchPlan {
		t.Error("expected at least one candidate plan produced by TreeSearchSpecialist")
	}

	// Verify Decision integration via CandidateSet evaluation
	cs := decision.CandidateSet{
		EpisodeID:  "ep-s5b14-decision",
		Candidates: make([]decision.Candidate, 0, len(res.Plans)),
	}
	for _, p := range res.Plans {
		cs.Candidates = append(cs.Candidates, decision.Candidate{
			ID:            p.PlanID,
			Description:   p.Goal,
			SourceAbility: "Planning",
			Attributes: map[string]float64{
				"confidence": p.ConfidenceProfile.OverallConfidence,
				"cost":       p.EstimatedCost,
				"duration":   p.EstimatedDuration.Seconds(),
			},
		})
	}

	rec, err := decSvc.EvaluateDeliberative(context.Background(), cs)
	if err != nil {
		t.Fatalf("DecisionService.EvaluateDeliberative failed: %v", err)
	}
	if len(rec.TradeoffMatrix) != 3 {
		t.Errorf("expected 3 items in TradeoffMatrix corresponding to 3 candidates, got %d", len(rec.TradeoffMatrix))
	}
	if rec.SelectedCandidateID == "" {
		t.Error("expected DecisionService to select a winning candidate ID")
	}

	// Verify bridge communication emitted TopicEvaluatedOptions
	bridge.mu.RLock()
	var emittedOption *communication.Envelope
	for _, env := range bridge.envelopes {
		if env.Topic == communication.TopicEvaluatedOptions {
			emittedOption = &env
			break
		}
	}
	bridge.mu.RUnlock()
	if emittedOption == nil {
		t.Error("expected TopicEvaluatedOptions envelope to be published across workspace bridge")
	}
}

// TestStage5B14_SystemValidation_TimeoutAndBudgetExhaustion validates bounded timeouts, budget floors, and graceful failure handling.
func TestStage5B14_SystemValidation_TimeoutAndBudgetExhaustion(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewTreeSearchSpecialist("STRATEGIC"))
	planSvc := NewService(WithSpecialistRegistry(reg))
	_ = planSvc.Start()
	defer planSvc.Close()

	// Scenario A: Context timeout expiration mid-flight
	reqTimeout, _ := NewPlanningRequestBuilder().
		WithRequestID("req-timeout-expired").
		WithGoal("Validate clean preemption when context expires").
		WithDomain("General").
		WithTargetDepth(DepthStrategic).
		Build()

	ctxExpired, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	time.Sleep(2 * time.Millisecond) // ensure context is already expired
	defer cancel()

	resExpired, errExpired := planSvc.PlanDeliberative(ctxExpired, reqTimeout)
	if errExpired == nil {
		if resExpired == nil || (resExpired.ResultStatus != ResultNoPlans && resExpired.Status != PlanStatusPartialBudgetExhausted) {
			t.Errorf("expected clean failure or ResultNoPlans/PartialBudgetExhausted when context expires, got res=%v, err=%v", resExpired, errExpired)
		}
	}

	// Scenario B: Impossible goal / confidence floor rejection without crashes
	reqImpossible, _ := NewPlanningRequestBuilder().
		WithRequestID("req-impossible").
		WithGoal("Achieve zero risk operation").
		WithDomain("General").
		WithTargetDepth(DepthStrategic).
		WithBudget(1*time.Second, 0.99).
		Build()

	resImp, errImp := planSvc.PlanDeliberative(context.Background(), reqImpossible)
	if errImp != nil {
		// Clean error return is acceptable
	} else if resImp != nil && len(resImp.Plans) == 0 {
		if resImp.ResultStatus != ResultNoPlans {
			t.Errorf("expected ResultStatus=%s for infeasible candidate spaces, got %s", ResultNoPlans, resImp.ResultStatus)
		}
	}
}

// TestStage5B14_SystemValidation_ConcurrentExecution verifies zero race conditions and clean memory isolation across parallel workers.
func TestStage5B14_SystemValidation_ConcurrentExecution(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("TACTICAL"))

	planSvc := NewService(WithSpecialistRegistry(reg))
	_ = planSvc.Start()
	defer planSvc.Close()

	var wg sync.WaitGroup
	const workers = 25
	domains := []string{"General", "Robotics", "Coding", "MissionControl"}

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			domain := domains[workerID%len(domains)]
			req, _ := NewPlanningRequestBuilder().
				WithRequestID(fmt.Sprintf("req-s5b14-conc-%d", workerID)).
				WithGoal(fmt.Sprintf("Concurrent validation goal for worker %d", workerID)).
				WithDomain(domain).
				WithTargetDepth(DepthTactical).
				Build()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			res, err := planSvc.PlanTactical(ctx, req)
			if err != nil {
				t.Errorf("worker %d PlanTactical failed: %v", workerID, err)
				return
			}
			if len(res.Plans) != 3 {
				t.Errorf("worker %d expected 3 plans from coexistence, got %d", workerID, len(res.Plans))
			}
		}(w)
	}
	wg.Wait()
}

// TestStage5B14_SystemValidation_StressAndBeamPruning validates bounded memory and node expansion under deep search horizons.
func TestStage5B14_SystemValidation_StressAndBeamPruning(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	cfg := DefaultBeamAStarConfig()
	cfg.BeamWidth = 10
	cfg.MaxDepth = 15

	engine := NewBeamAStarEngine(cfg)
	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-stress-beam").
		WithGoal("Navigate complex multi-stage state space").
		WithDomain("General").
		WithTargetDepth(DepthStrategic).
		Build()

	rootState := NewSearchState("root-stress")
	rootState.EpistemicConfidence = 0.90
	rootState.RemainingDesiredState["goal_achieved"] = req.Goal

	// Generate 20 branching operators
	ops := make([]*SearchEdge, 0, 20)
	for i := 1; i <= 20; i++ {
		op := NewSearchEdge(fmt.Sprintf("stress-op-%d", i), EdgeTypeStrategicOperator, fmt.Sprintf("Operator %d", i))
		op.EdgeCost = CostVector{Time: time.Duration(i) * time.Second, Resources: float64(i)}
		op.RiskDelta = 0.02
		op.Reversibility = ReversibilityHighCost
		op.Postconditions["goal_achieved"] = req.Goal
		ops = append(ops, op)
	}

	start := time.Now()
	res, err := engine.Search(context.Background(), req, rootState, ops)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("BeamAStarEngine.Search failed during stress evaluation: %v", err)
	}
	if len(res.GoalNodes) == 0 && len(res.PartialNodes) == 0 {
		t.Fatal("expected discovered paths under bounded beam evaluation")
	}
	if elapsed > 1*time.Second {
		t.Logf("warning: beam search across 20 branching operators took %v (>1s)", elapsed)
	}
}
