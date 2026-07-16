package executive

import (
	"context"
	"fmt"
	"sync"
	"time"

	"idun/intelligence/calibration"
	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	"idun/intelligence/workspace"
)

const (
	defaultAlpha  = 0.1 // Urgency multiplier
	defaultBeta   = 0.5 // Cost discount ratio multiplier
	defaultBudget = 1000
)

// ServiceV2 is the concrete, thread-safe implementation of Executive Functions Version 2.0.
// It embeds Version 1 (*Service) to reuse all existing administrative coordination code.
type ServiceV2 struct {
	*ExecutiveService // Embedded Version 1 Executive Service

	mu           sync.RWMutex
	closedV2     bool
	ws           workspace.Workspace
	cal          calibration.CalibrationService
	constGate    constitution.ActionGate
	budget       int
	alpha        float64
	beta         float64
	policyHolder *PolicySnapshotHolder
	capsHolder   *CapabilitiesSnapshotHolder
}

// OptionV2 configures functional options for ServiceV2 construction.
type OptionV2 func(*ServiceV2)

// WithAlphaBeta configures custom Effective Priority formula weights.
func WithAlphaBeta(alpha, beta float64) OptionV2 {
	return func(s *ServiceV2) {
		s.alpha = alpha
		s.beta = beta
	}
}

// WithPolicy configures a custom initial ExecutivePolicyProfile snapshot.
func WithPolicy(profile *ExecutivePolicyProfile) OptionV2 {
	return func(s *ServiceV2) {
		if profile != nil && s.policyHolder != nil {
			_ = s.policyHolder.Store(profile)
		}
	}
}

// WithCapabilities configures custom deployment ExecutiveCapabilities.
func WithCapabilities(caps *ExecutiveCapabilities) OptionV2 {
	return func(s *ServiceV2) {
		if caps != nil && s.capsHolder != nil {
			_ = s.capsHolder.Store(caps)
		}
	}
}

// NewServiceV2 constructs a new Executive Functions Version 2.0 service.
func NewServiceV2(
	ws workspace.Workspace,
	cal calibration.CalibrationService,
	constGate constitution.ActionGate,
	initialBudget int,
	opts ...OptionV2,
) (*ServiceV2, error) {
	if ws == nil {
		return nil, ErrNilWorkspace
	}
	if cal == nil {
		return nil, ErrNilCalibration
	}
	if constGate == nil {
		return nil, ErrNilConstitution
	}
	if initialBudget <= 0 {
		initialBudget = defaultBudget
	}

	v1Svc := NewExecutiveService(Config{})
	policyHolder, _ := NewPolicySnapshotHolder(DefaultExecutivePolicyProfile())
	capsHolder, _ := NewCapabilitiesSnapshotHolder(DefaultExecutiveCapabilities())

	s := &ServiceV2{
		ExecutiveService: v1Svc,
		ws:               ws,
		cal:              cal,
		constGate:        constGate,
		budget:           initialBudget,
		alpha:            defaultAlpha,
		beta:             defaultBeta,
		policyHolder:     policyHolder,
		capsHolder:       capsHolder,
	}

	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

// Workspace returns the integrated Global Workspace & Leveled Blackboard engine.
func (s *ServiceV2) Workspace() workspace.Workspace {
	return s.ws
}

// Calibration returns the integrated Epistemic Calibration service.
func (s *ServiceV2) Calibration() calibration.CalibrationService {
	return s.cal
}

// Constitution returns the integrated Pre-Broadcast Constitutional Action Gate.
func (s *ServiceV2) Constitution() constitution.ActionGate {
	return s.constGate
}

// Policy returns the currently active, immutable ExecutivePolicyProfile snapshot.
func (s *ServiceV2) Policy() *ExecutivePolicyProfile {
	return s.policyHolder.Load()
}

// Capabilities returns the immutable ExecutiveCapabilities snapshot for this deployment.
func (s *ServiceV2) Capabilities() *ExecutiveCapabilities {
	return s.capsHolder.Load()
}

// UpdatePolicy atomically replaces the active policy profile with a newly validated snapshot from Learning.
func (s *ServiceV2) UpdatePolicy(profile *ExecutivePolicyProfile) error {
	return s.policyHolder.Store(profile)
}

// Start boots both Version 1 and Version 2.0 lifecycle hooks.

func (s *ServiceV2) Start() error {
	if err := s.ExecutiveService.Start(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closedV2 {
		return ErrExecutiveV2Closed
	}
	return nil
}

// Close gracefully shuts down Version 2.0 and embedded Version 1 services.
func (s *ServiceV2) Close() error {
	s.mu.Lock()
	if !s.closedV2 {
		s.closedV2 = true
	}
	s.mu.Unlock()
	return s.ExecutiveService.Close()
}

// SubmitBid enqueues a candidate Envelope bid into the specified Horizon queue for arbitration.
func (s *ServiceV2) SubmitBid(ctx context.Context, env communication.Envelope, horizon Horizon) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := env.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidBid, err)
	}

	s.mu.RLock()
	if s.closedV2 {
		s.mu.RUnlock()
		return ErrExecutiveV2Closed
	}
	s.mu.RUnlock()

	return s.ws.StorePendingCandidate(ctx, env.Topic, workspace.PendingCandidate{
		Envelope:    env,
		Horizon:     int(horizon),
		SubmittedAt: time.Now(),
	})
}

// RemainingBudgetUnits returns available computational cost units.
func (s *ServiceV2) RemainingBudgetUnits() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.budget
}

// ConsumeBudget deducts computational cost units from the current cycle budget.
func (s *ServiceV2) ConsumeBudget(units int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closedV2 {
		return ErrExecutiveV2Closed
	}
	if units < 0 {
		return nil
	}
	if s.budget < units {
		return ErrBudgetExhausted
	}
	s.budget -= units
	return nil
}

// ArbitrateCompetition evaluates pending bids on a leveled topic channel using Calibrated Effective Priority.
// Executive Functions remains content-blind and never inspects PayloadRef contents.
// Pending candidate bids are stored and requested from Workspace.
func (s *ServiceV2) ArbitrateCompetition(ctx context.Context, topic communication.TopicID, admissionThreshold float64) (ArbiterDecision, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return ArbiterDecision{}, err
	}

	s.mu.RLock()
	if s.closedV2 {
		s.mu.RUnlock()
		return ArbiterDecision{}, ErrExecutiveV2Closed
	}
	s.mu.RUnlock()

	pending := s.ws.GetPendingCandidates(topic)
	if len(pending) == 0 {
		// Emit SOAR-style Impasse event
		_ = s.emitImpasse(ctx, topic, "no candidate bids received on topic")
		return ArbiterDecision{
			Admitted:       false,
			ImpasseEmitted: true,
			Reason:         "no candidate bids received on topic",
		}, nil
	}

	var winnerIdx = -1
	var bestPeff = -1.0
	s.mu.RLock()
	currentBudget := s.budget
	s.mu.RUnlock()

	for i, bid := range pending {
		peff := s.cal.CalibrateEnvelope(bid.Envelope, s.alpha, s.beta, currentBudget)
		if peff > bestPeff {
			bestPeff = peff
			winnerIdx = i
		}
	}

	if winnerIdx == -1 || bestPeff < admissionThreshold {
		_ = s.emitImpasse(ctx, topic, fmt.Sprintf("best calibrated priority %.2f below threshold %.2f", bestPeff, admissionThreshold))
		return ArbiterDecision{
			Admitted:       false,
			ImpasseEmitted: true,
			Reason:         fmt.Sprintf("best calibrated priority %.2f below threshold %.2f", bestPeff, admissionThreshold),
		}, nil
	}

	winnerCandidate := pending[winnerIdx]
	winner := winnerCandidate.Envelope

	// Check computational budget
	s.mu.Lock()
	if s.budget < winner.CostEstimateUnits {
		s.mu.Unlock()
		_ = s.emitImpasse(ctx, topic, fmt.Sprintf("budget exhausted for candidate cost %d", winner.CostEstimateUnits))
		return ArbiterDecision{
			Admitted:       false,
			ImpasseEmitted: true,
			Reason:         "insufficient budget for winning bid",
		}, nil
	}

	// Deduct budget
	s.budget -= winner.CostEstimateUnits
	s.mu.Unlock()

	// Request Workspace to remove winning bid from pending list
	_ = s.ws.RemovePendingCandidate(topic, winner.ID)

	// Publish winner to Global Workspace or route through Constitutional Action Gate
	if winner.Topic == communication.TopicActionExecution {
		res, err := s.constGate.InterceptAndPublish(ctx, winner, s.ws)
		if err != nil {
			return ArbiterDecision{
				Admitted:          false,
				Winner:            winner,
				EffectivePriority: bestPeff,
				Reason:            fmt.Sprintf("action gate intercepted: %v", err),
			}, err
		}
		return ArbiterDecision{
			Admitted:          true,
			Winner:            winner,
			EffectivePriority: bestPeff,
			Reason:            fmt.Sprintf("action approved with verdict %s", res.Verdict),
		}, nil
	}

	if err := s.ws.Publish(ctx, winner); err != nil {
		return ArbiterDecision{}, err
	}

	return ArbiterDecision{
		Admitted:          true,
		Winner:            winner,
		EffectivePriority: bestPeff,
		Reason:            "admitted to global workspace",
	}, nil
}

// emitImpasse publishes a content-blind impasse alert to TopicImpasses.
func (s *ServiceV2) emitImpasse(ctx context.Context, failedTopic communication.TopicID, reason string) error {
	impasseEnv, err := communication.NewEnvelopeBuilder().
		WithSource("Intelligence.Executive").
		WithTopic(communication.TopicImpasses).
		WithPayloadRef("storage://impasses/" + string(failedTopic)).
		WithModality("impasse-alert").
		WithConfidence(1.0).
		WithUrgency(50).
		Build()
	if err != nil {
		return err
	}
	return s.ws.Publish(ctx, impasseEnv)
}

// Ensure ServiceV2 implements ExecutiveV2.
var _ ExecutiveV2 = (*ServiceV2)(nil)
