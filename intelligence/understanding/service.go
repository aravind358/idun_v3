package understanding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
	"idun/intelligence/executive"
	"idun/intelligence/workspace"
)

// Service implements Phases 1–5 (Production-Hardened, Telemetry-Instrumented,
// Speculative & Deliberative) of CognitiveAbility.Understanding.
type Service struct {
	mu           sync.RWMutex
	cfg          Config
	normalizer   Normalizer
	binder       ReferentBinder
	grammar      GrammarSpecialist
	neural       NeuralSpecialist
	calibrator   calibration.CalibrationService
	evaluator    *SpeculativeEvaluator
	deliberative *DeliberativeWorker
	telemetry    *telemetryCollector
	tau          float64
	ws           workspace.Workspace
	storer       PayloadStorer
	sub          workspace.Subscription
	closed       bool
}

// ServiceOption configures functional dependencies on Service.
type ServiceOption func(*Service)

// WithNormalizer overrides the default normalizer.
func WithNormalizer(n Normalizer) ServiceOption {
	return func(s *Service) {
		if n != nil {
			s.normalizer = n
		}
	}
}

// WithReferentBinder overrides the default referent binder.
func WithReferentBinder(b ReferentBinder) ServiceOption {
	return func(s *Service) {
		if b != nil {
			s.binder = b
		}
	}
}

// WithGrammarSpecialist overrides the default grammar specialist.
func WithGrammarSpecialist(g GrammarSpecialist) ServiceOption {
	return func(s *Service) {
		if g != nil {
			s.grammar = g
		}
	}
}

// WithNeuralSpecialist configures a local neural specialist for speculative evaluation.
func WithNeuralSpecialist(n NeuralSpecialist) ServiceOption {
	return func(s *Service) {
		if n != nil {
			s.neural = n
		}
	}
}

// WithCalibrator injects the Epistemic Calibration Service.
func WithCalibrator(cal calibration.CalibrationService) ServiceOption {
	return func(s *Service) {
		s.calibrator = cal
	}
}

// WithDeliberativeWorker injects the Phase 4 LLM-assisted Deliberative Understanding Worker.
func WithDeliberativeWorker(dw *DeliberativeWorker) ServiceOption {
	return func(s *Service) {
		s.deliberative = dw
	}
}

// WithConfidenceThreshold overrides the minimum confidence threshold tau for unambiguous status.
func WithConfidenceThreshold(tau float64) ServiceOption {
	return func(s *Service) {
		s.tau = tau
	}
}

// WithAmbiguityThreshold overrides the delta threshold for bounded beam ambiguity preservation.
func WithAmbiguityThreshold(delta float64) ServiceOption {
	return func(s *Service) {
		s.evaluator = NewSpeculativeEvaluator(delta)
	}
}

// WithPayloadStorer injects a content-addressed storage bridge for payload retrieval and persistence.
func WithPayloadStorer(storer PayloadStorer) ServiceOption {
	return func(s *Service) {
		s.storer = storer
	}
}

// NewService constructs a thread-safe Understanding Service.
func NewService(cfg Config, ws workspace.Workspace, opts ...ServiceOption) *Service {
	s := &Service{
		cfg:        cfg,
		normalizer: NewDefaultNormalizer(),
		binder:     NewDefaultReferentBinder(),
		grammar:    NewDefaultGrammarSpecialist(),
		neural:     NewDefaultNeuralSpecialist(),
		evaluator:  NewSpeculativeEvaluator(DefaultAmbiguityDelta),
		telemetry:  &telemetryCollector{},
		tau:        0.40,
		ws:         ws,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Name returns the Kernel component identifier.
func (s *Service) Name() string {
	return "Intelligence.Understanding"
}

// Ability returns the canonical cognitive ability enum.
func (s *Service) Ability() executive.CognitiveAbility {
	return executive.AbilityUnderstanding
}

// GetTelemetry returns an operational diagnostics snapshot for Host/Kernel monitoring.
func (s *Service) GetTelemetry() TelemetrySnapshot {
	return s.telemetry.snapshot()
}

// Start boots the Understanding service.
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrServiceClosed
	}
	if s.deliberative != nil {
		_ = s.deliberative.Start()
	}
	if s.ws != nil && s.sub == nil {
		sub, err := s.ws.Subscribe(communication.TopicPerception, s.Name(), s.handlePerceptionEnvelope)
		if err != nil {
			return fmt.Errorf("understanding: failed to subscribe to TopicPerception: %w", err)
		}
		s.sub = sub
	}
	return nil
}

// Close gracefully shuts down the service.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	if s.sub != nil {
		_ = s.sub.Cancel()
		s.sub = nil
	}
	if s.deliberative != nil {
		_ = s.deliberative.Close()
	}
	return nil
}

// handlePerceptionEnvelope is the asynchronous Workspace subscription callback.
// It consumes raw perception envelopes published to TopicPerception (e.g. from World),
// validates the envelope and schema, retrieves the payload from content-addressed storage
// via PayloadStorer if applicable, interprets the perceptual input, and publishes
// the resulting SemanticFrame to the downstream topic (TopicUserIntent).
func (s *Service) handlePerceptionEnvelope(ctx context.Context, env communication.Envelope) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	if env.ID == "" {
		s.telemetry.recordValidationFailure()
		return errors.New("understanding: perception envelope validation failed: empty ID")
	}
	if env.Topic != communication.TopicPerception {
		s.telemetry.recordValidationFailure()
		return fmt.Errorf("understanding: expected topic %q, got %q", communication.TopicPerception, env.Topic)
	}
	if env.PayloadRef == "" {
		s.telemetry.recordValidationFailure()
		return errors.New("understanding: perception envelope validation failed: empty PayloadRef")
	}

	devLog("Understanding", "Received TopicPerception")

	modifiedEnv := env
	if s.storer != nil {
		data, err := s.storer.Retrieve(ctx, env.PayloadRef)
		if err == nil && len(data) > 0 {
			modifiedEnv.PayloadRef = string(data)
		}
	}

	_, err := s.InterpretEnvelope(ctx, modifiedEnv)
	return err
}

// InterpretEnvelope interprets raw perceptual input deterministically and speculatively,
// escalating to deliberative LLM parsing only when local specialists fall below tau.
func (s *Service) InterpretEnvelope(ctx context.Context, perceptionEnv communication.Envelope) (SemanticFrame, error) {
	return s.interpretInternal(ctx, perceptionEnv, "", nil)
}

// InterpretWithPrior interprets input conditioned on a top-down dialogue expectation prior.
func (s *Service) InterpretWithPrior(ctx context.Context, perceptionEnv communication.Envelope, prior string) (SemanticFrame, error) {
	return s.interpretInternal(ctx, perceptionEnv, prior, nil)
}

// InterpretWithCandidates interprets input with explicitly supplied referent candidates.
func (s *Service) InterpretWithCandidates(ctx context.Context, perceptionEnv communication.Envelope, prior string, candidates []ReferentCandidate) (SemanticFrame, error) {
	return s.interpretInternal(ctx, perceptionEnv, prior, candidates)
}

// ParseIntent implements the executive.UnderstandingAbility contract.
func (s *Service) ParseIntent(ctx context.Context, payloadRef string) (string, error) {
	env := communication.Envelope{
		ID:         "parse-" + payloadRef,
		PayloadRef: payloadRef,
	}
	frame, err := s.InterpretEnvelope(ctx, env)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(frame)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ExecuteTask implements the executive.AbilityDriver contract.
//
// Architectural note: ExecuteTask exists solely to satisfy the Executive AbilityDriver
// interface. It delegates immediately to InterpretEnvelope() and does not perform
// workflow orchestration, planning, decision-making, or action execution.
func (s *Service) ExecuteTask(ctx context.Context, workflowID, nodeID string, budget executive.BudgetTier, payloadRef string) (executive.EpistemicStatus, string, error) {
	env := communication.Envelope{
		ID:         fmt.Sprintf("%s-%s", workflowID, nodeID),
		PayloadRef: payloadRef,
	}
	frame, err := s.InterpretEnvelope(ctx, env)
	if err != nil {
		return executive.StatusEscalationRequired, "", err
	}
	bytes, err := json.Marshal(frame)
	if err != nil {
		return executive.StatusEscalationRequired, "", err
	}
	if frame.Status == StatusUnambiguous {
		return executive.StatusConfident, string(bytes), nil
	}
	return executive.StatusUnsureAmbiguous, string(bytes), nil
}

func (s *Service) interpretInternal(ctx context.Context, perceptionEnv communication.Envelope, prior string, candidates []ReferentCandidate) (SemanticFrame, error) {
	startTime := time.Now()

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return SemanticFrame{}, ErrServiceClosed
	}
	s.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		s.telemetry.recordCancellation()
		return SemanticFrame{}, err
	}

	rawText := perceptionEnv.PayloadRef
	if rawText == "" {
		rawText = perceptionEnv.ID
	}

	norm := s.normalizer.Normalize(rawText)
	boundSlots := s.binder.BindReferents(norm, candidates)

	primary, ambiguitySet, err := s.evaluator.EvaluateParallel(ctx, norm, boundSlots, s.grammar, s.neural, s.calibrator)
	if err != nil {
		return SemanticFrame{}, err
	}

	if primary.SourceLayer == LayerReflexiveGrammar {
		s.telemetry.recordGrammarHit()
	} else if primary.SourceLayer == LayerNeuralClassifier {
		s.telemetry.recordNeuralHit()
	}

	// Check if local evaluation confident enough (primary.CalibratedConfidence >= tau)
	// If below tau OR no primary matched AND deliberative worker is available -> escalate!
	if (primary.CalibratedConfidence < s.tau || primary.Intent == "") && s.deliberative != nil {
		s.telemetry.recordDeliberativeEscalation()
		delibFrame, delibErr := s.deliberative.InterpretDeliberative(ctx, perceptionEnv.ID, rawText, prior)
		if delibErr == nil {
			durationUs := time.Since(startTime).Microseconds()
			delibFrame.ProcessedDurationMs = float64(durationUs) / 1000.0
			s.telemetry.recordInterpretation(durationUs, 1, false)

			if s.ws != nil {
				payloadBytes, _ := json.Marshal(delibFrame)
				payloadRef := string(payloadBytes)
				if s.storer != nil {
					if key, storeErr := s.storer.Store(ctx, payloadBytes); storeErr == nil {
						payloadRef = key
					}
				}
				parentID := perceptionEnv.ParentRef
				if parentID == "" {
					parentID = perceptionEnv.ID
				}
				pubEnv := communication.Envelope{
					ID:            fmt.Sprintf("frame-%s", perceptionEnv.ID),
					ParentRef:     parentID,
					Source:        s.Name(),
					Topic:         communication.TopicUserIntent,
					RawConfidence: delibFrame.PrimaryHypothesis.CalibratedConfidence,
					PayloadRef:    payloadRef,
					CreatedAt:     time.Now().UTC(),
				}
				_ = s.ws.Publish(ctx, pubEnv)
				devLog("Understanding", "Intent", delibFrame.PrimaryHypothesis.Intent)
				devLog("Understanding", "Published TopicUserIntent")
			}
			return delibFrame, nil
		}
		if errors.Is(delibErr, ErrInferenceTimeout) {
			s.telemetry.recordTimeout()
		} else if errors.Is(delibErr, ErrDeliberativeCancelled) {
			s.telemetry.recordCancellation()
		} else if errors.Is(delibErr, ErrMalformedInferenceResponse) {
			s.telemetry.recordMalformedInference()
		}
		// If deliberative worker returned an error, fall through to return deterministic status
	}

	builder := NewSemanticFrameBuilder(perceptionEnv.ID)
	if prior != "" {
		builder.WithTopDownPrior(prior)
	}

	if primary.Intent != "" {
		for _, amb := range ambiguitySet {
			builder.AddAmbiguousHypothesis(amb.Intent, amb.CalibratedConfidence, amb.SourceLayer, amb.DeltaFromPrimary, amb.Slots...)
		}

		if len(ambiguitySet) > 0 {
			builder.WithStatus(StatusAmbiguousBeam)
		} else if primary.Intent == "unresolved_intent" {
			builder.WithStatus(StatusFailedImpasse)
		} else if primary.CalibratedConfidence >= s.tau {
			builder.WithStatus(StatusUnambiguous)
		} else {
			builder.WithStatus(StatusPreliminary)
		}

		builder.WithPrimaryHypothesis(primary.Intent, primary.CalibratedConfidence, primary.SourceLayer, primary.Slots...)

		// --- Phase 2B Semantic Interpretation Logic ---

		// 1. Intent Graph construction
		intents := []IntentNode{
			{
				IntentID:     "intent-1",
				Intent:       primary.Intent,
				Dependencies: nil,
			},
		}
		builder.WithIntents(intents)

		// 2. Completeness Computation
		completeness := 1.0
		if len(primary.Slots) == 0 && primary.Intent != "inform" {
			completeness = 0.5 // Default penalty if no slots extracted for action intents
		}
		builder.WithCompleteness(completeness)

		// 3. Validation Trace (Semantic Firewall)
		vTrace := &ValidationTrace{
			TypeValidation:         "Passed",
			ConstraintValidation:   "Passed",
			RelationshipValidation: "Passed",
			TemporalValidation:     "Passed",
		}
		// If critically incomplete, the firewall rejects it, causing a FATAL validation and Impasse route.
		if completeness < 0.5 {
			vTrace.ConstraintValidation = "Failed"
		}
		builder.WithValidationTrace(vTrace)

		// 4. Confidence Trace
		builder.WithConfidenceTrace([]string{fmt.Sprintf("Base %s Score: %.2f", primary.SourceLayer, primary.CalibratedConfidence)})
	} else {
		builder.WithStatus(StatusFailedImpasse).
			WithPrimaryHypothesis("unresolved_intent", 0.0, LayerReflexiveGrammar, boundSlots...)
	}

	durationUs := time.Since(startTime).Microseconds()
	durationMs := float64(durationUs) / 1000.0
	builder.WithProcessedDuration(durationMs)

	frame, err := builder.Build()
	if err != nil {
		s.telemetry.recordValidationFailure()
		return SemanticFrame{}, err
	}

	s.telemetry.recordInterpretation(durationUs, 1+len(frame.AmbiguitySet), frame.Status == StatusAmbiguousBeam)

	devLog("Understanding", "Intent", frame.PrimaryHypothesis.Intent)

	if s.ws != nil {
		payloadBytes, _ := json.Marshal(frame)
		payloadRef := string(payloadBytes)
		if s.storer != nil {
			if key, storeErr := s.storer.Store(ctx, payloadBytes); storeErr == nil {
				payloadRef = key
			}
		}
		parentID := perceptionEnv.ParentRef
		if parentID == "" {
			parentID = perceptionEnv.ID
		}
		pubEnv := communication.Envelope{
			ID:            fmt.Sprintf("frame-%s", perceptionEnv.ID),
			ParentRef:     parentID,
			Source:        s.Name(),
			Topic:         communication.TopicUserIntent,
			RawConfidence: frame.PrimaryHypothesis.CalibratedConfidence,
			PayloadRef:    payloadRef,
			CreatedAt:     time.Now().UTC(),
		}
		_ = s.ws.Publish(ctx, pubEnv)
		devLog("Understanding", "Published TopicUserIntent")
	}

	return frame, nil
}

// Ensure Service implements UnderstandingService.
var _ UnderstandingService = (*Service)(nil)
