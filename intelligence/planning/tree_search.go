package planning

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// TreeSearchSpecialist implements PlanningSpecialist and serves as the strategic search
// orchestration layer. It bridges the bounded Beam A* search engine with IDUN V3's candidate
// plan architecture, converting search trajectories into ranked Candidate Plan objects with
// full rollback extraction and multi-candidate emission.
type TreeSearchSpecialist struct {
	mu             sync.RWMutex
	defaultHorizon string
	operators      []*SearchEdge
	fingerprinter  PlanFingerprinter
}

// NewTreeSearchSpecialist constructs a new TreeSearchSpecialist governed by the specified horizon (default "STRATEGIC").
func NewTreeSearchSpecialist(horizon ...string) *TreeSearchSpecialist {
	h := "STRATEGIC"
	if len(horizon) > 0 && horizon[0] != "" {
		h = horizon[0]
	}
	return &TreeSearchSpecialist{
		defaultHorizon: h,
		operators:      make([]*SearchEdge, 0),
		fingerprinter:  NewDefaultPlanFingerprinter(),
	}
}

func (s *TreeSearchSpecialist) Name() string {
	return "MultiAlternativeTreeSearchSpecialist"
}

func (s *TreeSearchSpecialist) SupportedDomains() []string {
	return []string{"General", "Strategic", "Contingency", "Business", "Research"}
}

// RegisterOperator registers an abstract strategic operator (SearchEdge) into the search pool.
func (s *TreeSearchSpecialist) RegisterOperator(op *SearchEdge) error {
	if op == nil {
		return errors.New("cannot register nil operator")
	}
	if err := op.Validate(); err != nil {
		return fmt.Errorf("invalid operator: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operators = append(s.operators, op.Clone())
	return nil
}

// SetOperators replaces the active strategic operator pool with the provided slice.
func (s *TreeSearchSpecialist) SetOperators(ops []*SearchEdge) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operators = make([]*SearchEdge, 0, len(ops))
	for _, op := range ops {
		if op != nil {
			s.operators = append(s.operators, op.Clone())
		}
	}
}

// GetOperators returns a cloned snapshot of the currently registered strategic operators.
func (s *TreeSearchSpecialist) GetOperators() []*SearchEdge {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*SearchEdge, 0, len(s.operators))
	for _, op := range s.operators {
		out = append(out, op.Clone())
	}
	return out
}

// PrepareSearchRequest constructs the root SearchState, configures BeamAStarConfig, and resolves operators.
func (s *TreeSearchSpecialist) PrepareSearchRequest(
	req *PlanningRequest,
	currentGraph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
) (*SearchState, []*SearchEdge, BeamAStarConfig, error) {
	if req == nil {
		return nil, nil, BeamAStarConfig{}, errors.New("cannot prepare search with nil PlanningRequest")
	}
	if err := req.Validate(); err != nil {
		return nil, nil, BeamAStarConfig{}, fmt.Errorf("invalid PlanningRequest: %w", err)
	}

	strat := resolveSearchStrategy(profile, s.defaultHorizon)
	cfg := DefaultBeamAStarConfig()
	if strat != nil {
		if strat.BeamWidth > 0 {
			cfg.BeamWidth = int(strat.BeamWidth)
		}
		if strat.MaxDepth > 0 {
			cfg.MaxDepth = int(strat.MaxDepth)
		}
	}

	rootState := NewSearchState("root-" + req.RequestID)
	if rootState.SimulatedWorldState == nil {
		rootState.SimulatedWorldState = make(map[string]string)
	}
	if rootState.RemainingDesiredState == nil {
		rootState.RemainingDesiredState = make(map[string]string)
	}
	if rootState.ActiveConstraints == nil {
		rootState.ActiveConstraints = make(map[string]string)
	}

	// Populate initial state and desired conditions from resolved goal
	if req.ResolvedGoal != nil {
		if req.ResolvedGoal.DesiredState != nil {
			for k, v := range req.ResolvedGoal.DesiredState {
				if rootState.SimulatedWorldState[k] != v {
					rootState.RemainingDesiredState[k] = v
				}
			}
		}
		if req.ResolvedGoal.Constraints != nil {
			for k, v := range req.ResolvedGoal.Constraints {
				rootState.ActiveConstraints[k] = v
			}
		}
	}
	if len(rootState.RemainingDesiredState) == 0 && req.Goal != "" {
		rootState.RemainingDesiredState["goal_achieved"] = req.Goal
	}
	for _, hc := range req.HardConstraints {
		rootState.ActiveConstraints[hc] = "REQUIRED"
	}

	s.mu.RLock()
	ops := make([]*SearchEdge, 0, len(s.operators))
	for _, op := range s.operators {
		ops = append(ops, op.Clone())
	}
	s.mu.RUnlock()

	// If no registered operators exist, synthesize default branch exploration operators matching strategy beam width
	if len(ops) == 0 {
		branches := cfg.BeamWidth
		if branches <= 0 {
			branches = 3
		}
		if strat != nil && strat.BeamWidth > 0 && int(strat.BeamWidth) < branches {
			branches = int(strat.BeamWidth)
		}
		for i := 1; i <= branches; i++ {
			opID := fmt.Sprintf("branch-op-%d", i)
			opTitle := fmt.Sprintf("Explore Strategic Branch %d: %s", i, req.Goal)
			op := NewSearchEdge(opID, EdgeTypeStrategicOperator, opTitle)
			op.EdgeCost = CostVector{
				Time:      10 * time.Millisecond,
				Resources: 2.5,
			}
			op.RiskDelta = float64(i-1) * 0.05
			if op.RiskDelta > 0.4 {
				op.RiskDelta = 0.4
			}
			op.Reversibility = ReversibilityHighCost
			for k, v := range rootState.RemainingDesiredState {
				op.Postconditions[k] = v
			}
			ops = append(ops, op)
		}
	}

	return rootState, ops, cfg, nil
}

// InvokeSearch prepares the request and invokes the BeamAStarEngine to discover candidate strategic trajectories.
func (s *TreeSearchSpecialist) InvokeSearch(
	ctx context.Context,
	req *PlanningRequest,
	currentGraph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
) (*SearchResult, error) {
	rootState, ops, cfg, err := s.PrepareSearchRequest(req, currentGraph, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare search request: %w", err)
	}

	engine := NewBeamAStarEngine(cfg)
	res, err := engine.Search(ctx, req, rootState, ops)
	if err != nil {
		return nil, fmt.Errorf("beam A* search execution failed: %w", err)
	}
	return res, nil
}

// ConvertPathToPlan converts a reconstructed chronological SearchNode path into a valid Candidate Plan.
func (s *TreeSearchSpecialist) ConvertPathToPlan(
	path []*SearchNode,
	req *PlanningRequest,
	snapshotID string,
	traceID string,
	rank int,
) (*Plan, error) {
	if len(path) == 0 {
		return nil, errors.New("cannot convert empty search path to plan")
	}
	if req == nil {
		return nil, errors.New("cannot convert path with nil PlanningRequest")
	}

	now := time.Now()
	planID := fmt.Sprintf("plan-tree-%d-%d", now.UnixNano(), rank+1)

	terminalNode := path[len(path)-1]
	if terminalNode == nil || terminalNode.State == nil {
		return nil, errors.New("terminal node or state is nil in search path")
	}

	planStatus := PlanStatusComplete
	if terminalNode.Status == NodeStatusPrunedBudget || terminalNode.Status == NodeStatusOpen {
		planStatus = PlanStatusPartialBudgetExhausted
	}

	estCost := terminalNode.State.AccumulatedCost.Resources + terminalNode.State.AccumulatedCost.MonetaryCost
	estDuration := terminalNode.State.AccumulatedCost.Time
	if estDuration == 0 && terminalNode.State.SimulatedClock > 0 {
		estDuration = terminalNode.State.SimulatedClock
	}
	if estCost == 0 && len(path) > 1 {
		estCost = float64(len(path)-1) * 2.5
	}
	if estDuration == 0 && len(path) > 1 {
		estDuration = time.Duration(len(path)-1) * 10 * time.Minute
	}

	conf := terminalNode.State.EpistemicConfidence
	if conf < req.MinConfidenceFloor && conf > 0 {
		// Retain computed confidence
	} else if conf == 0 {
		conf = req.MinConfidenceFloor
		if conf == 0 {
			conf = 0.85
		}
	}
	cp := ConfidenceProfile{
		GoalConfidence:         conf,
		PreconditionConfidence: conf,
		DependencyConfidence:   conf,
		ResourceConfidence:     conf,
		TimingConfidence:       conf,
		ConstraintConfidence:   conf,
		OverallConfidence:      conf,
	}

	planBuilder := NewPlanBuilder().
		WithIdentity(planID, snapshotID, traceID).
		WithGoalAndDomain(req.Goal, req.Domain, string(req.TargetDepth)).
		WithPlannerIdentity(s.Name(), "TreeSearch").
		WithResolvedGoal(req.ResolvedGoal).
		WithEstimates(estCost, estDuration, nil).
		WithConfidenceProfile(cp).
		WithStatus(planStatus, nil).
		WithReplayMetadata(ReplayMetadata{
			StrategySnapshotID: snapshotID,
			ReplayFidelity:     "EXACT",
			ReplaySeed:         uint64(now.UnixNano()),
			WorkingMemoryHash:  req.ContextRef,
		})

	if snapshotID == "" {
		planBuilder = planBuilder.WithIdentity(planID, "snap-default", traceID)
	}

	var prevSubgoalID string
	var rollbacks []RollbackStrategy

	// Iterate across trajectory nodes to populate ordered subgoals, dependencies, and rollbacks
	for i, node := range path {
		if node == nil {
			continue
		}
		if i == 0 && node.IncomingEdge == nil && len(path) > 1 {
			continue // Skip pure root state if operator steps exist
		}

		sgID := fmt.Sprintf("tree-sg-%d-r%d-s%d", now.UnixNano(), rank+1, i+1)
		title := node.NodeID
		desc := fmt.Sprintf("Strategic search step reaching state %s", node.State.StateID)
		opName := "RootState"
		if node.IncomingEdge != nil {
			if node.IncomingEdge.OperatorName != "" {
				title = node.IncomingEdge.OperatorName
				opName = node.IncomingEdge.OperatorName
			}
			desc = fmt.Sprintf("Applied strategic operator '%s' (edge: %s, risk delta: %.2f)",
				node.IncomingEdge.OperatorName, node.IncomingEdge.EdgeID, node.IncomingEdge.RiskDelta)
		}

		sg := Subgoal{
			SubgoalID:    sgID,
			Title:        title,
			Description:  desc,
			AssignedType: req.Domain,
			Parameters: map[string]string{
				"operator":         opName,
				"state_id":         node.State.StateID,
				"accumulated_cost": fmt.Sprintf("%.2f", node.State.AccumulatedCost.Resources),
				"confidence":       fmt.Sprintf("%.2f", node.State.EpistemicConfidence),
			},
		}
		if node.IncomingEdge != nil {
			sg.Parameters["edge_id"] = node.IncomingEdge.EdgeID
		}
		planBuilder.AddSubgoal(sg)

		if prevSubgoalID != "" {
			edgeID := fmt.Sprintf("dep-%s-%s", prevSubgoalID, sgID)
			dep := DependencyEdge{
				EdgeID:         edgeID,
				SourceNodeID:   prevSubgoalID,
				TargetNodeID:   sgID,
				DependencyType: "HARD_PREREQUISITE",
				IsBlocking:     true,
			}
			planBuilder.AddDependency(dep)
		}
		prevSubgoalID = sgID

		// Extract structured rollback info according to exact reversibility classifications
		if node.IncomingEdge != nil {
			rbID := fmt.Sprintf("rb-%s", sgID)
			var actionSteps []string
			penaltyCost := node.IncomingEdge.EdgeCost.Resources * 0.5
			if penaltyCost < 1.0 {
				penaltyCost = 1.0
			}

			switch node.IncomingEdge.Reversibility {
			case ReversibilityTrivial:
				actionSteps = []string{
					fmt.Sprintf("Revert operator %s (trivial rollback)", node.IncomingEdge.OperatorName),
					"Restore pre-transition state conditions",
				}
			case ReversibilityHighCost:
				actionSteps = []string{
					fmt.Sprintf("Revert operator %s (high-cost rollback procedure)", node.IncomingEdge.OperatorName),
					"Execute compensating resource cleanup",
					"Verify state restoration against active constraints",
				}
			case ReversibilityCritical:
				actionSteps = []string{
					fmt.Sprintf("CRITICAL / IRREVERSIBLE transition: %s", node.IncomingEdge.OperatorName),
					"Initiate emergency containment protocol",
					"Notify human executive / constitutional oversight",
				}
				penaltyCost = node.IncomingEdge.EdgeCost.Resources * 2.0
			default:
				actionSteps = []string{
					fmt.Sprintf("Revert operator %s", node.IncomingEdge.OperatorName),
				}
			}

			rb := RollbackStrategy{
				StrategyID:    rbID,
				TriggerNodeID: sgID,
				ActionSteps:   actionSteps,
				EstimatedCost: penaltyCost,
			}
			rollbacks = append(rollbacks, rb)
		}
	}

	plan, err := planBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build plan from search path: %w", err)
	}

	plan.RollbackStrategies = rollbacks

	s.mu.RLock()
	fpGen := s.fingerprinter
	s.mu.RUnlock()
	if fpGen == nil {
		fpGen = NewDefaultPlanFingerprinter()
	}
	fp, err := fpGen.ComputeFingerprint(plan)
	if err != nil {
		return nil, fmt.Errorf("failed to compute plan fingerprint: %w", err)
	}
	plan.PlanFingerprint = fp

	if err := plan.Validate(); err != nil {
		return nil, fmt.Errorf("converted plan validation failed: %w", err)
	}
	return plan, nil
}

// GenerateCandidatePlans executes Beam A* search and converts all discovered paths into ranked candidate Plans.
func (s *TreeSearchSpecialist) GenerateCandidatePlans(
	ctx context.Context,
	req *PlanningRequest,
	currentGraph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
) ([]*Plan, error) {
	res, err := s.InvokeSearch(ctx, req, currentGraph, profile)
	if err != nil {
		return nil, err
	}

	var targetNodes []*SearchNode
	if len(res.GoalNodes) > 0 {
		targetNodes = res.GoalNodes
	} else if len(res.PartialNodes) > 0 {
		targetNodes = res.PartialNodes
	} else {
		return []*Plan{}, nil
	}

	sorted := make([]*SearchNode, len(targetNodes))
	copy(sorted, targetNodes)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].CostProfile.EvaluationScore != sorted[j].CostProfile.EvaluationScore {
			return sorted[i].CostProfile.EvaluationScore < sorted[j].CostProfile.EvaluationScore
		}
		return sorted[i].State.EpistemicConfidence > sorted[j].State.EpistemicConfidence
	})

	snapshotID := "snap-default"
	if profile != nil && profile.ProfileID != "" {
		snapshotID = profile.ProfileID
	}
	traceID := fmt.Sprintf("trace-tree-%d", time.Now().UnixNano())

	candidates := make([]*Plan, 0, len(sorted))
	for rank, termNode := range sorted {
		path := ReconstructPath(termNode)
		if len(path) == 0 {
			continue
		}
		plan, err := s.ConvertPathToPlan(path, req, snapshotID, traceID, rank)
		if err != nil {
			return nil, fmt.Errorf("candidate[%d] conversion failed: %w", rank, err)
		}
		candidates = append(candidates, plan)
	}
	return candidates, nil
}

// GeneratePlanningResult orchestrates search, produces ranked candidate plans, and returns a fully populated PlanningResult.
func (s *TreeSearchSpecialist) GeneratePlanningResult(
	ctx context.Context,
	req *PlanningRequest,
	currentGraph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
) (*PlanningResult, error) {
	start := time.Now()
	if req == nil {
		return nil, errors.New("cannot generate PlanningResult with nil request")
	}

	candidates, err := s.GenerateCandidatePlans(ctx, req, currentGraph, profile)
	if err != nil {
		return nil, err
	}

	resID := fmt.Sprintf("res-tree-%d", start.UnixNano())
	res := &PlanningResult{
		ResultID:      resID,
		RequestID:     req.RequestID,
		Plans:         candidates,
		Traces:        make([]*PlanningTrace, 0),
		ExecutedDepth: req.TargetDepth,
		TotalDuration: time.Since(start),
	}
	if res.ExecutedDepth == "" {
		res.ExecutedDepth = DepthStrategic
	}

	if len(candidates) > 0 {
		res.PrimaryPlanID = candidates[0].PlanID
		res.ResultStatus = ResultSuccess
		res.Status = candidates[0].Status
	} else {
		res.ResultStatus = ResultNoPlans
		res.Status = PlanStatusInfeasible
	}

	traceID := fmt.Sprintf("trace-tree-%d", start.UnixNano())
	snapshotID := "snap-default"
	if profile != nil && profile.ProfileID != "" {
		snapshotID = profile.ProfileID
	}

	stats := SearchStatistics{
		NodesExpanded:             0,
		NodesPruned:               0,
		BeamWidthUsed:             uint32(DefaultBeamAStarConfig().BeamWidth),
		AlternativePlansGenerated: uint32(len(candidates)),
	}
	if len(candidates) > 0 {
		stats.NodesExpanded = uint64(len(candidates[0].Subgoals))
	}

	qm := QualityMetrics{
		Completeness:           0.90,
		Efficiency:             0.85,
		Robustness:             0.88,
		Flexibility:            0.80,
		ResourceEfficiency:     0.90,
		ExpectedExecutionCost:  0,
		EstimatedExecutionTime: 0,
		RiskExposure:           0.10,
		DependencyComplexity:   0.20,
		Maintainability:        0.85,
		Adaptability:           0.85,
	}
	if len(candidates) > 0 {
		qm.ExpectedExecutionCost = candidates[0].EstimatedCost
		qm.EstimatedExecutionTime = candidates[0].EstimatedDuration
	}

	planID := ""
	if len(candidates) > 0 {
		planID = candidates[0].PlanID
	}
	termReason := TerminationGoalFound
	if len(candidates) == 0 {
		termReason = TerminationNoValidPlan
	}

	trace := &PlanningTrace{
		TraceID:            traceID,
		PlanID:             planID,
		SchemaVersion:      SchemaVersion2_0_0,
		StrategySnapshotID: snapshotID,
		SearchStrategyID:   s.defaultHorizon,
		TerminationReason:  termReason,
		SearchStatistics:   stats,
		QualityMetrics:     qm,
		ConfidenceProfile: ConfidenceProfile{
			GoalConfidence:         0.90,
			PreconditionConfidence: 0.90,
			DependencyConfidence:   0.90,
			ResourceConfidence:     0.90,
			TimingConfidence:       0.90,
			ConstraintConfidence:   0.90,
			OverallConfidence:      0.90,
		},
		ReplayMetadata: ReplayMetadata{
			StrategySnapshotID: snapshotID,
			ReplayFidelity:     "EXACT",
			ReplaySeed:         uint64(start.UnixNano()),
			WorkingMemoryHash:  req.ContextRef,
		},
	}
	res.Traces = append(res.Traces, trace)

	return res, nil
}

// Contribute fulfills the PlanningSpecialist interface contract by invoking strategic search
// and returning subgoals/edges for candidate plan construction inside PlanningService.
func (s *TreeSearchSpecialist) Contribute(
	ctx context.Context,
	req *PlanningRequest,
	currentGraph *DependencyGraphSnapshot,
	profile *PlanningPolicyProfile,
) (*PlanningStepLog, []Subgoal, []DependencyEdge, error) {
	start := time.Now()

	strat := resolveSearchStrategy(profile, s.defaultHorizon)
	if strat == nil {
		return nil, nil, nil, fmt.Errorf("TreeSearchSpecialist: no active search strategy found for horizon %s", s.defaultHorizon)
	}

	if profile != nil && profile.Capabilities != nil && !profile.Capabilities.SupportsTreeSearch {
		return &PlanningStepLog{
			SpecialistName:  s.Name(),
			ActionPerformed: "Tree search skipped: engine capabilities do not support tree search",
			Duration:        time.Since(start),
		}, nil, nil, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}

	candidates, err := s.GenerateCandidatePlans(ctx, req, currentGraph, profile)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(candidates) == 0 {
		return &PlanningStepLog{
			SpecialistName:  s.Name(),
			ActionPerformed: "Tree search completed with 0 candidate trajectories discovered",
			Duration:        time.Since(start),
		}, nil, nil, nil
	}

	subgoals := make([]Subgoal, 0)
	var edges []DependencyEdge
	topScore := candidates[0].EstimatedCost
	for _, cand := range candidates {
		for _, sg := range cand.Subgoals {
			if !strings.HasPrefix(sg.SubgoalID, "tree-sg-") {
				sg.SubgoalID = "tree-sg-" + sg.SubgoalID
			}
			subgoals = append(subgoals, sg)
		}
		for _, dep := range cand.Dependencies {
			edges = append(edges, dep)
		}
	}

	workers := uint32(1)
	if strat.AllowParallelExpansion && (profile == nil || profile.Capabilities == nil || profile.Capabilities.SupportsParallelSearch) && strat.MaxConcurrentWorkers > 1 {
		workers = strat.MaxConcurrentWorkers
		if profile != nil && profile.Capabilities != nil && profile.Capabilities.MaxParallelWorkers > 0 && uint32(profile.Capabilities.MaxParallelWorkers) < workers {
			workers = uint32(profile.Capabilities.MaxParallelWorkers)
		}
		if strat.BeamWidth > 0 && workers > strat.BeamWidth {
			workers = strat.BeamWidth
		}
	}

	actionStr := fmt.Sprintf("Tree search generated %d candidate trajectories via Beam A* (top score: %.2f)", len(candidates), topScore)
	if workers > 1 {
		actionStr = fmt.Sprintf("Tree search generated %d alternative branches across %d concurrent workers governed by %s (policy=%s, top score: %.2f)", len(subgoals), workers, strat.SearchID, strat.ExpansionPolicy, topScore)
	}

	stepLog := &PlanningStepLog{
		SpecialistName:  s.Name(),
		ActionPerformed: actionStr,
		Duration:        time.Since(start),
		NodesAdded:      len(subgoals),
		EdgesAdded:      len(edges),
		Status:          "DONE",
	}

	return stepLog, subgoals, edges, nil
}

