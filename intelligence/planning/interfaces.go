package planning

import (
	"context"

	"idun/intelligence/executive"
)

// ============================================================================
// Core Planning Service Interface
// ============================================================================

// PlanningService defines the public contract for CognitiveAbility.Planning.
// It implements executive.PlanningAbility and provides explicit structured
// planning across Reflexive (<10ms), Tactical (10-100ms), and Deliberative (100-500ms)
// computational horizons.
type PlanningService interface {
	executive.PlanningAbility

	// Start initializes the Planning service lifecycle and background workers.
	Start() error

	// Close gracefully shuts down the Planning service.
	Close() error

	// PlanReflexive performs fast cache lookup or memoized template planning (<10ms).
	PlanReflexive(ctx context.Context, req *PlanningRequest) (*PlanningResult, error)

	// PlanTactical performs domain-weighted specialist HTN/GOAP decomposition (10-100ms).
	PlanTactical(ctx context.Context, req *PlanningRequest) (*PlanningResult, error)

	// PlanDeliberative performs wide multi-alternative tree search and contingency generation (100-500ms).
	PlanDeliberative(ctx context.Context, req *PlanningRequest) (*PlanningResult, error)

	// GetTrace retrieves a diagnostic PlanningTrace by its unique trace ID.
	GetTrace(traceID string) (*PlanningTrace, bool)

	// Capabilities returns the immutable structural features advertised by this planning engine deployment.
	Capabilities() *PlanningCapabilities
}

// ============================================================================
// Internal Specialist & Modular Roster Interfaces
// ============================================================================

// PlanningSpecialist defines the internal contract implemented by domain specialists
// (e.g., GoalDecomposition, TaskSequencing, ResourcePlanning, AcquisitionPlanning).
type PlanningSpecialist interface {
	// Name returns the unique canonical identifier for the specialist.
	Name() string

	// SupportedDomains returns the list of open domain tags supported by this specialist.
	SupportedDomains() []string

	// Contribute evaluates the current planning graph and returns new subgoals, edges, and step logs.
	Contribute(
		ctx context.Context,
		req *PlanningRequest,
		currentGraph *DependencyGraphSnapshot,
		profile *PlanningPolicyProfile,
	) (*PlanningStepLog, []Subgoal, []DependencyEdge, error)
}

// PlanFingerprinter computes canonical content-addressed hashes over structural plan data,
// explicitly excluding estimates and numerical metrics to guarantee deduplication accuracy.
type PlanFingerprinter interface {
	// ComputeFingerprint returns the deterministic SHA-256 hex digest for the Plan.
	ComputeFingerprint(plan *Plan) (string, error)
}

// StrategyProvider defines lock-free access to the active PlanningStrategySnapshot.
type StrategyProvider interface {
	// ActiveSnapshot returns the currently active strategy snapshot.
	ActiveSnapshot() *PlanningStrategySnapshot
}
