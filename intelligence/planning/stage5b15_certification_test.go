package planning

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// faultySpecialist injects artificial errors to test isolation and fault recovery across multi-specialist pipelines.
type faultySpecialist struct {
	name      string
	failCount int32
}

func newFaultySpecialist(name string) *faultySpecialist {
	return &faultySpecialist{name: name}
}

func (s *faultySpecialist) Name() string { return s.name }
func (s *faultySpecialist) SupportedDomains() []string {
	return []string{"General", "FaultyDomain"}
}
func (s *faultySpecialist) Contribute(ctx context.Context, req *PlanningRequest, graph *DependencyGraphSnapshot, profile *PlanningPolicyProfile) (*PlanningStepLog, []Subgoal, []DependencyEdge, error) {
	atomic.AddInt32(&s.failCount, 1)
	return nil, nil, nil, errors.New("faultySpecialist: injected failure for fault injection certification")
}

// TestStage5B15_Certification_EdgeCases verifies boundary behavior across extreme beam widths, cyclic graphs, and disconnected state spaces.
func TestStage5B15_Certification_EdgeCases(t *testing.T) {
	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-s5b15-edge").
		WithGoal("Navigate edge case boundaries").
		WithDomain("General").
		WithTargetDepth(DepthStrategic).
		Build()

	// Scenario 1: Zero Beam Width defaults to 15 safely without panics
	cfgZero := DefaultBeamAStarConfig()
	cfgZero.BeamWidth = 0
	engineZero := NewBeamAStarEngine(cfgZero)
	if engineZero.Config.BeamWidth != 15 {
		t.Errorf("expected BeamWidth=0 to default to 15, got %d", engineZero.Config.BeamWidth)
	}

	// Scenario 2: One Beam Width (Greedy Best-First) retains exactly top candidate per expansion level
	cfgOne := DefaultBeamAStarConfig()
	cfgOne.BeamWidth = 1
	cfgOne.MaxDepth = 10
	engineOne := NewBeamAStarEngine(cfgOne)

	rootOne := NewSearchState("root-one")
	rootOne.EpistemicConfidence = 0.90
	rootOne.RemainingDesiredState["goal"] = req.Goal

	opsOne := []*SearchEdge{
		NewSearchEdge("op-worse", EdgeTypeStrategicOperator, "Worse Op"),
		NewSearchEdge("op-better", EdgeTypeStrategicOperator, "Better Op"),
	}
	opsOne[0].EdgeCost = CostVector{Time: 10 * time.Second, Resources: 50}
	opsOne[0].Postconditions["goal"] = req.Goal
	opsOne[1].EdgeCost = CostVector{Time: 1 * time.Second, Resources: 5}
	opsOne[1].Postconditions["goal"] = req.Goal

	resOne, errOne := engineOne.Search(context.Background(), req, rootOne, opsOne)
	if errOne != nil {
		t.Fatalf("one-beam search failed: %v", errOne)
	}
	if len(resOne.GoalNodes) == 0 {
		t.Errorf("expected goal discovery under greedy one-beam exploration")
	}

	// Scenario 3: Huge Beam Width handles massive frontier capacities bounded by MaxNodes ceiling
	cfgHuge := DefaultBeamAStarConfig()
	cfgHuge.BeamWidth = 10000
	cfgHuge.MaxDepth = 5
	engineHuge := NewBeamAStarEngine(cfgHuge)
	resHuge, errHuge := engineHuge.Search(context.Background(), req, rootOne, opsOne)
	if errHuge != nil {
		t.Fatalf("huge-beam search failed: %v", errHuge)
	}
	if len(resHuge.GoalNodes) == 0 {
		t.Errorf("expected goal discovery under huge beam evaluation")
	}

	// Scenario 4: Cyclic Transition Graph (A -> B -> A) terminates via O(1) visited-state deduplication
	cfgCyclic := DefaultBeamAStarConfig()
	cfgCyclic.BeamWidth = 10
	cfgCyclic.MaxDepth = 50
	engineCyclic := NewBeamAStarEngine(cfgCyclic)

	rootCyclic := NewSearchState("state-A")
	rootCyclic.EpistemicConfidence = 0.90
	rootCyclic.RemainingDesiredState["goal"] = "Unreachable Cyclic Goal"

	opAtoB := NewSearchEdge("op-A-B", EdgeTypeStrategicOperator, "Transition A -> B")
	opAtoB.Postconditions["state_pos"] = "B"
	opAtoB.EdgeCost = CostVector{Time: 1 * time.Second, Resources: 1}

	opBtoA := NewSearchEdge("op-B-A", EdgeTypeStrategicOperator, "Transition B -> A")
	opBtoA.Preconditions["state_pos"] = "B"
	opBtoA.Postconditions["state_pos"] = "A"
	opBtoA.EdgeCost = CostVector{Time: 1 * time.Second, Resources: 1}

	resCyclic, errCyclic := engineCyclic.Search(context.Background(), req, rootCyclic, []*SearchEdge{opAtoB, opBtoA})
	if errCyclic != nil {
		t.Fatalf("cyclic graph search failed: %v", errCyclic)
	}
	if resCyclic.Status != StatusFailedNoPath && resCyclic.Status != StatusPartialBudget {
		t.Errorf("expected cyclic graph to terminate safely with StatusFailedNoPath or StatusPartialBudget, got status %s", resCyclic.Status)
	}

	// Scenario 5: Disconnected State Space (no operators satisfy target condition)
	rootDis := NewSearchState("root-dis")
	rootDis.RemainingDesiredState["unreachable_condition"] = "true"
	resDis, errDis := engineCyclic.Search(context.Background(), req, rootDis, []*SearchEdge{opAtoB})
	if errDis != nil {
		t.Fatalf("disconnected search returned error: %v", errDis)
	}
	if len(resDis.GoalNodes) > 0 {
		t.Errorf("expected 0 goal nodes for disconnected state space, got %d", len(resDis.GoalNodes))
	}
}

// TestStage5B15_Certification_FaultInjection verifies clean isolation when specialists or search operations fail mid-flight.
func TestStage5B15_Certification_FaultInjection(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("STRATEGIC"))
	_ = reg.Register(newFaultySpecialist("FaultyInjectorSpecialist"))

	planSvc := NewService(WithSpecialistRegistry(reg))
	_ = planSvc.Start()
	defer planSvc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-s5b15-fault").
		WithGoal("Achieve goal despite faulty specialist injection").
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := planSvc.PlanTactical(ctx, req)
	if err != nil {
		t.Fatalf("PlanTactical failed when faulty specialist was present: %v", err)
	}

	// The faulty specialist returns an error, which invokeSpecialists captures in StepLog/contribution,
	// while the healthy specialists (HTN, GOAP, TreeSearch) successfully contribute valid plans.
	if len(res.Plans) < 1 {
		t.Errorf("expected healthy specialists to contribute plans despite faulty specialist error, got %d plans", len(res.Plans))
	}
}

// TestStage5B15_Certification_HighConcurrency1000Workers verifies zero data races, deadlocks, or shared state contamination across 1,000 concurrent workers.
func TestStage5B15_Certification_HighConcurrency1000Workers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 1000 worker concurrency stress test in short mode")
	}

	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("TACTICAL"))

	planSvc := NewService(WithSpecialistRegistry(reg))
	_ = planSvc.Start()
	defer planSvc.Close()

	const workers = 1000
	var wg sync.WaitGroup
	var successCount int64
	var failCount int64

	domains := []string{"General", "Robotics", "Coding", "MissionControl"}

	start := time.Now()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			domain := domains[workerID%len(domains)]
			req, _ := NewPlanningRequestBuilder().
				WithRequestID(fmt.Sprintf("req-s5b15-1000w-%d", workerID)).
				WithGoal(fmt.Sprintf("Concurrent certification task %d", workerID)).
				WithDomain(domain).
				WithTargetDepth(DepthTactical).
				Build()

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			res, err := planSvc.PlanTactical(ctx, req)
			if err != nil || res == nil || len(res.Plans) == 0 {
				atomic.AddInt64(&failCount, 1)
			} else {
				atomic.AddInt64(&successCount, 1)
			}
		}(w)
	}
	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Completed 1,000 concurrent planning requests across 3 specialists in %v (Success: %d, Fail: %d)", elapsed, successCount, failCount)
	if failCount > 0 {
		t.Errorf("expected 0 failures across 1,000 concurrent workers, got %d failures", failCount)
	}
}
