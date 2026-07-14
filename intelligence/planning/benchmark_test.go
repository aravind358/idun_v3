package planning

import (
	"context"
	"testing"
)

// BenchmarkPlanningService_ReflexivePlanning benchmarks Phase 1 Reflexive planning (<2ms target).
func BenchmarkPlanningService_ReflexivePlanning(b *testing.B) {
	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	defer service.Close()
	_ = service.Start()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("bench-refl").
		WithGoal("Resolve emergency collision alert immediately").
		WithDomain("General").
		WithTargetDepth(DepthReflexive).
		Build()

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.PlanReflexive(ctx, req)
		if err != nil {
			b.Fatalf("PlanReflexive failed: %v", err)
		}
	}
}

// BenchmarkPlanningService_TacticalPlanning benchmarks Phase 1/2 Tactical planning.
func BenchmarkPlanningService_TacticalPlanning(b *testing.B) {
	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	defer service.Close()
	_ = service.Start()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("bench-tact").
		WithGoal("Execute multi-stage deployment").
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		Build()

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.PlanTactical(ctx, req)
		if err != nil {
			b.Fatalf("PlanTactical failed: %v", err)
		}
	}
}

// BenchmarkPlanningService_StrategicDeliberative benchmarks Phase 1/2/3 Deliberative/Strategic planning.
func BenchmarkPlanningService_StrategicDeliberative(b *testing.B) {
	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("STRATEGIC"))
	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	defer service.Close()
	_ = service.Start()

	req, _ := NewPlanningRequestBuilder().
		WithRequestID("bench-strat").
		WithGoal("Develop long-term architectural transformation plan").
		WithDomain("General").
		WithTargetDepth(DepthStrategic).
		Build()

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.PlanDeliberative(ctx, req)
		if err != nil {
			b.Fatalf("PlanDeliberative failed: %v", err)
		}
	}
}

// BenchmarkHTNSpecialist_Contribute benchmarks direct HTN decomposition evaluation.
func BenchmarkHTNSpecialist_Contribute(b *testing.B) {
	s := NewHTNSpecialist("TACTICAL")
	req, _ := NewPlanningRequestBuilder().
		WithRequestID("bench-htn").
		WithGoal("Decompose complex goal into subgoals").
		WithDomain("General").
		Build()
	graph := &DependencyGraphSnapshot{Nodes: make(map[string]string), Edges: []DependencyEdge{}}
	profile := DefaultPlanningPolicyProfile()

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _, err := s.Contribute(ctx, req, graph, profile)
		if err != nil {
			b.Fatalf("HTN Contribute failed: %v", err)
		}
	}
}

// BenchmarkGOAPSpecialist_Contribute benchmarks direct GOAP action sequence evaluation.
func BenchmarkGOAPSpecialist_Contribute(b *testing.B) {
	s := NewGOAPSpecialist("TACTICAL")
	req, _ := NewPlanningRequestBuilder().
		WithRequestID("bench-goap").
		WithGoal("Execute state transition sequence").
		WithDomain("General").
		Build()
	graph := &DependencyGraphSnapshot{Nodes: make(map[string]string), Edges: []DependencyEdge{}}
	profile := DefaultPlanningPolicyProfile()

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _, err := s.Contribute(ctx, req, graph, profile)
		if err != nil {
			b.Fatalf("GOAP Contribute failed: %v", err)
		}
	}
}

// BenchmarkTreeSearchSpecialist_Contribute benchmarks parallel tree search evaluation.
func BenchmarkTreeSearchSpecialist_Contribute(b *testing.B) {
	s := NewTreeSearchSpecialist("STRATEGIC")
	req, _ := NewPlanningRequestBuilder().
		WithRequestID("bench-tree").
		WithGoal("Evaluate multi-alternative contingency paths").
		WithDomain("General").
		Build()
	graph := &DependencyGraphSnapshot{Nodes: make(map[string]string), Edges: []DependencyEdge{}}
	profile := DefaultPlanningPolicyProfile()

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _, err := s.Contribute(ctx, req, graph, profile)
		if err != nil {
			b.Fatalf("TreeSearch Contribute failed: %v", err)
		}
	}
}

// BenchmarkPlanFingerprinter_ComputeFingerprint benchmarks SHA-256 structural plan hashing.
func BenchmarkPlanFingerprinter_ComputeFingerprint(b *testing.B) {
	fp := NewDefaultPlanFingerprinter()
	cp := ConfidenceProfile{
		GoalConfidence: 0.9, PreconditionConfidence: 0.9, DependencyConfidence: 0.9,
		ResourceConfidence: 0.9, TimingConfidence: 0.9, ConstraintConfidence: 0.9,
		OverallConfidence: 0.9,
	}
	plan, err := NewPlanBuilder().
		WithIdentity("plan-bench", "snap-bench", "trace-bench").
		WithGoalAndDomain("Benchmark plan hashing", "General", "TACTICAL").
		AddSubgoal(Subgoal{SubgoalID: "sg-1", Title: "Step 1", Description: "Step 1 description"}).
		AddSubgoal(Subgoal{SubgoalID: "sg-2", Title: "Step 2", Description: "Step 2 description"}).
		AddDependency(DependencyEdge{EdgeID: "e-1", SourceNodeID: "sg-1", TargetNodeID: "sg-2", DependencyType: "CAUSAL"}).
		WithConfidenceProfile(cp).
		WithReplayMetadata(ReplayMetadata{StrategySnapshotID: "snap-bench", ReplayFidelity: "EXACT"}).
		Build()

	if err != nil {
		b.Fatalf("failed to build benchmark plan: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := fp.ComputeFingerprint(plan)
		if err != nil {
			b.Fatalf("ComputeFingerprint failed: %v", err)
		}
	}
}

// BenchmarkCapabilityFingerprint_Compute benchmarks CapabilityFingerprint SHA-256 calculation.
func BenchmarkCapabilityFingerprint_Compute(b *testing.B) {
	caps := DefaultPlanningCapabilities()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ComputeCapabilityFingerprint(caps)
	}
}

// BenchmarkPlanningTrace_Validate benchmarks structural validation of complete traces.
func BenchmarkPlanningTrace_Validate(b *testing.B) {
	trace, _ := NewPlanningTraceBuilder().
		WithIdentity("trace-bench", "plan-bench", "snap-bench").
		WithDiagnostics(TerminationGoalFound, SearchStatistics{NodesExpanded: 100}, 5.0, ConfidenceProfile{OverallConfidence: 0.9, TimingConfidence: 0.9, ResourceConfidence: 0.9, GoalConfidence: 0.9, PreconditionConfidence: 0.9, DependencyConfidence: 0.9, ConstraintConfidence: 0.9}, QualityMetrics{Completeness: 1.0, Efficiency: 0.9, Robustness: 0.9}).
		WithProvenance("policy-fp-123", "cap-fp-123", "strat-123", ReplayMetadata{StrategySnapshotID: "snap-bench", ReplayFidelity: "EXACT", ReplaySeed: 42}).
		Build()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := trace.Validate(); err != nil {
			b.Fatalf("Trace Validate failed: %v", err)
		}
	}
}

// BenchmarkStrategyProvider_ActiveSnapshot benchmarks lock-free atomic pointer snapshot access.
func BenchmarkStrategyProvider_ActiveSnapshot(b *testing.B) {
	prov := NewDefaultStrategyProvider(nil)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = prov.ActiveSnapshot()
	}
}

// BenchmarkReflexivePlanningCache_Operations benchmarks concurrent cache lookup and storage.
func BenchmarkReflexivePlanningCache_Operations(b *testing.B) {
	cache := NewReflexivePlanningCache("ep-bench", "snap-bench")
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.RecordEvaluation(true, 10, 2, 0, 0, 5, 120)
			} else {
				cache.RecordEvaluation(false, 50, 10, 1, 0, 10, 450)
			}
			i++
		}
	})
}
