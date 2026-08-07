package reasoning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	"idun/intelligence/executive"
	"idun/intelligence/infrastructure/embedding"
	"idun/intelligence/infrastructure/inference"
	"idun/intelligence/understanding"
	"idun/intelligence/workspace"
)

// Service implements ReasoningService and executive.ReasoningAbility.
// It orchestrates the complete 10-stage Reasoning Cascade over incoming perceptual envelopes
// and publishes structured ReasoningResult envelopes to the Global Workspace.
type Service struct {
	mu sync.RWMutex

	cfg              Config
	ws               workspace.Workspace
	mem              MemoryProvider
	strategySelector StrategySelector
	contextAssembler ContextAssembler
	symbolicSpec     *SymbolicSpecialist
	relationalSpec   *RelationalGraphSpecialist
	cspSpec          *CSPCheckSpecialist
	bayesianSpec     *BayesianFusionSpecialist
	analogySpec      *CaseAnalogySpecialist
	beamSpec         *BeamSelectionSpecialist
	calibSpec        *CalibrationSpecialist
	delibSpec        *DeliberativeSpecialist
	constSpec        *ConstitutionSpecialist
	telemetry        *telemetryCollector

	subTopic communication.TopicID
	pubTopic communication.TopicID
	sub      workspace.Subscription
	storer   PayloadStorer

	started bool
	closed  bool
}

// ServiceOption configures a Service instance at initialization time.
type ServiceOption func(*Service)

// WithStrategySelector injects a custom strategy routing selector.
func WithStrategySelector(s StrategySelector) ServiceOption {
	return func(srv *Service) {
		srv.strategySelector = s
	}
}

// WithContextAssembler injects a custom Stage S0 context assembler.
func WithContextAssembler(ca ContextAssembler) ServiceOption {
	return func(srv *Service) {
		srv.contextAssembler = ca
	}
}

// WithSymbolicSpecialist injects a custom Stage S1 symbolic specialist.
func WithSymbolicSpecialist(s *SymbolicSpecialist) ServiceOption {
	return func(srv *Service) {
		srv.symbolicSpec = s
	}
}

// WithRelationalSpecialist injects a custom Stage S2 relational graph specialist.
func WithRelationalSpecialist(r *RelationalGraphSpecialist) ServiceOption {
	return func(srv *Service) {
		srv.relationalSpec = r
	}
}

// WithCSPCheckSpecialist injects a custom Stage S3 CSP contradiction specialist.
func WithCSPCheckSpecialist(c *CSPCheckSpecialist) ServiceOption {
	return func(srv *Service) {
		srv.cspSpec = c
	}
}

// WithBayesianSpecialist injects a custom Stage S4 Bayesian evidence fusion specialist.
func WithBayesianSpecialist(b *BayesianFusionSpecialist) ServiceOption {
	return func(srv *Service) {
		srv.bayesianSpec = b
	}
}

// WithAnalogySpecialist injects a custom Stage S5 case analogy specialist.
func WithAnalogySpecialist(a *CaseAnalogySpecialist) ServiceOption {
	return func(srv *Service) {
		srv.analogySpec = a
	}
}

// WithBeamSpecialist injects a custom Stage S6 beam selection specialist.
func WithBeamSpecialist(bm *BeamSelectionSpecialist) ServiceOption {
	return func(srv *Service) {
		srv.beamSpec = bm
	}
}

// WithCalibrationService injects shared Stage S7 calibration service.
func WithCalibrationService(calib calibration.CalibrationService) ServiceOption {
	return func(srv *Service) {
		srv.calibSpec = NewCalibrationSpecialist(calib)
	}
}

// WithInferenceService injects shared Stage S8 inference service.
func WithInferenceService(infer inference.InferenceService) ServiceOption {
	return func(srv *Service) {
		srv.delibSpec = NewDeliberativeSpecialist(infer)
	}
}

// WithConstitutionGate injects shared Stage S9 constitutional action gate.
func WithConstitutionGate(gate constitution.ActionGate) ServiceOption {
	return func(srv *Service) {
		srv.constSpec = NewConstitutionSpecialist(gate)
	}
}

// WithTopics overrides the default subscription and publication topics.
func WithTopics(subTopic, pubTopic communication.TopicID) ServiceOption {
	return func(srv *Service) {
		srv.subTopic = subTopic
		srv.pubTopic = pubTopic
	}
}

// WithPayloadStorer injects a CAS storage bridge for persisting ReasoningResult payloads.
func WithPayloadStorer(storer PayloadStorer) ServiceOption {
	return func(srv *Service) {
		srv.storer = storer
	}
}

// NewService constructs a new Reasoning Service.
func NewService(cfg Config, ws workspace.Workspace, mem MemoryProvider, opts ...ServiceOption) *Service {
	if err := cfg.Validate(); err != nil {
		cfg = DefaultConfig()
	}
	s := &Service{
		cfg:              cfg,
		ws:               ws,
		mem:              mem,
		strategySelector: NewDefaultStrategySelector(),
		contextAssembler: NewDefaultContextAssembler(mem),
		symbolicSpec:     NewSymbolicSpecialist(),
		relationalSpec:   NewRelationalGraphSpecialist(),
		cspSpec:          NewCSPCheckSpecialist(),
		bayesianSpec:     NewBayesianFusionSpecialist(),
		analogySpec:      NewCaseAnalogySpecialist(nil, mem),
		beamSpec:         NewBeamSelectionSpecialist(),
		calibSpec:        NewCalibrationSpecialist(nil),
		delibSpec:        NewDeliberativeSpecialist(nil),
		constSpec:        NewConstitutionSpecialist(nil),
		telemetry:        &telemetryCollector{},
		subTopic:         communication.TopicResolvedIntent,
		pubTopic:         communication.TopicActiveGoals,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
}

// GetTelemetry retrieves an operational statistics snapshot for monitoring.
// Strictly content-blind: exposes counts and durations only.
func (s *Service) GetTelemetry() TelemetrySnapshot {
	return s.telemetry.snapshot()
}

// NewServiceWithEmbedding constructs a new Reasoning Service with shared Embedding service.
func NewServiceWithEmbedding(cfg Config, ws workspace.Workspace, mem MemoryProvider, embedder embedding.EmbeddingService, opts ...ServiceOption) *Service {
	s := NewService(cfg, ws, mem, opts...)
	s.analogySpec = NewCaseAnalogySpecialist(embedder, mem)
	return s
}

// Name returns the canonical Kernel registry name for this cognitive service.
func (s *Service) Name() string {
	return "intelligence.reasoning"
}

// Ability returns the cognitive ability type.
func (s *Service) Ability() executive.CognitiveAbility {
	return executive.AbilityReasoning
}

// Start boots the Reasoning Service and subscribes to workspace perception envelopes if available.
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrServiceClosed
	}
	if s.started {
		return nil
	}

	if s.ws != nil && s.subTopic.IsValid() {
		sub, err := s.ws.Subscribe(s.subTopic, s.Name(), s.handleEnvelope)
		if err != nil {
			return fmt.Errorf("reasoning: failed to subscribe to %q: %w", s.subTopic, err)
		}
		s.sub = sub
	}

	s.started = true
	return nil
}

// Close gracefully shuts down the Reasoning Service and cancels active subscriptions.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrServiceClosed
	}
	if s.sub != nil {
		_ = s.sub.Cancel()
		s.sub = nil
	}
	s.started = false
	s.closed = true
	return nil
}

// ExecuteTask satisfies executive.AbilityDriver.
func (s *Service) ExecuteTask(ctx context.Context, payloadRef string) (executive.EpistemicStatus, string, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return executive.StatusUnsureAmbiguous, "", ErrServiceClosed
	}
	s.mu.RUnlock()

	env := communication.Envelope{
		ID:              fmt.Sprintf("exec-task-%d", time.Now().UnixNano()),
		Source:          "executive",
		Topic:           communication.TopicResolvedIntent,
		PayloadRef:      payloadRef,
		PayloadModality: "structured-frame",
		RawConfidence:   1.0,
	}

	result, err := s.ReasonEnvelope(ctx, env, StrategySpec{})
	if err != nil {
		return executive.StatusUnsureAmbiguous, "", err
	}

	return executive.StatusConfident, result.EnvelopeID, nil
}

// SynthesizeInference satisfies executive.ReasoningAbility.
func (s *Service) SynthesizeInference(ctx context.Context, premisesRef string) (string, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return "", ErrServiceClosed
	}
	s.mu.RUnlock()

	env := communication.Envelope{
		ID:              fmt.Sprintf("synth-inf-%d", time.Now().UnixNano()),
		Source:          "executive.reasoning",
		Topic:           communication.TopicResolvedIntent,
		PayloadRef:      premisesRef,
		PayloadModality: "structured-frame",
		RawConfidence:   0.90,
	}

	result, err := s.ReasonEnvelope(ctx, env, StrategySpec{})
	if err != nil {
		return "", err
	}

	return result.PrimaryHypothesis.Conclusion, nil
}

// ReasonEnvelope executes the complete Reasoning Cascade for an incoming perceptual Envelope
// and publishes the resulting ReasoningResult to the Global Workspace.
func (s *Service) ReasonEnvelope(ctx context.Context, perceptionEnv communication.Envelope, spec StrategySpec) (ReasoningResult, error) {
	startTime := time.Now()

	s.mu.RLock()
	closed := s.closed
	strategySelector := s.strategySelector
	contextAssembler := s.contextAssembler
	symbolicSpec := s.symbolicSpec
	relationalSpec := s.relationalSpec
	cspSpec := s.cspSpec
	bayesianSpec := s.bayesianSpec
	analogySpec := s.analogySpec
	beamSpec := s.beamSpec
	calibSpec := s.calibSpec
	delibSpec := s.delibSpec
	constSpec := s.constSpec
	ws := s.ws
	pubTopic := s.pubTopic
	storer := s.storer
	s.mu.RUnlock()

	if closed {
		return ReasoningResult{}, ErrServiceClosed
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			s.telemetry.recordTimeout()
		} else {
			s.telemetry.recordCancellation()
		}
		return ReasoningResult{}, err
	}

	if spec.StrategyID == "" {
		selected, err := strategySelector.SelectStrategy(ctx, perceptionEnv)
		if err != nil {
			return ReasoningResult{}, err
		}
		spec = selected
	}

	traceLogs := make([]StageTraceLog, 0, 10)
	execSpecialists := make([]StageIdentifier, 0, 10)

	// Stage S0: Context & Strategy Assembly
	s0Start := time.Now()
	memContext, err := contextAssembler.AssembleContext(ctx, perceptionEnv, 20)
	if err != nil {
		return ReasoningResult{}, fmt.Errorf("stage S0 context assembly failed: %w", err)
	}
	traceLogs = append(traceLogs, StageTraceLog{
		Stage:               StageS0ContextAssembly,
		ExecutionDurationMs: float64(time.Since(s0Start).Microseconds()) / 1000.0,
		Description:         fmt.Sprintf("Retrieved %d memory records for context assembly", len(memContext)),
		OutputSummary:       fmt.Sprintf("strategy=%s records=%d", spec.StrategyID, len(memContext)),
	})
	execSpecialists = append(execSpecialists, StageS0ContextAssembly)

	var frame *understanding.SemanticFrame
	if perceptionEnv.PayloadRef != "" {
		var frameBytes []byte
		if storer != nil {
			if data, err := storer.Retrieve(ctx, perceptionEnv.PayloadRef); err == nil && len(data) > 0 {
				frameBytes = data
			}
		}
		if len(frameBytes) == 0 {
			frameBytes = []byte(perceptionEnv.PayloadRef)
		}
		var decoded understanding.SemanticFrame
		if err := json.Unmarshal(frameBytes, &decoded); err == nil {
			if valErr := decoded.Validate(); valErr == nil {
				frame = &decoded
			}
		}
	}

	var hypotheses []ReasoningHypothesis

	// Stage S1: Symbolic Fast Path
	if spec.IsStageEnabled(StageS1SymbolicFast) {
		s1Start := time.Now()
		hyps, err := symbolicSpec.Evaluate(ctx, perceptionEnv, frame, memContext)
		if err != nil {
			return ReasoningResult{}, fmt.Errorf("stage S1 symbolic evaluation failed: %w", err)
		}
		hypotheses = append(hypotheses, hyps...)
		traceLogs = append(traceLogs, StageTraceLog{
			Stage:               StageS1SymbolicFast,
			ExecutionDurationMs: float64(time.Since(s1Start).Microseconds()) / 1000.0,
			Description:         "Evaluated symbolic forward-chaining rules",
			OutputSummary:       fmt.Sprintf("hypotheses=%d", len(hyps)),
		})
		execSpecialists = append(execSpecialists, StageS1SymbolicFast)
	}

	// Stage S2: Relational Graph Reasoning (Ephemeral session-scoped graph)
	if spec.IsStageEnabled(StageS2RelationalGraph) {
		s2Start := time.Now()
		hyps, err := relationalSpec.Evaluate(ctx, perceptionEnv, frame, memContext, spec)
		if err != nil {
			return ReasoningResult{}, fmt.Errorf("stage S2 relational graph evaluation failed: %w", err)
		}
		hypotheses = append(hypotheses, hyps...)
		traceLogs = append(traceLogs, StageTraceLog{
			Stage:               StageS2RelationalGraph,
			ExecutionDurationMs: float64(time.Since(s2Start).Microseconds()) / 1000.0,
			Description:         "Evaluated session-scoped relational working graph",
			OutputSummary:       fmt.Sprintf("paths_hypotheses=%d", len(hyps)),
		})
		execSpecialists = append(execSpecialists, StageS2RelationalGraph)
	}

	// Stage S5: Case-Based / Analogical Reasoning
	if spec.IsStageEnabled(StageS5CaseAnalogy) {
		s5Start := time.Now()
		hyps, err := analogySpec.EvaluateAnalogy(ctx, perceptionEnv, frame, memContext)
		if err != nil {
			return ReasoningResult{}, fmt.Errorf("stage S5 case analogy evaluation failed: %w", err)
		}
		hypotheses = append(hypotheses, hyps...)
		traceLogs = append(traceLogs, StageTraceLog{
			Stage:               StageS5CaseAnalogy,
			ExecutionDurationMs: float64(time.Since(s5Start).Microseconds()) / 1000.0,
			Description:         "Evaluated case-based analogical retrieval",
			OutputSummary:       fmt.Sprintf("analogies=%d", len(hyps)),
		})
		execSpecialists = append(execSpecialists, StageS5CaseAnalogy)
	}

	if len(hypotheses) == 0 {
		hypotheses = []ReasoningHypothesis{
			{
				ID:                  fmt.Sprintf("default-hyp-%s", perceptionEnv.ID),
				Type:                HypothesisInference,
				Conclusion:          fmt.Sprintf("Inferred default conclusion for %s", perceptionEnv.ID),
				ReasoningConfidence: 0.80,
				ContributingStages:  []StageIdentifier{StageS0ContextAssembly},
			},
		}
	}

	// Stage S4: Bayesian Evidence Fusion
	if spec.IsStageEnabled(StageS4EvidenceFusion) && len(hypotheses) > 0 {
		s4Start := time.Now()
		fused, err := bayesianSpec.FuseEvidence(ctx, hypotheses)
		if err != nil {
			return ReasoningResult{}, fmt.Errorf("stage S4 bayesian evidence fusion failed: %w", err)
		}
		hypotheses = fused
		traceLogs = append(traceLogs, StageTraceLog{
			Stage:               StageS4EvidenceFusion,
			ExecutionDurationMs: float64(time.Since(s4Start).Microseconds()) / 1000.0,
			Description:         "Fused evidence and updated hypothesis confidences",
			OutputSummary:       fmt.Sprintf("fused=%d", len(fused)),
		})
		execSpecialists = append(execSpecialists, StageS4EvidenceFusion)
	}

	// Stage S3: CSP Constraint Consistency Check
	var contradictions []ContradictionFlag
	if spec.IsStageEnabled(StageS3CSPCheck) {
		s3Start := time.Now()
		cFlags, err := cspSpec.CheckConsistency(ctx, hypotheses, memContext)
		if err != nil {
			return ReasoningResult{}, fmt.Errorf("stage S3 CSP consistency check failed: %w", err)
		}
		contradictions = cFlags
		traceLogs = append(traceLogs, StageTraceLog{
			Stage:               StageS3CSPCheck,
			ExecutionDurationMs: float64(time.Since(s3Start).Microseconds()) / 1000.0,
			Description:         "Evaluated CSP consistency check against memory context",
			OutputSummary:       fmt.Sprintf("contradictions=%d", len(cFlags)),
		})
		execSpecialists = append(execSpecialists, StageS3CSPCheck)
	} else {
		contradictions = []ContradictionFlag{}
	}

	// Stage S6: Multi-Hypothesis Beam Selection
	s6Start := time.Now()
	primary, beam, err := beamSpec.SelectBeam(hypotheses, MaxBeamWidth, 0.25)
	if err != nil {
		return ReasoningResult{}, fmt.Errorf("stage S6 beam selection failed: %w", err)
	}
	traceLogs = append(traceLogs, StageTraceLog{
		Stage:               StageS6BeamSelection,
		ExecutionDurationMs: float64(time.Since(s6Start).Microseconds()) / 1000.0,
		Description:         fmt.Sprintf("Selected primary hypothesis and %d beam ambiguity runners-up", len(beam)),
		OutputSummary:       fmt.Sprintf("primary=%s beam_size=%d", primary.ID, len(beam)),
	})
	execSpecialists = append(execSpecialists, StageS6BeamSelection)

	escalated := false
	// Stage S8: Deliberative LLM Reasoning (escalate only if primary confidence < EscalationThreshold or if no internal specialist produced a valid ProposedGoal above threshold)
	if spec.IsStageEnabled(StageS8DeliberativeLLM) && delibSpec != nil {
		shouldEscalate := primary.ReasoningConfidence < spec.EscalationThreshold
		if !shouldEscalate {
			hasValidProposedGoal := false
			for _, h := range hypotheses {
				if h.ProposedGoal != nil && h.ProposedGoal.Validate() == nil && h.ReasoningConfidence >= spec.EscalationThreshold {
					hasValidProposedGoal = true
					break
				}
			}
			if !hasValidProposedGoal {
				shouldEscalate = true
			}
		}

		if shouldEscalate {
			s8Start := time.Now()
			triggerConf := primary.ReasoningConfidence
			if triggerConf >= spec.EscalationThreshold {
				triggerConf = 0.0
			}
			delibHyps, err := delibSpec.EvaluateDeliberative(ctx, perceptionEnv, triggerConf, spec.EscalationThreshold, s.storer)
			if err == nil && len(delibHyps) > 0 {
				escalated = true
				combined := append(hypotheses, delibHyps...)
				if spec.IsStageEnabled(StageS4EvidenceFusion) && bayesianSpec != nil {
					if fusedCombined, err := bayesianSpec.FuseEvidence(ctx, combined); err == nil {
						combined = fusedCombined
					}
				}
				primary, beam, _ = beamSpec.SelectBeam(combined, MaxBeamWidth, 0.25)
				traceLogs = append(traceLogs, StageTraceLog{
					Stage:               StageS8DeliberativeLLM,
					ExecutionDurationMs: float64(time.Since(s8Start).Microseconds()) / 1000.0,
					Description:         "Escalated to Stage S8 Deliberative LLM and updated beam selection",
					OutputSummary:       fmt.Sprintf("deliberative_hyps=%d", len(delibHyps)),
				})
				execSpecialists = append(execSpecialists, StageS8DeliberativeLLM)
			}
		}
	}

	// Stage S7: Calibration Integration (Single-Owner writer of CalibratedConfidence)
	s7Start := time.Now()
	calPrimary, calBeam, err := calibSpec.CalibrateHypotheses(ctx, s.Name(), pubTopic, primary, beam)
	if err != nil {
		return ReasoningResult{}, fmt.Errorf("stage S7 calibration integration failed: %w", err)
	}
	traceLogs = append(traceLogs, StageTraceLog{
		Stage:               StageS7Calibration,
		ExecutionDurationMs: float64(time.Since(s7Start).Microseconds()) / 1000.0,
		Description:         "Applied historical calibration to primary and beam hypotheses",
		OutputSummary:       fmt.Sprintf("calibrated_conf=%.4f", calPrimary.CalibratedConfidence),
	})
	execSpecialists = append(execSpecialists, StageS7Calibration)

	durationMs := float64(time.Since(startTime).Microseconds()) / 1000.0

	// Stage S10: Final Assembly
	execSpecialists = append(execSpecialists, StageS10ResultAssembly)
	telemetry := StrategyTelemetry{
		EpisodeID:            fmt.Sprintf("ep-%d", time.Now().UnixNano()),
		StrategySelected:     spec.StrategyID,
		SpecialistsExecuted:  execSpecialists,
		ExecutionDurationMs:  durationMs,
		CalibratedConfidence: calPrimary.CalibratedConfidence,
		ResourceCostTier:     "LOCAL_HYBRID",
		EscalatedToLLM:       escalated,
		OutcomeStatus:        StatusUnambiguousSolved,
	}

	var resolvedGoal *SemanticGoal
	if calPrimary.ProposedGoal != nil && calPrimary.ProposedGoal.Validate() == nil {
		resolvedGoal = calPrimary.ProposedGoal.Clone()
	}

	// --- Phase 2C Deliberation Episode Construction ---
	episode := &DeliberationEpisode{
		EpisodeID:              telemetry.EpisodeID,
		SemanticFrameReference: perceptionEnv.ID,
		GoalGraph:              []*SemanticGoal{},
		AlternativeEvaluations: []AlternativeEvaluation{},
		ConstraintEvaluations:  []string{},
		RiskAssessments:        []RiskAssessment{},
		AcceptedAssumptions:    []string{},
		RejectedAssumptions:    []string{},
		Trace: &DeliberationTrace{
			Events: []string{"Deliberation started", "Evaluated candidates", "Selected primary goal"},
		},
		FinalGoalSelection:  []string{},
		ReasoningConfidence: calPrimary.CalibratedConfidence,
	}

	if resolvedGoal != nil {
		episode.GoalGraph = append(episode.GoalGraph, resolvedGoal)
		episode.FinalGoalSelection = append(episode.FinalGoalSelection, resolvedGoal.GoalID)

		// Map alternative goals
		for _, b := range calBeam {
			if b.ProposedGoal != nil {
				episode.GoalGraph = append(episode.GoalGraph, b.ProposedGoal)
				altEval := AlternativeEvaluation{
					AlternativeID:    b.ProposedGoal.GoalID,
					GenerationReason: "Ambiguity beam runner-up",
					Confidence:       b.CalibratedConfidence,
					RejectionReason:  "Lower confidence than primary",
				}
				episode.AlternativeEvaluations = append(episode.AlternativeEvaluations, altEval)
			}
		}
	}

	resultEnvID := fmt.Sprintf("rs-env-%d", time.Now().UnixNano())
	result := ReasoningResult{
		SchemaVersion:           SchemaVersion,
		EnvelopeID:              resultEnvID,
		SourceFrameID:           perceptionEnv.ID,
		Status:                  StatusUnambiguousSolved,
		StrategyUsed:            spec.StrategyID,
		PrimaryHypothesis:       calPrimary,
		AmbiguitySet:            calBeam,
		ContradictionsFlagged:   contradictions,
		ProposedBeliefUpdates:   []BeliefUpdateProposal{},
		ResolvedGoal:            resolvedGoal,
		StrategyTelemetry:       telemetry,
		ConstitutionAnnotations: []string{},
		ReasoningTrace:          traceLogs,
		OfflineMode:             !escalated,
		ProcessedDurationMs:     durationMs,
		DeliberationEpisode:     episode,
	}

	// Stage S9: Constitution Integration
	if constSpec != nil {
		s9Start := time.Now()
		_ = constSpec.EvaluateResult(ctx, &result)
		traceLogs = append(traceLogs, StageTraceLog{
			Stage:               StageS9Constitution,
			ExecutionDurationMs: float64(time.Since(s9Start).Microseconds()) / 1000.0,
			Description:         "Submitted reasoning result to Constitutional Action Gate",
			OutputSummary:       fmt.Sprintf("annotations=%d", len(result.ConstitutionAnnotations)),
		})
	}

	if err := result.Validate(); err != nil {
		s.telemetry.recordValidationFailure()
		return ReasoningResult{}, fmt.Errorf("reasoning result validation failed: %w", err)
	}

	devLog("Reasoning", "Goal created")

	s.telemetry.recordEpisode(
		durationMs,
		len(calBeam)+1,
		spec.IsStageEnabled(StageS1SymbolicFast),
		spec.IsStageEnabled(StageS2RelationalGraph),
		spec.IsStageEnabled(StageS4EvidenceFusion),
		spec.IsStageEnabled(StageS5CaseAnalogy),
		escalated,
		len(calBeam) > 0,
		calibSpec != nil,
		constSpec != nil,
	)

	if ws != nil && pubTopic.IsValid() {
		payloadRef := resultEnvID
		if s.storer != nil {
			if data, jerr := json.Marshal(result); jerr == nil {
				if key, serr := s.storer.Store(ctx, data); serr == nil {
					payloadRef = key
				}
			}
		}
		parentID := perceptionEnv.ParentRef
		if parentID == "" {
			parentID = perceptionEnv.ID
		}
		pubEnv := communication.Envelope{
			ID:              resultEnvID,
			Source:          s.Name(),
			Topic:           pubTopic,
			ParentRef:       parentID,
			PayloadRef:      payloadRef,
			PayloadModality: "reasoning-result",
			RawConfidence:   calPrimary.CalibratedConfidence,
		}
		_ = ws.Publish(ctx, pubEnv)
		devLog("Reasoning", "Published TopicActiveGoals")
	}

	return result, nil
}

func (s *Service) handleEnvelope(ctx context.Context, env communication.Envelope) error {
	devLog("Reasoning", "Received TopicResolvedIntent")
	_, err := s.ReasonEnvelope(ctx, env, StrategySpec{})
	return err
}
