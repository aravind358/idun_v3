// Package executive implements IDUN's Intelligence Pillar Executive Functions.
//
// As frozen in Architecture Specification v1.0.0-FROZEN, Executive Functions
// coordinates the seven immutable cognitive abilities:
// Understanding, Reasoning, Decision, Planning, Learning, Reflection, and Value.
//
// The Executive Functions module strictly adheres to the Single Responsibility
// Principle and Anti-God-Object constraints: it coordinates cognitive workflows,
// attentional salience, priority preemption, budget allocation, and cognitive
// homeostasis, but never performs domain thinking or duplicates OS kernel primitives.
package executive

import (
	"errors"
	"time"
)

// ============================================================================
// Fixed Cognitive Abilities
// ============================================================================

// CognitiveAbility identifies one of the seven immutable cognitive abilities
// within the IDUN Intelligence Pillar.
type CognitiveAbility string

const (
	// AbilityUnderstanding parses perceptual inputs, semantic decoding, and intent resolution.
	AbilityUnderstanding CognitiveAbility = "Understanding"

	// AbilityReasoning performs logical inference, deductive/inductive reasoning, and causal modeling.
	AbilityReasoning CognitiveAbility = "Reasoning"

	// AbilityDecision selects actions, evaluates trade-offs, and optimizes utility under constraints.
	AbilityDecision CognitiveAbility = "Decision"

	// AbilityPlanning decomposes complex goals into hierarchical task networks and schedules subtasks.
	AbilityPlanning CognitiveAbility = "Planning"

	// AbilityLearning synthesizes long-term patterns, consolidates experiences, and updates skills.
	AbilityLearning CognitiveAbility = "Learning"

	// AbilityReflection performs metacognition, self-critique, error analysis, and contradiction auditing.
	AbilityReflection CognitiveAbility = "Reflection"

	// AbilityValue checks constitutional alignment, ethical boundaries, and normative invariants.
	AbilityValue CognitiveAbility = "Value"
)

// ============================================================================
// Epistemic Status & Budgeting Enums
// ============================================================================

// EpistemicStatus represents the actionable discrete epistemic state returned
// by a cognitive ability upon completing a workflow step.
type EpistemicStatus int

const (
	// StatusConfident indicates the ability successfully solved the step with high certainty.
	StatusConfident EpistemicStatus = iota

	// StatusUnsureAmbiguous indicates input ambiguity requiring Understanding or user clarification.
	StatusUnsureAmbiguous

	// StatusUnsureConflicting indicates logical contradiction requiring Reflection.
	StatusUnsureConflicting

	// StatusUnsureConstitutionalRisk indicates potential ethical/safety conflict requiring Value check.
	StatusUnsureConstitutionalRisk

	// StatusInsufficientData indicates missing domain knowledge requiring Learning or Memory retrieval.
	StatusInsufficientData

	// StatusEscalationRequired indicates the ability cannot solve the task within the assigned budget.
	StatusEscalationRequired

	// StatusUnresolvableContradiction indicates terminal failure after maximum reflection or budget exhaustion.
	StatusUnresolvableContradiction
)

// String returns the canonical human-readable string representation of EpistemicStatus.
func (e EpistemicStatus) String() string {
	switch e {
	case StatusConfident:
		return "CONFIDENT"
	case StatusUnsureAmbiguous:
		return "UNSURE_AMBIGUOUS"
	case StatusUnsureConflicting:
		return "UNSURE_CONFLICTING"
	case StatusUnsureConstitutionalRisk:
		return "UNSURE_CONSTITUTIONAL_RISK"
	case StatusInsufficientData:
		return "INSUFFICIENT_DATA"
	case StatusEscalationRequired:
		return "STATUS_ESCALATION_REQUIRED"
	case StatusUnresolvableContradiction:
		return "STATUS_UNRESOLVABLE_CONTRADICTION"
	default:
		return "UNKNOWN"
	}
}

// BudgetTier defines the computational effort and latency budget allocated to a workflow step.
type BudgetTier int

const (
	// BudgetReflexive allocates minimal execution fuel (<15ms SLA, max 2 transitions) for simple tasks.
	BudgetReflexive BudgetTier = iota

	// BudgetStandard allocates standard execution fuel (<250ms SLA, max 8 transitions) for interactive tasks.
	BudgetStandard

	// BudgetDeliberative allocates deep deliberative fuel (asynchronous, max 32 transitions) for complex goals.
	BudgetDeliberative
)

// String returns the string representation of BudgetTier.
func (b BudgetTier) String() string {
	switch b {
	case BudgetReflexive:
		return "REFLEXIVE"
	case BudgetStandard:
		return "STANDARD"
	case BudgetDeliberative:
		return "DELIBERATIVE"
	default:
		return "UNKNOWN"
	}
}

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

// ============================================================================
// Sentinel Structured Errors
// ============================================================================

var (
	// ErrAbilityNotRegistered indicates that a requested cognitive ability driver is not registered.
	ErrAbilityNotRegistered = errors.New("executive: cognitive ability driver not registered")

	// ErrMaxFuelExceeded indicates that a workflow exceeded its maximum allowed transition fuel.
	ErrMaxFuelExceeded = errors.New("executive: workflow exceeded maximum execution fuel")

	// ErrMaxReflectionExceeded indicates that a workflow exceeded its maximum allowed reflection recursion depth.
	ErrMaxReflectionExceeded = errors.New("executive: maximum reflection depth exceeded")

	// ErrConstitutionalVeto indicates that a workflow was aborted by CognitiveAbility.Value safety check.
	ErrConstitutionalVeto = errors.New("executive: workflow aborted by constitutional value check")

	// ErrWorkflowCancelled indicates that a workflow execution was cancelled.
	ErrWorkflowCancelled = errors.New("executive: workflow cancelled")

	// ErrBudgetExhausted indicates that budget escalation requests could not be satisfied.
	ErrBudgetExhausted = errors.New("executive: budget escalation exhausted")

	// ErrServiceClosed indicates that an operation was attempted on a closed ExecutiveService.
	ErrServiceClosed = errors.New("executive: service closed")
)

// ============================================================================
// Core Domain Data Structures
// ============================================================================

// ActiveGoalContext holds a lightweight reference header to the currently active long-term goal.
// Goal management belongs to Planning and Memory; Executive holds this header solely for salience evaluation.
type ActiveGoalContext struct {
	ID             string
	Summary        string
	PriorityWeight int
}

// Stimulus represents an incoming perceptual event, user prompt, or internal system alert.
type Stimulus struct {
	ID            string
	Source        string
	PayloadRef    string // Immutable storage reference URI (e.g. SHA-256 hash in Core.Storage)
	SafetyFlag    bool   // True if triggered by hardware fault or constitutional safety tripwire
	SalienceScore int    // 0..100 salience score
}

// WorkflowNode represents a single abstract transition step in a cognitive workflow graph.
type WorkflowNode struct {
	ID         string
	Ability    CognitiveAbility
	InputRef   string
	NextNodeID string
}

// WorkflowGraph represents a bounded, acyclic or controlled-recurrence cognitive execution plan.
type WorkflowGraph struct {
	ID              string
	Nodes           map[string]*WorkflowNode
	StartNodeID     string
	GoalContext     ActiveGoalContext
	MaxFuel         int
	ReflectionDepth int
	MaxReflection   int
	Priority        PriorityBand
	Budget          BudgetTier
}

// ExecutionResult summarizes the final outcome of a cognitive workflow graph execution.
type ExecutionResult struct {
	WorkflowID       string
	FinalStatus      EpistemicStatus
	OutputRef        string
	Error            error
	TransitionsCount int
	Duration         time.Duration
}

// WorkflowCheckpoint captures intermediate state during a preemptive interruption.
type WorkflowCheckpoint struct {
	WorkflowID      string
	CurrentNodeID   string
	PayloadRef      string
	RemainingFuel   int
	ReflectionDepth int
	Budget          BudgetTier
	Priority        PriorityBand
}
