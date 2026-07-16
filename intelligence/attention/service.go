package attention

import (
	"context"
	"fmt"
	"sync"
	"time"

	"idun/core/logger"
)

// Service is the concrete, thread-safe implementation of the Attention Gate and GateV2 interfaces.
// It owns focus selection, salience evaluation, and interrupt triage without performing
// domain reasoning, planning, or execution orchestration.
type Service struct {
	mu                 sync.RWMutex
	activeGoal         ActiveGoalContext
	currentFocus       string
	log                logger.Writer
	policyHolder       *PolicySnapshotHolder
	capabilitiesHolder *CapabilitiesSnapshotHolder
	summary            *AttentionSummary
	focusHistory       []FocusHistoryEntry
	eventSummary       *AttentionEventSummary
	publisher          WorkspacePublisher
	payloadStorer      PayloadStorer
	replaySeed         int64
	closed             bool
}

// Option configures functional options for Attention Service construction.
type Option func(*Service)

// WithLogger configures a custom logger for the Attention Service.
func WithLogger(log logger.Writer) Option {
	return func(s *Service) {
		s.log = log
	}
}

// WithPolicyProfile injects a custom AttentionPolicyProfile during construction.
func WithPolicyProfile(profile *AttentionPolicyProfile) Option {
	return func(s *Service) {
		if profile != nil {
			s.policyHolder.Store(profile)
		}
	}
}

// WithCapabilities injects custom AttentionCapabilities during construction.
func WithCapabilities(caps *AttentionCapabilities) Option {
	return func(s *Service) {
		if caps != nil {
			s.capabilitiesHolder.Store(caps)
		}
	}
}

// WithWorkspacePublisher sets the Workspace event publisher and CAS storer.
func WithWorkspacePublisher(pub WorkspacePublisher, storer PayloadStorer) Option {
	return func(s *Service) {
		s.publisher = pub
		s.payloadStorer = storer
	}
}

// WithReplaySeed configures the deterministic replay seed.
func WithReplaySeed(seed int64) Option {
	return func(s *Service) {
		s.replaySeed = seed
	}
}

// NewService constructs a new Attention Service with default snapshots and bounded buffers.
func NewService(opts ...Option) *Service {
	s := &Service{
		policyHolder:       NewPolicySnapshotHolder(DefaultAttentionPolicyProfile()),
		capabilitiesHolder: NewCapabilitiesSnapshotHolder(DefaultAttentionCapabilities()),
		summary:            &AttentionSummary{},
		focusHistory:       make([]FocusHistoryEntry, 0, 16),
		eventSummary:       &AttentionEventSummary{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// NewServiceFromConfig constructs an Attention Service from a validated Configuration struct.
func NewServiceFromConfig(cfg *Configuration) *Service {
	if cfg == nil {
		cfg = DefaultConfiguration()
	}
	svc := NewService(
		WithLogger(cfg.Logger),
		WithPolicyProfile(cfg.PolicyProfile),
		WithCapabilities(cfg.Capabilities),
		WithWorkspacePublisher(cfg.WorkspacePublisher, nil),
		WithReplaySeed(cfg.ReplaySeed),
	)
	return svc
}

// Name returns the canonical Kernel Component name ("Intelligence.Attention").
func (s *Service) Name() string {
	return "Intelligence.Attention"
}

// Start boots the Attention Service lifecycle.
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrServiceClosed
	}
	return nil
}

// Close gracefully shuts down the Attention Service.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return nil
}

// Evaluate inspects a Stimulus against current ActiveGoalContext and assigns triage salience and priority band.
// This method implements the V1 Gate interface while executing full V2 diagnostic tracking internally.
func (s *Service) Evaluate(stim Stimulus) (SalienceDecision, PriorityBand) {
	trace, err := s.EvaluateTrace(context.Background(), stim)
	if err != nil || trace == nil {
		// Fallback for closed service or validation failure matching heritage contract
		if stim.SafetyFlag {
			return SalienceFocusImmediately, PriorityBand0CriticalSafety
		}
		if stim.SalienceScore >= 85 {
			return SalienceFocusImmediately, PriorityBand1RealTime
		} else if stim.SalienceScore >= 50 {
			return SalienceFocusImmediately, PriorityBand2Interactive
		} else if stim.SalienceScore >= 20 {
			return SalienceSchedule, PriorityBand3Background
		}
		return SalienceFilter, PriorityBand4Idle
	}
	return trace.Decision, trace.PriorityBand
}

// EvaluateTrace inspects a Stimulus against active goal context, applies fingerprinted salience policies,
// updates bounded rolling summaries/histories, publishes observational events, and returns a diagnostic AttentionTrace.
func (s *Service) EvaluateTrace(ctx context.Context, stim Stimulus) (*AttentionTrace, error) {
	startTime := time.Now()

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, ErrServiceClosed
	}
	s.summary.TotalStimuli++
	focusBefore := s.currentFocus
	publisher := s.publisher
	storer := s.payloadStorer
	seed := s.replaySeed
	s.mu.Unlock()

	// Validate stimulus
	if err := stim.Validate(); err != nil {
		s.mu.Lock()
		s.summary.FilteredCount++
		s.mu.Unlock()
		return nil, err
	}

	// Load lock-free atomic snapshots
	profile := s.policyHolder.Load()
	caps := s.capabilitiesHolder.Load()

	// Determine decision and band
	var dec SalienceDecision
	var band PriorityBand
	var reason AttentionTerminationReason
	var status AttentionResultStatus

	if stim.SafetyFlag || stim.SalienceScore >= profile.Band0Threshold {
		dec = SalienceFocusImmediately
		band = PriorityBand0CriticalSafety
		reason = ReasonSafetyTripwire
		status = ResultStatusFocused
	} else if stim.SalienceScore >= profile.Band1Threshold {
		dec = SalienceFocusImmediately
		band = PriorityBand1RealTime
		reason = ReasonHighSalience
		status = ResultStatusFocused
	} else if stim.SalienceScore >= profile.Band2Threshold {
		dec = SalienceFocusImmediately
		band = PriorityBand2Interactive
		reason = ReasonInteractiveSalience
		status = ResultStatusFocused
	} else if stim.SalienceScore >= profile.Band3Threshold {
		dec = SalienceSchedule
		band = PriorityBand3Background
		reason = ReasonBackgroundSalience
		status = ResultStatusScheduled
	} else {
		dec = SalienceFilter
		band = PriorityBand4Idle
		reason = ReasonLowSalience
		status = ResultStatusFiltered
	}

	// Determine if focus switch occurred and handle interruptions
	switchOccurred := false
	s.mu.Lock()
	if dec == SalienceFocusImmediately {
		s.summary.ImmediateFocusCount++
		if s.currentFocus != stim.ID {
			switchOccurred = true
			prevFocus := s.currentFocus
			s.currentFocus = stim.ID
			s.summary.FocusSwitches++
			s.eventSummary.FocusChangedCount++

			// Append to bounded rolling focus history (max 16 entries)
			if caps.SupportsFocusHistory && profile.MaximumTrackedStimuli > 0 {
				entry := FocusHistoryEntry{
					PreviousFocus: prevFocus,
					CurrentFocus:  stim.ID,
					SwitchReason:  string(reason),
					Timestamp:     time.Now().UTC(),
				}
				s.focusHistory = append(s.focusHistory, entry)
				if len(s.focusHistory) > 16 {
					s.focusHistory = s.focusHistory[len(s.focusHistory)-16:]
				}
			}
		}
		if band <= PriorityBand1RealTime && switchOccurred && caps.SupportsInterruptions {
			s.summary.InterruptAccepted++
			s.eventSummary.InterruptAcceptedCount++
		}
	} else if dec == SalienceSchedule {
		s.summary.ScheduledCount++
		if stim.SalienceScore >= profile.Band3Threshold && stim.SalienceScore < profile.Band2Threshold && s.currentFocus != "" {
			s.summary.InterruptRejected++
			s.eventSummary.InterruptRejectedCount++
		}
	} else {
		s.summary.FilteredCount++
	}
	if stim.SafetyFlag {
		s.eventSummary.SafetyTripwireCount++
	}

	execDuration := time.Since(startTime)
	s.summary.TotalEvaluationTime += execDuration
	if s.summary.TotalStimuli > 0 {
		s.summary.AverageEvaluationTime = time.Duration(int64(s.summary.TotalEvaluationTime) / s.summary.TotalStimuli)
	}
	focusAfter := s.currentFocus
	s.mu.Unlock()

	// Build immutable diagnostic trace
	trace := &AttentionTrace{
		TraceID:               fmt.Sprintf("trace-att-%d", time.Now().UnixNano()),
		StimulusID:            stim.ID,
		StimulusSource:        stim.Source,
		FocusBefore:           focusBefore,
		FocusAfter:            focusAfter,
		Decision:              dec,
		PriorityBand:          band,
		SwitchOccurred:        switchOccurred,
		ExecutionTime:         execDuration,
		PolicyFingerprint:     profile.PolicyFingerprint,
		CapabilityFingerprint: caps.CapabilityFingerprint,
		AttentionVersion:      AttentionVersion,
		ReplayMetadata: AttentionReplayMetadata{
			PolicyFingerprint:     profile.PolicyFingerprint,
			CapabilityFingerprint: caps.CapabilityFingerprint,
			AttentionVersion:      AttentionVersion,
			ReplaySeed:            seed,
		},
		ResultStatus:      status,
		TerminationReason: reason,
		EvaluatedAt:       time.Now().UTC(),
	}

	if err := trace.Validate(); err != nil {
		return nil, err
	}

	// Publish Workspace observational events asynchronously / non-blocking if configured
	if publisher != nil && storer != nil {
		if switchOccurred {
			if env, err := EnvelopeFromFocusChange(ctx, storer, FocusChangedPayload{
				PreviousFocus:     focusBefore,
				CurrentFocus:      focusAfter,
				SwitchReason:      string(reason),
				PolicyFingerprint: profile.PolicyFingerprint,
				Timestamp:         trace.EvaluatedAt,
			}); err == nil {
				_ = publisher.Publish(ctx, env)
			}
		}
		if band <= PriorityBand1RealTime && switchOccurred && caps.SupportsInterruptions {
			if env, err := EnvelopeFromInterrupt(ctx, storer, InterruptPayload{
				StimulusID:        stim.ID,
				StimulusSource:    stim.Source,
				Accepted:          true,
				AssignedBand:      band,
				Decision:          dec,
				PolicyFingerprint: profile.PolicyFingerprint,
				Timestamp:         trace.EvaluatedAt,
			}); err == nil {
				_ = publisher.Publish(ctx, env)
			}
		}
	}

	return trace, nil
}

// SetActiveGoal updates the lightweight active goal header reference.
func (s *Service) SetActiveGoal(goal ActiveGoalContext) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeGoal = goal
	s.eventSummary.GoalSwitchCount++
	if s.log != nil {
		s.log.Info("Active goal updated in Attention subsystem",
			logger.Field{Key: "goal_id", Value: goal.ID},
			logger.Field{Key: "summary", Value: goal.Summary},
		)
	}
}

// GetActiveGoal returns the currently active goal header reference.
func (s *Service) GetActiveGoal() ActiveGoalContext {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeGoal
}

// GetPolicyProfile returns the active AttentionPolicyProfile snapshot.
func (s *Service) GetPolicyProfile() *AttentionPolicyProfile {
	return s.policyHolder.Load()
}

// GetCapabilities returns the advertised AttentionCapabilities snapshot.
func (s *Service) GetCapabilities() *AttentionCapabilities {
	return s.capabilitiesHolder.Load()
}

// GetSummary returns a snapshot copy of the bounded statistical telemetry.
func (s *Service) GetSummary() AttentionSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.summary
}

// GetFocusHistory returns a snapshot copy of the bounded rolling focus transitions (max 16 entries).
func (s *Service) GetFocusHistory() []FocusHistoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]FocusHistoryEntry, len(s.focusHistory))
	copy(out, s.focusHistory)
	return out
}

// GetEventSummary returns a snapshot copy of the bounded event summary counters.
func (s *Service) GetEventSummary() AttentionEventSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.eventSummary
}

// SetWorkspacePublisher injects or updates the Workspace publisher and CAS storer.
func (s *Service) SetWorkspacePublisher(pub WorkspacePublisher, storer PayloadStorer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.publisher = pub
	s.payloadStorer = storer
}

// Ensure Service implements Gate, GateV2, and AttentionService.
var (
	_ Gate             = (*Service)(nil)
	_ GateV2           = (*Service)(nil)
	_ AttentionService = (*Service)(nil)
)
