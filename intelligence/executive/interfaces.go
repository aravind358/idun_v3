// Package executive defines interfaces and contracts for IDUN Executive Functions.
package executive

import (
	"context"
	"time"

	"idun/intelligence/attention"
)

// ============================================================================
// Executive Sub-Component Capability Interfaces
// ============================================================================

// AttentionGate defines the capability to triage incoming stimuli against active goals.
type AttentionGate = attention.Gate

// PriorityEngine defines the capability to rank and preempt workflows across Priority Bands 0..4.
type PriorityEngine interface {
	// Enqueue schedules a workflow into the specified PriorityBand queue.
	Enqueue(wg *WorkflowGraph) error

	// Dequeue returns the highest priority workflow currently pending execution.
	Dequeue() (*WorkflowGraph, bool)

	// Preempt signals immediate interruption of lower-priority active tasks when Band 0/1 arrives.
	Preempt(ctx context.Context, incomingBand PriorityBand) error
}

// BudgetManager defines the capability to assign and arbitrate computational effort budgets.
type BudgetManager interface {
	// AssignBudget determines the initial BudgetTier for a stimulus or workflow.
	AssignBudget(salience SalienceDecision, priority PriorityBand) BudgetTier

	// EvaluateEscalation decides whether an escalation request from a cognitive ability is granted.
	EvaluateEscalation(current BudgetTier, priority PriorityBand) (BudgetTier, bool)
}

// WorkflowCoordinator defines the capability to execute bounded cognitive workflow graphs.
type WorkflowCoordinator interface {
	// Execute runs a WorkflowGraph through registered cognitive ability drivers until completion or termination.
	Execute(ctx context.Context, wg *WorkflowGraph) ExecutionResult
}

// CancellationCoordinator defines the capability to broadcast cooperative and forcible cancellation signals.
type CancellationCoordinator interface {
	// RegisterTask registers a workflow context for potential preemption or cancellation.
	RegisterTask(workflowID string, cancelFunc context.CancelFunc)

	// CancelTask cancels a specific running workflow by ID.
	CancelTask(workflowID string) error

	// CancelAll cancels all currently active workflows.
	CancelAll()
}

// HomeostasisController defines the capability to sense idle periods and schedule consolidation workflows.
type HomeostasisController interface {
	// RecordActivity records that external stimulus or reactive processing occurred.
	RecordActivity()

	// ShouldConsolidate returns true if the system has been idle longer than the idle threshold.
	ShouldConsolidate() bool

	// IdleDuration returns the duration since last external activity.
	IdleDuration() time.Duration
}

// AbilityRegistry defines the capability to register, inspect, and retrieve cognitive ability drivers.
type AbilityRegistry interface {
	// RegisterDriver registers an AbilityDriver for its declared CognitiveAbility.
	RegisterDriver(driver AbilityDriver) error

	// GetDriver returns the registered driver for a given CognitiveAbility.
	GetDriver(ability CognitiveAbility) (AbilityDriver, error)

	// ListAbilities returns all currently registered cognitive abilities.
	ListAbilities() []CognitiveAbility
}

// Executive defines the composite service capability interface of Executive Functions.
type Executive interface {
	AttentionGate
	PriorityEngine
	BudgetManager
	WorkflowCoordinator
	CancellationCoordinator
	HomeostasisController
	AbilityRegistry

	// Name returns the canonical Kernel Component name ("Intelligence.Executive").
	Name() string

	// Start boots the Executive Service lifecycle.
	Start() error

	// Close shuts down the Executive Service gracefully.
	Close() error
}

// ============================================================================
// Cognitive Ability Driver Contract & Placeholder Interfaces
// ============================================================================

// AbilityDriver defines the version-invariant contract implemented by every cognitive ability.
type AbilityDriver interface {
	// Ability returns the CognitiveAbility identifier that this driver implements.
	Ability() CognitiveAbility

	// ExecuteTask runs a single step of a cognitive workflow and returns an EpistemicStatus and output reference URI.
	ExecuteTask(ctx context.Context, workflowID, nodeID string, budget BudgetTier, payloadRef string) (EpistemicStatus, string, error)
}

// UnderstandingAbility represents the future interface contract for CognitiveAbility.Understanding.
// TODO(intelligence): Implement concrete Understanding service driver.
type UnderstandingAbility interface {
	AbilityDriver
	ParseIntent(ctx context.Context, payloadRef string) (string, error)
}

// ReasoningAbility represents the future interface contract for CognitiveAbility.Reasoning.
// TODO(intelligence): Implement concrete Reasoning service driver.
type ReasoningAbility interface {
	AbilityDriver
	SynthesizeInference(ctx context.Context, premisesRef string) (string, error)
}

// DecisionAbility represents the future interface contract for CognitiveAbility.Decision.
// TODO(intelligence): Implement concrete Decision service driver.
type DecisionAbility interface {
	AbilityDriver
	SelectAction(ctx context.Context, optionsRef string) (string, error)
}

// PlanningAbility represents the future interface contract for CognitiveAbility.Planning.
// TODO(intelligence): Implement concrete Planning service driver.
type PlanningAbility interface {
	AbilityDriver
	DecomposeGoal(ctx context.Context, goalRef string) (string, error)
}

// LearningAbility represents the future interface contract for CognitiveAbility.Learning.
// TODO(intelligence): Implement concrete Learning service driver.
type LearningAbility interface {
	AbilityDriver
	ConsolidateExperience(ctx context.Context, episodicRef string) (string, error)
}

// ReflectionAbility represents the future interface contract for CognitiveAbility.Reflection.
// TODO(intelligence): Implement concrete Reflection service driver.
type ReflectionAbility interface {
	AbilityDriver
	AuditContradiction(ctx context.Context, traceRef string) (string, error)
}

// ValueAbility represents the future interface contract for CognitiveAbility.Value (Constitution).
// TODO(intelligence): Implement concrete Value service driver.
type ValueAbility interface {
	AbilityDriver
	VerifyConstitutionalAlignment(ctx context.Context, proposalRef string) (bool, string, error)
}
