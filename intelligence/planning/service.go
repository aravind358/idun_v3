package planning

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"idun/intelligence/executive"
)

// DefaultPlanningService implements PlanningService and executive.PlanningAbility.
// It orchestrates the Planning Episode Pipeline across Reflexive, Tactical, and Deliberative
// computational horizons while strictly preserving frozen responsibility boundaries:
// Planning is purely a plan constructor and diagnostic generator; it never interprets text (`Understanding`),
// derives logical proofs (`Reasoning`), selects actions (`Decision`), executes workflows (`Executive`),
// mutates strategy weights (`Learning`), or writes across episodes (`Memory`).
type DefaultPlanningService struct {
	mu            sync.RWMutex
	started       bool
	config        *Config
	strategyProv  StrategyProvider
	registry      PlanningSpecialistRegistry
	fingerprinter PlanFingerprinter
	publisher     WorkspacePublisher
	storer        PayloadStorer

	// traces holds bounded in-memory Ring-Buffer of PlanningTrace items keyed by TraceID
	traces    map[string]*PlanningTrace
	traceKeys []string

	seqCounter uint64
}

// Option configures DefaultPlanningService dependencies.
type Option func(*DefaultPlanningService)

// WithConfig overrides the default service configuration.
func WithConfig(cfg *Config) Option {
	return func(s *DefaultPlanningService) {
		if cfg != nil {
			s.config = cfg
		}
	}
}

// WithStrategyProvider overrides the default StrategyProvider.
func WithStrategyProvider(prov StrategyProvider) Option {
	return func(s *DefaultPlanningService) {
		s.strategyProv = prov
	}
}

// WithSpecialistRegistry overrides the default PlanningSpecialistRegistry.
func WithSpecialistRegistry(reg PlanningSpecialistRegistry) Option {
	return func(s *DefaultPlanningService) {
		s.registry = reg
	}
}

// WithFingerprinter overrides the default PlanFingerprinter.
func WithFingerprinter(fp PlanFingerprinter) Option {
	return func(s *DefaultPlanningService) {
		s.fingerprinter = fp
	}
}

// WithWorkspaceBridge configures optional Global Workspace publication dependencies.
func WithWorkspaceBridge(storer PayloadStorer, publisher WorkspacePublisher) Option {
	return func(s *DefaultPlanningService) {
		s.storer = storer
		s.publisher = publisher
	}
}

// NewService constructs a DefaultPlanningService with safe production defaults.
func NewService(opts ...Option) *DefaultPlanningService {
	cfg := DefaultConfig()
	s := &DefaultPlanningService{
		config:        cfg,
		strategyProv:  NewDefaultStrategyProvider(nil),
		registry:      NewSpecialistRegistry(),
		fingerprinter: NewDefaultPlanFingerprinter(),
		traces:        make(map[string]*PlanningTrace, cfg.MaxTraceRetention),
		traceKeys:     make([]string, 0, cfg.MaxTraceRetention),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Ability returns executive.AbilityPlanning.
func (s *DefaultPlanningService) Ability() executive.CognitiveAbility {
	return executive.AbilityPlanning
}

// Start boots the Planning service lifecycle.
func (s *DefaultPlanningService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	if err := s.config.Validate(); err != nil {
		return fmt.Errorf("planning service start failed config validation: %w", err)
	}
	s.started = true
	return nil
}

// Close gracefully shuts down the Planning service.
func (s *DefaultPlanningService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = false
	return nil
}

// GetTrace retrieves an O(1) memory-retained PlanningTrace by its unique TraceID.
func (s *DefaultPlanningService) GetTrace(traceID string) (*PlanningTrace, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.traces[traceID]
	return t, ok
}

// Capabilities returns the immutable structural boundaries advertised by this planning engine deployment.
func (s *DefaultPlanningService) Capabilities() *PlanningCapabilities {
	// If the active profile explicitly provides capabilities, return those; otherwise fallback to service config capabilities.
	if s.strategyProv != nil {
		if snap := s.strategyProv.ActiveSnapshot(); snap != nil {
			if prof := snap.ActiveProfile(); prof != nil && prof.Capabilities != nil {
				return prof.Capabilities
			}
		}
	}
	if s.config != nil && s.config.Capabilities != nil {
		return s.config.Capabilities
	}
	return DefaultPlanningCapabilities()
}

// storeTrace safely retains a diagnostic trace inside the bounded ring buffer.
func (s *DefaultPlanningService) storeTrace(trace *PlanningTrace) {
	if trace == nil || trace.TraceID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.traces[trace.TraceID]; !exists {
		if len(s.traceKeys) >= s.config.MaxTraceRetention {
			oldest := s.traceKeys[0]
			s.traceKeys = s.traceKeys[1:]
			delete(s.traces, oldest)
		}
		s.traceKeys = append(s.traceKeys, trace.TraceID)
	}
	s.traces[trace.TraceID] = trace
}

// PlanReflexive executes fast-path cache/memoized template planning (<10ms budget).
func (s *DefaultPlanningService) PlanReflexive(ctx context.Context, req *PlanningRequest) (*PlanningResult, error) {
	return s.executePlanningEpisode(ctx, req, DepthReflexive)
}

// PlanTactical executes domain-weighted specialist HTN/GOAP decomposition (10-100ms budget).
func (s *DefaultPlanningService) PlanTactical(ctx context.Context, req *PlanningRequest) (*PlanningResult, error) {
	return s.executePlanningEpisode(ctx, req, DepthTactical)
}

// PlanDeliberative executes wide multi-alternative tree search and contingency planning (100-500ms budget).
func (s *DefaultPlanningService) PlanDeliberative(ctx context.Context, req *PlanningRequest) (*PlanningResult, error) {
	return s.executePlanningEpisode(ctx, req, DepthStrategic)
}

// DecomposeGoal implements executive.PlanningAbility.DecomposeGoal.
func (s *DefaultPlanningService) DecomposeGoal(ctx context.Context, goalRef string) (string, error) {
	if goalRef == "" {
		return "", errors.New("planning: DecomposeGoal called with empty goalRef")
	}

	req, err := NewPlanningRequestBuilder().
		WithRequestID(fmt.Sprintf("req-goal-%d", time.Now().UnixNano())).
		WithGoal(goalRef).
		WithDomain("General").
		WithTargetDepth(DepthTactical).
		Build()
	if err != nil {
		return "", err
	}

	res, err := s.PlanTactical(ctx, req)
	if err != nil {
		return "", err
	}

	if s.storer != nil && s.publisher != nil {
		env, err := PublishPlanningResult(ctx, res, s.storer, s.publisher)
		if err != nil {
			return "", err
		}
		return env.PayloadRef, nil
	}

	return res.PrimaryPlanID, nil
}

// ExecuteTask implements executive.AbilityDriver.ExecuteTask.
func (s *DefaultPlanningService) ExecuteTask(
	ctx context.Context,
	workflowID, nodeID string,
	budget executive.BudgetTier,
	payloadRef string,
) (executive.EpistemicStatus, string, error) {
	req, err := NewPlanningRequestBuilder().
		WithRequestID(fmt.Sprintf("req-task-%s-%s", workflowID, nodeID)).
		WithGoal(fmt.Sprintf("Execute planning workflow step for node %s", nodeID)).
		WithContextRef(payloadRef).
		WithTargetDepth(DepthTactical).
		Build()
	if err != nil {
		return executive.StatusEscalationRequired, "", err
	}

	var res *PlanningResult
	switch budget {
	case executive.BudgetReflexive:
		res, err = s.PlanReflexive(ctx, req)
	case executive.BudgetDeliberative:
		res, err = s.PlanDeliberative(ctx, req)
	default:
		res, err = s.PlanTactical(ctx, req)
	}

	if err != nil {
		return executive.StatusEscalationRequired, "", err
	}

	if res.ResultStatus == ResultEscalationRecommended {
		return executive.StatusEscalationRequired, res.PrimaryPlanID, nil
	}
	if res.ResultStatus == ResultNoPlans || res.Status == PlanStatusInfeasible {
		return executive.StatusUnsureConflicting, res.PrimaryPlanID, nil
	}
	if res.Status == PlanStatusInsufficientInfo {
		return executive.StatusInsufficientData, res.PrimaryPlanID, nil
	}

	return executive.StatusConfident, res.PrimaryPlanID, nil
}

// executePlanningEpisode orchestrates the canonical 8-Stage Planning Episode Pipeline.
func (s *DefaultPlanningService) executePlanningEpisode(
	ctx context.Context,
	req *PlanningRequest,
	depth PlanningDepth,
) (*PlanningResult, error) {
	start := time.Now()

	// Stage 0: Budget & Scope Validation Firewall
	s.mu.RLock()
	if !s.started {
		s.mu.RUnlock()
		return nil, errors.New("planning service not started")
	}
	s.mu.RUnlock()

	if req == nil {
		return nil, errors.New("cannot execute planning episode with nil request")
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("firewall stage 0 rejected request: %w", err)
	}

	// Acquire Strategy Snapshot & Active Profile
	snapshot := s.strategyProv.ActiveSnapshot()
	if snapshot == nil || snapshot.ActiveProfile() == nil {
		return nil, errors.New("active PlanningStrategySnapshot or profile is nil")
	}
	profile := snapshot.ActiveProfile()

	// Timeout Propagation & Isolation
	maxDuration := profile.MaxPlanningTime
	if req.MaxExecutionBudget > 0 && req.MaxExecutionBudget < maxDuration {
		maxDuration = req.MaxExecutionBudget
	}
	if maxDuration <= 0 {
		maxDuration = 250 * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, maxDuration)
	defer cancel()

	// Stage 1: Create ReflexivePlanningCache (destroyed immediately upon completion)
	cache := NewReflexivePlanningCache(fmt.Sprintf("ep-%s", req.RequestID), snapshot.Version)
	defer cache.Close()

	// Stage 2/3: Dispatch Specialists
	graph := &DependencyGraphSnapshot{
		Nodes: make(map[string]string),
	}
	stepLogs, subgoals, edges, specErr := s.registry.ExecuteSpecialists(ctx, req, graph, profile, cache)

	// Determine Termination Reason and Status
	termReason := TerminationGoalFound
	planStatus := PlanStatusComplete
	resultStatus := ResultSuccess

	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(specErr, context.DeadlineExceeded) {
		termReason = TerminationTimeLimit
		if len(subgoals) > 0 {
			planStatus = PlanStatusPartialBudgetExhausted
			resultStatus = ResultPartialPlans
		} else {
			planStatus = PlanStatusInfeasible
			resultStatus = ResultNoPlans
		}
	} else if errors.Is(ctx.Err(), context.Canceled) || errors.Is(specErr, context.Canceled) {
		termReason = TerminationExecutiveCancelled
		planStatus = PlanStatusInfeasible
		resultStatus = ResultNoPlans
	} else if specErr != nil {
		termReason = TerminationNoValidPlan
		planStatus = PlanStatusInfeasible
		resultStatus = ResultValidationFailed
	} else if len(subgoals) == 0 {
		termReason = TerminationNoValidPlan
		planStatus = PlanStatusInfeasible
		resultStatus = ResultNoPlans
	}

	// Stage 4: Assemble Plan estimates and ConfidenceProfile
	estCost := float64(len(subgoals)) * 2.5
	estDuration := time.Duration(len(subgoals)) * 10 * time.Minute
	if len(subgoals) == 0 {
		estCost = 0
		estDuration = 0
	}

	baseConf := 0.90
	if planStatus != PlanStatusComplete {
		baseConf = 0.40
	}
	cp := ConfidenceProfile{
		GoalConfidence:         baseConf,
		PreconditionConfidence: baseConf,
		DependencyConfidence:   baseConf,
		ResourceConfidence:     baseConf,
		TimingConfidence:       baseConf,
		ConstraintConfidence:   baseConf,
		OverallConfidence:      baseConf,
	}

	var escalation EscalationAction = ActionNone
	if len(subgoals) > 0 && (cp.OverallConfidence < req.MinConfidenceFloor || cp.OverallConfidence < profile.EscalationThresholds["ConfidenceFloor"]) {
		if depth == DepthReflexive {
			escalation = ActionRecommendHigherDepth
			resultStatus = ResultEscalationRecommended
		} else if depth == DepthTactical {
			escalation = ActionRecommendHigherDepth
			resultStatus = ResultEscalationRecommended
		} else {
			escalation = ActionRecommendMorePlanning
			resultStatus = ResultEscalationRecommended
		}
	}

	seq := atomic.AddUint64(&s.seqCounter, 1)
	planID := fmt.Sprintf("plan-%d-%d", time.Now().UnixNano(), seq)
	traceID := fmt.Sprintf("trace-%d-%d", time.Now().UnixNano(), seq)

	planBuilder := NewPlanBuilder().
		WithIdentity(planID, snapshot.SnapshotID, traceID).
		WithGoalAndDomain(req.Goal, req.Domain, string(depth)).
		WithEstimates(estCost, estDuration, nil).
		WithConfidenceProfile(cp).
		WithStatus(planStatus, nil).
		WithReplayMetadata(ReplayMetadata{
			StrategySnapshotID: snapshot.SnapshotID,
			ReplayFidelity:     "EXACT",
			ReplaySeed:         uint64(start.UnixNano()),
			WorkingMemoryHash:  req.ContextRef,
		})

	for _, sg := range subgoals {
		planBuilder.AddSubgoal(sg)
		graph.Nodes[sg.SubgoalID] = sg.Title
	}
	for _, edge := range edges {
		planBuilder.AddDependency(edge)
	}
	graph.Edges = edges

	plan, err := planBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("stage 4 plan build failed: %w", err)
	}

	// Stage 6: Compute PlanFingerprint
	fp, err := s.fingerprinter.ComputeFingerprint(plan)
	if err != nil {
		return nil, fmt.Errorf("stage 6 fingerprint computation failed: %w", err)
	}
	plan.PlanFingerprint = fp

	// Stage 5: Assemble PlanningTrace
	stats := cache.Summary()
	qm := QualityMetrics{
		Completeness:           baseConf,
		Efficiency:             0.85,
		Robustness:             0.88,
		Flexibility:            0.80,
		ResourceEfficiency:     0.90,
		ExpectedExecutionCost:  estCost,
		EstimatedExecutionTime: estDuration,
		RiskExposure:           0.15,
		DependencyComplexity:   0.20,
		Maintainability:        0.90,
		Adaptability:           0.85,
	}

	policyFP := ""
	if profile != nil {
		policyFP = profile.PolicyFingerprint
	}
	caps := s.Capabilities()
	capFP := ""
	if caps != nil {
		capFP = caps.CapabilityFingerprint
		if capFP == "" {
			capFP = ComputeCapabilityFingerprint(caps)
		}
	}
	horizon := "TACTICAL"
	if depth == DepthReflexive {
		horizon = "REFLEXIVE"
	} else if depth == DepthStrategic {
		horizon = "STRATEGIC"
	}
	strat := resolveSearchStrategy(profile, horizon)
	searchStrategyID := ""
	if strat != nil {
		searchStrategyID = strat.SearchID
	}

	traceBuilder := NewPlanningTraceBuilder().
		WithIdentity(traceID, planID, snapshot.SnapshotID).
		WithDiagnostics(termReason, stats, float64(len(subgoals))*1.5, cp, qm).
		WithProvenance(policyFP, capFP, searchStrategyID, plan.ReplayMetadata)

	for _, st := range stepLogs {
		traceBuilder.AddStepLog(st)
	}

	stepLogMap := make(map[string]PlanningStepLog, len(stepLogs))
	for _, st := range stepLogs {
		stepLogMap[st.SpecialistName] = st
	}
	var allSpecs []PlanningSpecialist
	if reg, ok := s.registry.(*DefaultSpecialistRegistry); ok {
		allSpecs = reg.GetAllSpecialists()
	} else {
		allSpecs = s.registry.GetSpecialistsForDomain(req.Domain, profile)
	}
	for _, spec := range allSpecs {
		st, invoked := stepLogMap[spec.Name()]
		usage := PlanningSpecialistUsage{
			SpecialistID: spec.Name(),
			Invoked:      invoked,
		}
		if invoked {
			usage.SkipReason = SkipNone
			usage.NodesExpanded = uint64(st.NodesAdded)
			usage.ExecutionTimeUs = uint64(st.Duration.Microseconds())
			if usage.ExecutionTimeUs == 0 {
				usage.ExecutionTimeUs = 1
			}
			usage.PlansGenerated = uint32(st.NodesAdded)
			usage.Success = (st.Status == "DONE")
			if len(subgoals) > 0 {
				usage.ContributionScore = float32(st.NodesAdded) / float32(len(subgoals))
				if usage.ContributionScore > 1.0 {
					usage.ContributionScore = 1.0
				}
			}
		} else {
			// Determine factual SkipReason
			domainSupported := false
			for _, d := range spec.SupportedDomains() {
				if d == req.Domain || d == "General" || req.Domain == "" {
					domainSupported = true
					break
				}
			}
			if !domainSupported {
				usage.SkipReason = SkipDomainMismatch
			} else if profile != nil && profile.Capabilities != nil {
				name := spec.Name()
				if (name == "HTNDecompositionSpecialist" || name == "HTNSpecialist") && !profile.Capabilities.SupportsHTN {
					usage.SkipReason = SkipCapabilityDisabled
				} else if (name == "GOAPActionSpecialist" || name == "GOAPSpecialist") && !profile.Capabilities.SupportsGOAP {
					usage.SkipReason = SkipCapabilityDisabled
				} else if (name == "MultiAlternativeTreeSearchSpecialist" || name == "TreeSearchSpecialist") && !profile.Capabilities.SupportsTreeSearch {
					usage.SkipReason = SkipCapabilityDisabled
				} else if ctx.Err() != nil {
					usage.SkipReason = SkipCancelled
				} else if len(subgoals) == 0 {
					usage.SkipReason = SkipNoApplicableGoal
				} else {
					usage.SkipReason = SkipHigherPrioritySpecialist
				}
			} else {
				if ctx.Err() != nil {
					usage.SkipReason = SkipCancelled
				} else if len(subgoals) == 0 {
					usage.SkipReason = SkipNoApplicableGoal
				} else {
					usage.SkipReason = SkipHigherPrioritySpecialist
				}
			}
		}
		traceBuilder.AddSpecialistUsage(usage)
	}

	trace, err := traceBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("stage 5 trace build failed: %w", err)
	}
	trace.DecompositionTree = DecompositionNode{
		NodeID:   planID,
		Title:    req.Goal,
		NodeType: "GOAL",
	}
	trace.DependencyGraph = *graph

	// Stage 8: Validation Firewall & Storage
	if err := plan.Validate(); err != nil {
		return nil, fmt.Errorf("firewall stage 8 plan validation failed: %w", err)
	}
	if err := trace.Validate(); err != nil {
		return nil, fmt.Errorf("firewall stage 8 trace validation failed: %w", err)
	}

	s.storeTrace(trace)

	res := &PlanningResult{
		ResultID:                 fmt.Sprintf("res-%d", time.Now().UnixNano()),
		RequestID:                req.RequestID,
		Plans:                    []*Plan{plan},
		Traces:                   []*PlanningTrace{trace},
		PrimaryPlanID:            plan.PlanID,
		ResultStatus:             resultStatus,
		Status:                   planStatus,
		EscalationRecommendation: escalation,
		ExecutedDepth:            depth,
		TotalDuration:            time.Since(start),
	}

	if err := res.Validate(); err != nil {
		return nil, fmt.Errorf("firewall stage 8 result validation failed: %w", err)
	}

	// Publish to Global Workspace if configured and depth warrants broadcast
	if s.publisher != nil && s.storer != nil && ShouldPublishToWorkspace(depth) {
		_, _ = PublishPlan(ctx, plan, s.storer, s.publisher)
		_, _ = PublishPlanningTrace(ctx, trace, s.storer, s.publisher)
		_, _ = PublishPlanningResult(ctx, res, s.storer, s.publisher)
	}

	return res, nil
}
