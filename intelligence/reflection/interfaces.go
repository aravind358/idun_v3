package reflection

import (
	"context"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
)

// ReflectionService defines the public API for IDUN's Reflection cognitive ability.
// Reflection is strictly read-only, non-blocking optimization infrastructure that produces
// structured raw material for Learning.
type ReflectionService interface {
	executive.AbilityDriver

	// ReflectEpisode evaluates a single completed cognitive episode from read-only Workspace traces.
	ReflectEpisode(ctx context.Context, episodeID string, traces []communication.Envelope) (ReflectionReport, error)

	// ReflectPeriodic evaluates longitudinal behavioral trends from read-only HistoricalSummary contracts.
	ReflectPeriodic(ctx context.Context, summary HistoricalSummary) (ReflectionReport, error)

	// Start boots the Reflection Service.
	Start() error

	// Close shuts down the Reflection Service cleanly.
	Close() error
}

// EvaluationStrategy represents a swappable evaluation engine behind any Reflection specialist.
// Strategies start as heuristic rules and evolve over decades into statistical or learned models.
type EvaluationStrategy interface {
	// Evaluate assesses execution traces and emits a SpecialistReport.
	Evaluate(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error)

	// StrategyID returns the unique canonical identifier for this strategy.
	StrategyID() string

	// Version returns the version string of this strategy.
	Version() string
}

// SpecialistEvaluator defines a cognitive specialist evaluator dedicated to reviewing a specific cognitive ability.
type SpecialistEvaluator interface {
	// ID returns the specialist identifier.
	ID() string

	// TargetAbility identifies which cognitive ability this specialist evaluates.
	TargetAbility() executive.CognitiveAbility

	// EvaluateEpisode evaluates execution traces for the target ability during an episode.
	EvaluateEpisode(ctx context.Context, traces []communication.Envelope) (SpecialistReport, error)
}
