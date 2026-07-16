// Package attention implements IDUN's Attention Subsystem.
//
// Attention owns all triage, prioritization, salience evaluation, focus
// allocation, and interruption routing across cognitive stimuli and goals.
// Executive Functions coordinates with the Attention subsystem without duplicating
// or executing attentional processing internally.
package attention

import "context"

// SalienceDecision represents the outcome of attentional gating evaluation.
type SalienceDecision string

const (
	// SalienceFocusImmediately routes the stimulus into Priority Bands 0..2 for immediate dispatch.
	SalienceFocusImmediately SalienceDecision = "FOCUS_IMMEDIATELY"

	// SalienceSchedule routes the stimulus into Priority Band 3 for deferred scheduling.
	SalienceSchedule SalienceDecision = "SCHEDULE"

	// SalienceFilter drops low-salience sensory flutter without spending cognitive effort.
	SalienceFilter SalienceDecision = "FILTER"
)

// PriorityBand defines the 5-band hierarchy for task prioritization and preemption.
type PriorityBand int

const (
	// PriorityBand0CriticalSafety is non-preemptible and immediately preempts any lower band.
	PriorityBand0CriticalSafety PriorityBand = 0

	// PriorityBand1RealTime is for time-critical synchronous user interactions.
	PriorityBand1RealTime PriorityBand = 1

	// PriorityBand2Interactive is for standard interactive dialogue workflows.
	PriorityBand2Interactive PriorityBand = 2

	// PriorityBand3Background is for scheduled background maintenance and reminders.
	PriorityBand3Background PriorityBand = 3

	// PriorityBand4Idle is for exploratory reflection and memory consolidation during zero salience.
	PriorityBand4Idle PriorityBand = 4
)

// Stimulus represents an incoming perceptual event, user prompt, or internal system alert.
type Stimulus struct {
	ID            string
	Source        string
	PayloadRef    string // Immutable storage reference URI (e.g. SHA-256 hash in Core.Storage)
	SafetyFlag    bool   // True if triggered by hardware fault or constitutional safety tripwire
	SalienceScore int    // 0..100 salience score
}

// ActiveGoalContext holds a lightweight reference header to the currently active long-term goal.
type ActiveGoalContext struct {
	ID             string
	Summary        string
	PriorityWeight int
}

// Gate defines the capability to triage incoming stimuli against active goals.
type Gate interface {
	// Evaluate inspects a Stimulus against current ActiveGoalContext and assigns triage salience and priority band.
	Evaluate(s Stimulus) (SalienceDecision, PriorityBand)

	// SetActiveGoal updates the lightweight active goal reference header.
	SetActiveGoal(goal ActiveGoalContext)

	// GetActiveGoal returns the currently active goal reference header.
	GetActiveGoal() ActiveGoalContext
}

// GateV2 extends Gate with complete Phase 1 Layer 1 capabilities (Version 2.0.0-FROZEN).
type GateV2 interface {
	Gate

	// Start boots the Attention Service lifecycle and initializes snapshots.
	Start(ctx context.Context) error

	// Close gracefully shuts down the Attention Service.
	Close() error

	// EvaluateTrace inspects a Stimulus and returns a full diagnostic AttentionTrace.
	EvaluateTrace(ctx context.Context, s Stimulus) (*AttentionTrace, error)

	// GetPolicyProfile returns the currently active AttentionPolicyProfile.
	GetPolicyProfile() *AttentionPolicyProfile

	// GetCapabilities returns the immutable structural features advertised by this deployment.
	GetCapabilities() *AttentionCapabilities

	// GetSummary returns bounded statistical telemetry over all triage evaluations.
	GetSummary() AttentionSummary

	// GetFocusHistory returns a bounded snapshot (max 16 entries) of focus transitions.
	GetFocusHistory() []FocusHistoryEntry

	// GetEventSummary returns bounded counters of published focus, interrupt, and goal events.
	GetEventSummary() AttentionEventSummary
}

// AttentionService defines the full canonical service interface for CognitiveAbility.Attention.
type AttentionService interface {
	GateV2
	Name() string
}

