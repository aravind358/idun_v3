package attention

import (
	"errors"
	"time"
)

// AttentionVersion is the canonical implementation version of the Attention subsystem.
const AttentionVersion = "2.0.0-FROZEN"

// Sentinel errors for Attention domain artifacts and firewalls.
var (
	ErrNilProfile            = errors.New("attention: AttentionPolicyProfile cannot be nil")
	ErrNilCapabilities       = errors.New("attention: AttentionCapabilities cannot be nil")
	ErrNilTrace              = errors.New("attention: AttentionTrace cannot be nil")
	ErrNilSummary            = errors.New("attention: AttentionSummary cannot be nil")
	ErrMissingTraceID        = errors.New("attention: TraceID is required")
	ErrMissingStimulusID     = errors.New("attention: Stimulus.ID is required")
	ErrInvalidSalienceScore  = errors.New("attention: Stimulus.SalienceScore must be between 0 and 100")
	ErrInvalidThreshold      = errors.New("attention: policy threshold must be between 0 and 100")
	ErrInvalidMargin         = errors.New("attention: policy switch margin cannot be negative")
	ErrNegativeTrackedLimit  = errors.New("attention: MaximumTrackedStimuli cannot be negative")
	ErrMissingPolicyVersion  = errors.New("attention: PolicyVersion is required")
	ErrMissingSchemaVersion  = errors.New("attention: SchemaVersion is required")
	ErrMissingFingerprint    = errors.New("attention: PolicyFingerprint or CapabilityFingerprint is required")
	ErrInvalidResultStatus   = errors.New("attention: invalid AttentionResultStatus")
	ErrInvalidReason         = errors.New("attention: invalid AttentionTerminationReason")
	ErrServiceClosed         = errors.New("attention: service is closed")
)

// AttentionResultStatus separates lifecycle status from termination reason.
type AttentionResultStatus string

const (
	ResultStatusFocused          AttentionResultStatus = "FOCUSED"
	ResultStatusScheduled        AttentionResultStatus = "SCHEDULED"
	ResultStatusFiltered         AttentionResultStatus = "FILTERED"
	ResultStatusPreempted        AttentionResultStatus = "PREEMPTED"
	ResultStatusEvaluationFailed AttentionResultStatus = "EVALUATION_FAILED"
)

// IsValid returns true if r is a known AttentionResultStatus.
func (r AttentionResultStatus) IsValid() bool {
	switch r {
	case ResultStatusFocused, ResultStatusScheduled, ResultStatusFiltered, ResultStatusPreempted, ResultStatusEvaluationFailed:
		return true
	default:
		return false
	}
}

// AttentionTerminationReason explains the factual cause of an attentional gating outcome.
type AttentionTerminationReason string

const (
	ReasonSafetyTripwire         AttentionTerminationReason = "SAFETY"
	ReasonHighSalience           AttentionTerminationReason = "HIGH_SALIENCE"
	ReasonInteractiveSalience    AttentionTerminationReason = "INTERACTIVE_SALIENCE"
	ReasonBackgroundSalience     AttentionTerminationReason = "BACKGROUND_SALIENCE"
	ReasonLowSalience            AttentionTerminationReason = "LOW_SALIENCE"
	ReasonConstitutionalVeto     AttentionTerminationReason = "CONSTITUTION"
	ReasonUserInterrupt          AttentionTerminationReason = "USER_INTERRUPT"
)

// IsValid returns true if r is a known AttentionTerminationReason.
func (r AttentionTerminationReason) IsValid() bool {
	switch r {
	case ReasonSafetyTripwire, ReasonHighSalience, ReasonInteractiveSalience, ReasonBackgroundSalience, ReasonLowSalience, ReasonConstitutionalVeto, ReasonUserInterrupt:
		return true
	default:
		return false
	}
}

// Validate checks Stimulus fields for structural validity.
func (s Stimulus) Validate() error {
	if s.ID == "" {
		return ErrMissingStimulusID
	}
	if s.SalienceScore < 0 || s.SalienceScore > 100 {
		return ErrInvalidSalienceScore
	}
	return nil
}

// Validate checks ActiveGoalContext fields for structural validity.
func (g ActiveGoalContext) Validate() error {
	if g.PriorityWeight < 0 {
		return errors.New("attention: ActiveGoalContext.PriorityWeight cannot be negative")
	}
	return nil
}

// AttentionCapabilities advertises the deployment-specific capabilities of the Attention engine.
type AttentionCapabilities struct {
	SupportsInterruptions       bool   `json:"supports_interruptions"`
	SupportsBackgroundAttention bool   `json:"supports_background_attention"`
	SupportsFocusSwitching      bool   `json:"supports_focus_switching"`
	SupportsMultimodalAttention bool   `json:"supports_multimodal_attention"`
	SupportsDistributedAttention bool  `json:"supports_distributed_attention"`
	SupportsFocusHistory        bool   `json:"supports_focus_history"`
	CapabilityFingerprint       string `json:"capability_fingerprint"`
}

// Validate verifies that the AttentionCapabilities struct is well-formed.
func (c *AttentionCapabilities) Validate() error {
	if c == nil {
		return ErrNilCapabilities
	}
	if c.CapabilityFingerprint == "" {
		return ErrMissingFingerprint
	}
	return nil
}

// AttentionReplayMetadata ensures deterministic replay across identical inputs.
type AttentionReplayMetadata struct {
	PolicyFingerprint     string `json:"policy_fingerprint"`
	CapabilityFingerprint string `json:"capability_fingerprint"`
	AttentionVersion      string `json:"attention_version"`
	ReplaySeed            int64  `json:"replay_seed"`
}

// Validate verifies that the ReplayMetadata is well-formed.
func (r *AttentionReplayMetadata) Validate() error {
	if r == nil {
		return errors.New("attention: AttentionReplayMetadata cannot be nil")
	}
	if r.PolicyFingerprint == "" || r.CapabilityFingerprint == "" {
		return ErrMissingFingerprint
	}
	if r.AttentionVersion == "" {
		return errors.New("attention: AttentionVersion cannot be empty in replay metadata")
	}
	return nil
}

// AttentionTrace records the immutable diagnostic provenance of a single triage evaluation.
// Reflection consumes this trace; Attention never analyzes its own traces.
type AttentionTrace struct {
	TraceID               string                     `json:"trace_id"`
	StimulusID            string                     `json:"stimulus_id"`
	StimulusSource        string                     `json:"stimulus_source"`
	FocusBefore           string                     `json:"focus_before"`
	FocusAfter            string                     `json:"focus_after"`
	Decision              SalienceDecision           `json:"decision"`
	PriorityBand          PriorityBand               `json:"priority_band"`
	SwitchOccurred        bool                       `json:"switch_occurred"`
	ExecutionTime         time.Duration              `json:"execution_time"`
	PolicyFingerprint     string                     `json:"policy_fingerprint"`
	CapabilityFingerprint string                     `json:"capability_fingerprint"`
	AttentionVersion      string                     `json:"attention_version"`
	ReplayMetadata        AttentionReplayMetadata    `json:"replay_metadata"`
	ResultStatus          AttentionResultStatus      `json:"result_status"`
	TerminationReason     AttentionTerminationReason `json:"termination_reason"`
	EvaluatedAt           time.Time                  `json:"evaluated_at"`
}

// Validate verifies that the AttentionTrace satisfies all Layer 1 diagnostic invariants.
func (t *AttentionTrace) Validate() error {
	if t == nil {
		return ErrNilTrace
	}
	if t.TraceID == "" {
		return ErrMissingTraceID
	}
	if t.StimulusID == "" {
		return ErrMissingStimulusID
	}
	if !t.ResultStatus.IsValid() {
		return ErrInvalidResultStatus
	}
	if !t.TerminationReason.IsValid() {
		return ErrInvalidReason
	}
	if t.PolicyFingerprint == "" || t.CapabilityFingerprint == "" {
		return ErrMissingFingerprint
	}
	if err := t.ReplayMetadata.Validate(); err != nil {
		return err
	}
	return nil
}

// AttentionSummary provides bounded statistical telemetry across all triage evaluations.
// Attention records facts; Reflection performs deeper statistical interpretations.
type AttentionSummary struct {
	TotalStimuli          int64         `json:"total_stimuli"`
	ImmediateFocusCount   int64         `json:"immediate_focus_count"`
	ScheduledCount        int64         `json:"scheduled_count"`
	FilteredCount         int64         `json:"filtered_count"`
	InterruptAccepted     int64         `json:"interrupt_accepted"`
	InterruptRejected     int64         `json:"interrupt_rejected"`
	FocusSwitches         int64         `json:"focus_switches"`
	AverageEvaluationTime time.Duration `json:"average_evaluation_time"`
	TotalEvaluationTime   time.Duration `json:"total_evaluation_time"`
}

// Validate verifies the consistency of bounded statistical telemetry.
func (s *AttentionSummary) Validate() error {
	if s == nil {
		return ErrNilSummary
	}
	if s.TotalStimuli < 0 || s.ImmediateFocusCount < 0 || s.ScheduledCount < 0 || s.FilteredCount < 0 {
		return errors.New("attention: summary counts cannot be negative")
	}
	return nil
}

// FocusHistoryEntry represents a single bounded transition in focus state.
type FocusHistoryEntry struct {
	PreviousFocus string    `json:"previous_focus"`
	CurrentFocus  string    `json:"current_focus"`
	SwitchReason  string    `json:"switch_reason"`
	Timestamp     time.Time `json:"timestamp"`
}

// Validate checks FocusHistoryEntry fields.
func (e *FocusHistoryEntry) Validate() error {
	if e == nil {
		return errors.New("attention: FocusHistoryEntry cannot be nil")
	}
	if e.SwitchReason == "" {
		return errors.New("attention: SwitchReason cannot be empty")
	}
	return nil
}

// AttentionEventSummary accumulates bounded counters of published observational events.
type AttentionEventSummary struct {
	FocusChangedCount      int64 `json:"focus_changed_count"`
	InterruptAcceptedCount int64 `json:"interrupt_accepted_count"`
	InterruptRejectedCount int64 `json:"interrupt_rejected_count"`
	GoalSwitchCount        int64 `json:"goal_switch_count"`
	SafetyTripwireCount    int64 `json:"safety_tripwire_count"`
}

// Validate checks AttentionEventSummary fields.
func (e *AttentionEventSummary) Validate() error {
	if e == nil {
		return errors.New("attention: AttentionEventSummary cannot be nil")
	}
	if e.FocusChangedCount < 0 || e.InterruptAcceptedCount < 0 || e.InterruptRejectedCount < 0 || e.GoalSwitchCount < 0 || e.SafetyTripwireCount < 0 {
		return errors.New("attention: event summary counts cannot be negative")
	}
	return nil
}
