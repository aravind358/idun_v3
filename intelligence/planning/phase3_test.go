package planning

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestPhase3_HTNSpecialist_StrategyConsumption(t *testing.T) {
	profile := DefaultPlanningPolicyProfile()
	// Override strategy to verify strict consumption
	profile.SearchStrategies["TACTICAL"].MaxDepth = 3
	profile.SearchStrategies["TACTICAL"].MaxNodes = 10

	spec := NewHTNSpecialist("TACTICAL")
	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-htn-1").
		WithGoal("Deploy distributed database cluster").
		Build()

	log, subgoals, edges, err := spec.Contribute(context.Background(), req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected error from HTNSpecialist: %v", err)
	}
	if log == nil || len(subgoals) != 3 || len(edges) != 2 {
		t.Errorf("expected 3 subgoals (bounded by MaxDepth=3) and 2 edges, got subgoals=%d edges=%d log=%+v", len(subgoals), len(edges), log)
	}
	if !strings.Contains(log.ActionPerformed, "max_depth=3") {
		t.Errorf("expected log to reference max_depth=3, got: %s", log.ActionPerformed)
	}

	// Test budget cutoff when existing nodes reach strategy MaxNodes
	nodesMap := make(map[string]string, 10)
	for i := 0; i < 10; i++ {
		nodesMap[fmt.Sprintf("node-%d", i)] = "Existing node"
	}
	graph := &DependencyGraphSnapshot{
		Nodes: nodesMap, // Exactly matches MaxNodes=10
	}
	logCutoff, sgCutoff, _, err := spec.Contribute(context.Background(), req, graph, profile)
	if err != nil {
		t.Fatalf("unexpected error on cutoff check: %v", err)
	}
	if len(sgCutoff) != 0 || !strings.Contains(logCutoff.ActionPerformed, "reached strategy MaxNodes") {
		t.Errorf("expected expansion cutoff log, got sg=%d log=%+v", len(sgCutoff), logCutoff)
	}
}

func TestPhase3_GOAPSpecialist_StrategyConsumption(t *testing.T) {
	profile := DefaultPlanningPolicyProfile()
	profile.SearchStrategies["TACTICAL"].BeamWidth = 2

	spec := NewGOAPSpecialist("TACTICAL")
	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-goap-1").
		WithGoal("Rebalance server load").
		Build()

	log, subgoals, edges, err := spec.Contribute(context.Background(), req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected error from GOAPSpecialist: %v", err)
	}
	// actionCount = BeamWidth * 2 = 4
	if len(subgoals) != 4 || len(edges) != 3 {
		t.Errorf("expected 4 subgoals and 3 edges, got subgoals=%d edges=%d log=%+v", len(subgoals), len(edges), log)
	}
	if !strings.Contains(log.ActionPerformed, "beam_width=2") {
		t.Errorf("expected log to reference beam_width=2, got: %s", log.ActionPerformed)
	}
}

func TestPhase3_TreeSearchSpecialist_ParallelExecution(t *testing.T) {
	profile := DefaultPlanningPolicyProfile()
	profile.SearchStrategies["STRATEGIC"].BeamWidth = 5
	profile.SearchStrategies["STRATEGIC"].AllowParallelExpansion = true
	profile.SearchStrategies["STRATEGIC"].MaxConcurrentWorkers = 4

	spec := NewTreeSearchSpecialist("STRATEGIC")
	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-tree-1").
		WithGoal("Orchestrate multi-region failover").
		Build()

	log, subgoals, _, err := spec.Contribute(context.Background(), req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected error from TreeSearchSpecialist: %v", err)
	}
	if len(subgoals) != 5 {
		t.Errorf("expected 5 branches explored, got %d", len(subgoals))
	}
	if !strings.Contains(log.ActionPerformed, "4 concurrent workers") {
		t.Errorf("expected log to reference 4 concurrent workers, got: %s", log.ActionPerformed)
	}
}

func TestPhase3_OrchestratedPipelineWithDomainEngines(t *testing.T) {
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("STRATEGIC"))

	svc := NewService(WithSpecialistRegistry(reg))
	_ = svc.Start()
	defer svc.Close()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("req-full-phase3").
		WithGoal("Build and deploy autonomous agent swarm").
		WithDomain("Coding").
		Build()

	res, err := svc.PlanTactical(context.Background(), req)
	if err != nil {
		t.Fatalf("PlanTactical with Phase 3 domain engines failed: %v", err)
	}
	if res.ResultStatus != ResultSuccess {
		t.Errorf("expected RESULT_SUCCESS, got %s", res.ResultStatus)
	}
	if len(res.Plans) == 0 || len(res.Plans[0].Subgoals) == 0 {
		t.Fatalf("expected non-empty subgoals produced by Phase 3 engines, got %+v", res.Plans)
	}

	traceID := res.Plans[0].TraceID
	trace, found := svc.GetTrace(traceID)
	if !found {
		t.Fatalf("expected trace %s to be stored in ring buffer", traceID)
	}
	if len(trace.PlanningSteps) < 2 {
		t.Errorf("expected at least 2 planning step logs from TACTICAL specialists (HTN + GOAP), got %d", len(trace.PlanningSteps))
	}
}

func TestPlanningCapabilities_ValidationAndClamping(t *testing.T) {
	caps := DefaultPlanningCapabilities()
	if err := caps.Validate(); err != nil {
		t.Fatalf("expected DefaultPlanningCapabilities to be valid, got %v", err)
	}

	caps.MaxPlanningDepth = 0
	if err := caps.Validate(); err == nil {
		t.Error("expected error on MaxPlanningDepth = 0, got nil")
	}

	// Verify specialist capability check skipping
	profile := DefaultPlanningPolicyProfile()
	profile.Capabilities.SupportsHTN = false

	spec := NewHTNSpecialist("TACTICAL")
	req, _ := NewPlanningRequestBuilder().WithRequestID("req-cap-1").WithGoal("Test skip").Build()
	log, sg, edges, err := spec.Contribute(context.Background(), req, nil, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sg) != 0 || len(edges) != 0 || !strings.Contains(log.ActionPerformed, "engine capabilities do not support HTN") {
		t.Errorf("expected HTN skip log, got log=%+v sg=%d", log, len(sg))
	}
}

func TestPlanningService_Capabilities(t *testing.T) {
	svc := NewService()
	c := svc.Capabilities()
	if c == nil || !c.SupportsHTN {
		t.Errorf("expected default service capabilities, got %+v", c)
	}
}
