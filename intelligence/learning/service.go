package learning

import (
	"context"
	"fmt"
	"sync"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

// Service implements LearningService, executive.LearningAbility, and executive.AbilityDriver.
// It orchestrates the asynchronous, offline Phase 1 foundational learning checks while
// enforcing strict responsibility boundaries and validation firewalls.
type Service struct {
	mu           sync.RWMutex
	started      bool
	config       *Config
	strategyProv StrategyProvider
	learnerReg   *LearnerRegistry
	snapshotReg  SnapshotRegistry
	aggregator   ExperienceAggregator
	validation      ValidationPipeline
	experiment      ExperimentManager
	rolloutExecutor RolloutExecutor
	governance      GovernanceBridge
	workspace       workspace.Workspace
	ranking         CandidateRankingEngine

	traces             map[string]*LearningTrace
	learnerPerformance map[string]*LearnerPerformanceSummary
}

// ServiceOption configures dependencies for Service at initialization.
type ServiceOption func(*Service)

// WithConfig sets a custom service configuration.
func WithConfig(cfg *Config) ServiceOption {
	return func(s *Service) {
		if cfg != nil {
			s.config = cfg
		}
	}
}

// WithStrategyProvider overrides the default strategy provider.
func WithStrategyProvider(prov StrategyProvider) ServiceOption {
	return func(s *Service) {
		if prov != nil {
			s.strategyProv = prov
		}
	}
}

// WithSnapshotRegistry overrides the default snapshot registry.
func WithSnapshotRegistry(reg SnapshotRegistry) ServiceOption {
	return func(s *Service) {
		if reg != nil {
			s.snapshotReg = reg
		}
	}
}

// WithAggregator sets the experience aggregator for the service.
func WithAggregator(agg ExperienceAggregator) ServiceOption {
	return func(s *Service) {
		if agg != nil {
			s.aggregator = agg
		}
	}
}

// WithValidationPipeline sets the candidate validation pipeline.
func WithValidationPipeline(val ValidationPipeline) ServiceOption {
	return func(s *Service) {
		if val != nil {
			s.validation = val
		}
	}
}

// WithExperimentManager sets the bounded experiment manager.
func WithExperimentManager(exp ExperimentManager) ServiceOption {
	return func(s *Service) {
		if exp != nil {
			s.experiment = exp
		}
	}
}

// WithRolloutExecutor sets the rollout executor interface for experiment coordination.
func WithRolloutExecutor(executor RolloutExecutor) ServiceOption {
	return func(s *Service) {
		if executor != nil {
			s.rolloutExecutor = executor
		}
	}
}

// WithGovernanceBridge sets the governance bridge for diagnostic publishing.
func WithGovernanceBridge(bridge GovernanceBridge) ServiceOption {
	return func(s *Service) {
		if bridge != nil {
			s.governance = bridge
		}
	}
}

// WithWorkspace sets the global workspace for publishing immutable learning results.
func WithWorkspace(ws workspace.Workspace) ServiceOption {
	return func(s *Service) {
		if ws != nil {
			s.workspace = ws
		}
	}
}

// WithCandidateRankingEngine sets a custom candidate ranking engine.
func WithCandidateRankingEngine(ranking CandidateRankingEngine) ServiceOption {
	return func(s *Service) {
		if ranking != nil {
			s.ranking = ranking
		}
	}
}

// WithLearnerRegistry sets a custom learner registry.
func WithLearnerRegistry(reg *LearnerRegistry) ServiceOption {
	return func(s *Service) {
		if reg != nil {
			s.learnerReg = reg
		}
	}
}

// NewService constructs a new Learning Service coordinator.
func NewService(opts ...ServiceOption) (*Service, error) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("default config validation failed: %w", err)
	}

	sp, err := NewDefaultStrategyProvider(nil)
	if err != nil {
		return nil, fmt.Errorf("default strategy provider initialization failed: %w", err)
	}

	s := &Service{
		config:             cfg,
		strategyProv:       sp,
		learnerReg:         NewLearnerRegistry(),
		snapshotReg:        NewDefaultSnapshotRegistry(),
		ranking:            NewDefaultCandidateRankingEngine(),
		traces:             make(map[string]*LearningTrace),
		learnerPerformance: make(map[string]*LearnerPerformanceSummary),
	}

	for _, opt := range opts {
		opt(s)
	}

	if err := s.config.Validate(); err != nil {
		return nil, fmt.Errorf("configured service validation failed: %w", err)
	}

	return s, nil
}

// Ability returns executive.AbilityLearning.
func (s *Service) Ability() executive.CognitiveAbility {
	return executive.AbilityLearning
}

// Start boots the Learning Service lifecycle.
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	if err := s.config.Validate(); err != nil {
		return fmt.Errorf("learning service start failed validation: %w", err)
	}
	s.started = true
	return nil
}

// Close cleanly shuts down the Learning Service coordinator.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = false
	return nil
}

// GetLearnerPerformanceSummary retrieves the bounded performance summary for a learner.
func (s *Service) GetLearnerPerformanceSummary(ctx context.Context, learnerID string) (*LearnerPerformanceSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summary, exists := s.learnerPerformance[learnerID]
	if !exists || summary == nil {
		return nil, fmt.Errorf("%w: learner performance summary not found for %q", ErrNotFound, learnerID)
	}
	cp := *summary
	return &cp, nil
}

// ListLearnerPerformanceSummaries returns all accumulated learner performance summaries.
func (s *Service) ListLearnerPerformanceSummaries(ctx context.Context) []*LearnerPerformanceSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]*LearnerPerformanceSummary, 0, len(s.learnerPerformance))
	for _, lps := range s.learnerPerformance {
		if lps != nil {
			cp := *lps
			res = append(res, &cp)
		}
	}
	return res
}

// RegisterLearner registers a signature-based Learner in the open LearnerRegistry.
func (s *Service) RegisterLearner(learner Learner) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.started {
		return ErrServiceClosed
	}
	return s.learnerReg.Register(learner)
}

// GetActiveSnapshot returns the currently active CandidateSnapshot for a domain schema.
func (s *Service) GetActiveSnapshot(ctx context.Context, schemaID string) (*CandidateSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.started {
		return nil, ErrServiceClosed
	}
	return s.snapshotReg.GetActive(ctx, schemaID)
}

// RunCycle executes an offline, windowed Phase 1 learning check against aggregated experiences.
func (s *Service) RunCycle(ctx context.Context, req *LearningRequest) (*LearningResult, error) {
	start := time.Now()

	s.mu.RLock()
	if !s.started {
		s.mu.RUnlock()
		return nil, ErrServiceClosed
	}
	s.mu.RUnlock()

	if req == nil {
		return nil, fmt.Errorf("%w: nil request", ErrValidationFailed)
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	snapshot := s.strategyProv.ActiveSnapshot()
	if snapshot == nil || snapshot.ActiveProfile == nil || snapshot.Capabilities == nil {
		return nil, fmt.Errorf("%w: missing active strategy snapshot profiles", ErrValidationFailed)
	}

	// Verify explicit capability fail-fast invariant:
	// If the active profile or capability declaration does not support required capabilities, abstain immediately.
	caps := snapshot.Capabilities
	if !caps.SupportsOfflineLearning {
		res := &LearningResult{
			ResultID:          fmt.Sprintf("res-abstain-%d", time.Now().UnixNano()),
			RequestID:         req.RequestID,
			Status:            StatusAbstained,
			TerminationReason: ReasonCapabilityUnavailable,
			Candidates:        []*CandidateSnapshot{},
			Traces:            []*LearningTrace{},
			TotalDuration:     time.Since(start),
		}
		return res, nil
	}

	// Lookup candidate learners
	learners := s.learnerReg.LookupByProduces(req.DomainSchemaID)
	if len(learners) == 0 {
		// If no specific learners produce this schema, fall back to listing all
		learners = s.learnerReg.ListLearners()
	}

	var summary *AggregationSummary
	var lineage ReplayMetadata
	if s.aggregator != nil {
		var errAgg error
		summary, lineage, errAgg = s.aggregator.AggregateWindow(ctx, req, snapshot)
		if errAgg != nil {
			return nil, fmt.Errorf("experience aggregation failed: %w", errAgg)
		}
	} else {
		summary = &AggregationSummary{
			SummaryID:              fmt.Sprintf("sum-%d", time.Now().UnixNano()),
			TimeWindowStart:        req.TimeWindowStart,
			TimeWindowEnd:          req.TimeWindowEnd,
			TotalArtifactsIngested: 0,
			SourceArtifactHash:     "sha256-phase1-placeholder",
			DomainSchemaIDs:        []string{req.DomainSchemaID},
		}
		lineage = ReplayMetadata{
			LearningFingerprint: "fp-learn",
			PolicyFingerprint:   req.PolicyFingerprint,
			SourceArtifactHash:  summary.SourceArtifactHash,
		}
	}
	_ = lineage

	var acceptedCandidates []*CandidateSnapshot
	var usages []LearnerUsage
	validationFailedCount := 0
	sampleFloorFailed := false
	var failedCheckIDs []string

	for _, l := range learners {
		if ctx.Err() != nil {
			break
		}
		luStart := time.Now()
		cands, err := l.Generate(ctx, summary)
		if err != nil {
			usages = append(usages, LearnerUsage{
				LearnerID:          l.LearnerID(),
				DomainSchemaID:     req.DomainSchemaID,
				Invoked:            true,
				Skipped:            false,
				SkipReason:         err.Error(),
				CandidatesProduced: 0,
				CandidatesAccepted: 0,
				ExecutionTime:      time.Since(luStart),
				ContributionScore:  0.0,
			})
			continue
		}

		var acceptedForLearner int
		for _, cand := range cands {
			// Phase 3: Candidate Lineage Integration
			if s.snapshotReg != nil {
				parent, errVal := s.snapshotReg.GetActive(ctx, cand.SchemaID)
				if errVal == nil && parent != nil && parent.SnapshotID != "" {
					anc := parent.Lineage.AncestorSnapshot
					if anc == "" {
						anc = parent.SnapshotID
					}
					depth := parent.Lineage.GenerationDepth + 1
					cand.Provenance = &CandidateLineage{
						ParentSnapshot:   parent.SnapshotID,
						AncestorSnapshot: anc,
						GenerationDepth:  depth,
					}
					cand.Lineage.ParentSnapshot = parent.SnapshotID
					cand.Lineage.AncestorSnapshot = anc
					cand.Lineage.GenerationDepth = depth
				} else if cand.Provenance == nil {
					cand.Provenance = &CandidateLineage{
						ParentSnapshot:   "",
						AncestorSnapshot: cand.SnapshotID,
						GenerationDepth:  0,
					}
					cand.Lineage.AncestorSnapshot = cand.SnapshotID
					cand.Lineage.GenerationDepth = 0
				}
			} else if cand.Provenance == nil {
				cand.Provenance = &CandidateLineage{
					ParentSnapshot:   "",
					AncestorSnapshot: cand.SnapshotID,
					GenerationDepth:  0,
				}
				cand.Lineage.AncestorSnapshot = cand.SnapshotID
				cand.Lineage.GenerationDepth = 0
			}
			if cand.Lineage.LearnerFingerprint == "" {
				cand.Lineage.LearnerFingerprint = l.LearnerFingerprint()
			}

			if s.validation != nil {
				valRecords, structRes, errVal := s.validation.ValidateCandidate(ctx, cand, summary, snapshot.ActiveProfile)
				cand.ValidationRecords = valRecords
				cand.StructuralValidation = structRes
				if errVal != nil {
					cand.ValidationHash = "hash-val-error"
					validationFailedCount++
					failedCheckIDs = append(failedCheckIDs, "SCHEMA_PAYLOAD_CHECK")
					continue
				}
				allPassed := true
				for _, vr := range valRecords {
					if !vr.Passed {
						allPassed = false
						failedCheckIDs = append(failedCheckIDs, vr.CheckID)
						if vr.CheckID == "STAT_SAMPLE_FLOOR" {
							sampleFloorFailed = true
						}
					}
				}
				if structRes != nil && !structRes.Passed {
					allPassed = false
					failedCheckIDs = append(failedCheckIDs, "STRUCTURAL_CHECK")
				}
				if !allPassed {
					cand.ValidationHash = "hash-val-failed"
					validationFailedCount++
					continue
				}
				cand.ValidationHash = fmt.Sprintf("hash-val-passed-%d", time.Now().UnixNano())
				// Enforce Phase 2 promotion boundary: strictly transition Draft -> Validated
				if cand.Lifecycle == LifecycleDraft {
					cand.Lifecycle = LifecycleValidated
				}
			}

			if s.snapshotReg != nil && cand.Lifecycle == LifecycleValidated {
				_ = s.snapshotReg.Publish(ctx, cand)
			}
			if s.rolloutExecutor != nil && (cand.Lifecycle == LifecycleValidated || cand.Lifecycle == LifecycleShadow || cand.Lifecycle == LifecycleCanary) {
				_ = s.rolloutExecutor.PromoteCandidate(ctx, cand.SnapshotID, cand.Lifecycle)
			}

			acceptedCandidates = append(acceptedCandidates, cand)
			acceptedForLearner++
		}

		contrib := 1.0
		if len(cands) > 0 {
			contrib = float64(acceptedForLearner) / float64(len(cands))
		}

		usages = append(usages, LearnerUsage{
			LearnerID:          l.LearnerID(),
			DomainSchemaID:     req.DomainSchemaID,
			Invoked:            true,
			Skipped:            false,
			CandidatesProduced: len(cands),
			CandidatesAccepted: acceptedForLearner,
			ExecutionTime:      time.Since(luStart),
			ContributionScore:  contrib,
		})
	}

	if s.ranking != nil && len(acceptedCandidates) > 1 {
		if ranked, errRank := s.ranking.RankCandidates(ctx, acceptedCandidates, summary, snapshot.ActiveProfile); errRank == nil {
			acceptedCandidates = ranked
		}
	}

	status := StatusPublished
	reason := ReasonSuccess
	if len(acceptedCandidates) == 0 {
		if validationFailedCount > 0 {
			status = StatusValidationFail
			reason = ReasonDriftAnomalyDetected
			if sampleFloorFailed {
				reason = ReasonSampleFloorNotMet
			}
		} else {
			status = StatusNoCandidates
			reason = ReasonNoCandidates
		}
	}

	lineageMeta := ReplayMetadata{
		LearningFingerprint: LearningFingerprint(fmt.Sprintf("learning-service-%d", time.Now().UnixNano())),
		PolicyFingerprint:   snapshot.ActiveProfile.PolicyFingerprint,
		SourceArtifactHash:  summary.SourceArtifactHash,
		ReplaySeed:          uint64(time.Now().UnixNano()),
	}
	if len(acceptedCandidates) > 0 {
		lineageMeta.LearnerFingerprint = acceptedCandidates[0].Lineage.LearnerFingerprint
		lineageMeta.ParentSnapshot = acceptedCandidates[0].Lineage.ParentSnapshot
		lineageMeta.AncestorSnapshot = acceptedCandidates[0].Lineage.AncestorSnapshot
		lineageMeta.GenerationDepth = acceptedCandidates[0].Lineage.GenerationDepth
	} else if len(learners) > 0 {
		lineageMeta.LearnerFingerprint = learners[0].LearnerFingerprint()
	}

	rejectionSummary := ComputeCandidateRejectionSummary(failedCheckIDs)

	s.mu.Lock()
	var learnerPerfs []*LearnerPerformanceSummary
	for _, u := range usages {
		existing := s.learnerPerformance[u.LearnerID]
		var ver, fp string
		for _, l := range learners {
			if l.LearnerID() == u.LearnerID {
				ver = l.LearnerVersion()
				fp = l.LearnerFingerprint()
				break
			}
		}
		updated := UpdateLearnerPerformanceSummary(existing, u, ver, fp)
		s.learnerPerformance[u.LearnerID] = updated
		learnerPerfs = append(learnerPerfs, updated)
	}
	s.mu.Unlock()

	traceBuilder := NewLearningTraceBuilder(
		fmt.Sprintf("trace-%d", time.Now().UnixNano()),
		req.RequestID,
		req.DomainSchemaID,
		snapshot.ActiveProfile.PolicyFingerprint,
	).
		WithLineage(lineageMeta).
		WithAggregation(*summary).
		WithCandidateCount(len(acceptedCandidates)).
		WithStatus(status).
		WithTerminationReason(reason).
		WithTotalDuration(time.Since(start)).
		WithStatisticalSummary(ComputeTraceStatisticalSummaryWithRejections(summary, acceptedCandidates, rejectionSummary)).
		WithLearnerPerformance(learnerPerfs)
	if req.CampaignID != "" {
		traceBuilder = traceBuilder.WithCampaignID(req.CampaignID)
	}
	trace, err := traceBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("trace build failed: %w", err)
	}

	for _, u := range usages {
		trace.LearnerUsages = append(trace.LearnerUsages, u)
	}

	s.mu.Lock()
	s.traces[trace.TraceID] = trace
	s.mu.Unlock()

	// Phase 3: Governance Bridge Integration
	// Learning publishes LearningDiagnostics (trace ID reference). GovernanceBridge produces HealthRecommendations.
	// Learning never interprets its own diagnostics. Executive consumes only HealthRecommendations.
	if s.governance != nil {
		_, _ = s.governance.EvaluateDiagnostics(ctx, trace.TraceID)
	}

	res := &LearningResult{
		ResultID:          fmt.Sprintf("res-%d", time.Now().UnixNano()),
		RequestID:         req.RequestID,
		CampaignID:        req.CampaignID,
		Status:            status,
		TerminationReason: reason,
		Candidates:        acceptedCandidates,
		Traces:            []*LearningTrace{trace},
		TotalDuration:     time.Since(start),
	}

	if err := res.Validate(); err != nil {
		return nil, fmt.Errorf("result validation failed: %w", err)
	}

	// Phase 3: Workspace Publishing
	// Publish validated candidate summaries and immutable LearningResults to the Global Workspace.
	if s.workspace != nil && res.Status == StatusPublished {
		env, errEnv := communication.NewEnvelopeBuilder().
			WithID(fmt.Sprintf("env-res-%d", time.Now().UnixNano())).
			WithSource("idun.intelligence.learning").
			WithTopic(communication.TopicReflections).
			WithPayloadRef(res.ResultID).
			WithConfidence(1.0).
			WithUrgency(50).
			Build()
		if errEnv == nil {
			_ = s.workspace.Publish(ctx, env)
		}
	}

	return res, nil
}

// ExecuteTask implements executive.AbilityDriver.ExecuteTask.
func (s *Service) ExecuteTask(
	ctx context.Context,
	workflowID, nodeID string,
	budget executive.BudgetTier,
	payloadRef string,
) (executive.EpistemicStatus, string, error) {
	req, err := NewLearningRequestBuilder().
		WithRequestID(fmt.Sprintf("req-exec-%s-%s", workflowID, nodeID)).
		WithDomainSchemaID("idun.reasoning.strategy.v1").
		WithTimeWindow(time.Now().Add(-24*time.Hour), time.Now()).
		WithPolicyFingerprint(s.config.PolicyProfile.PolicyFingerprint).
		Build()
	if err != nil {
		return executive.StatusEscalationRequired, "", err
	}

	res, err := s.RunCycle(ctx, req)
	if err != nil {
		return executive.StatusEscalationRequired, "", err
	}

	if res.Status == StatusAbstained {
		return executive.StatusInsufficientData, res.ResultID, nil
	}
	return executive.StatusConfident, res.ResultID, nil
}

// ConsolidateExperience implements executive.LearningAbility.ConsolidateExperience.
func (s *Service) ConsolidateExperience(ctx context.Context, episodicRef string) (string, error) {
	req, err := NewLearningRequestBuilder().
		WithRequestID(fmt.Sprintf("req-consolidate-%d", time.Now().UnixNano())).
		WithDomainSchemaID("idun.episodic.consolidation.v1").
		WithTimeWindow(time.Now().Add(-24*time.Hour), time.Now()).
		WithPolicyFingerprint(s.config.PolicyProfile.PolicyFingerprint).
		Build()
	if err != nil {
		return "", err
	}

	res, err := s.RunCycle(ctx, req)
	if err != nil {
		return "", err
	}
	return res.ResultID, nil
}
