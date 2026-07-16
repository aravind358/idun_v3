package kernel

import "context"

// Lifecycle defines the optional contract that kernel components or registered services
// may satisfy to participate in startup and shutdown orchestration.
//
// The Kernel detects implementations of this interface during Boot and Shutdown.
// Components without this interface continue to operate without lifecycle management.
type Lifecycle interface {
	// Start boots the component's internal workers, timers, or processing loops.
	Start(ctx context.Context) error

	// Close gracefully shuts down the component and releases internal resources.
	Close() error
}

// Phase represents the topological initialization tier of a component.
//
// Startup sequence executes Phase 1 through Phase 6 sequentially.
// Shutdown sequence executes Phase 6 through Phase 1 in exact reverse order.
type Phase int

const (
	// PhaseCore (Phase 1) is for foundational Core Services (Memory, Storage, Logger, Scheduler).
	PhaseCore Phase = iota + 1

	// PhaseInfrastructure (Phase 2) is for Kernel routing, Boundary validation, and Permission rules.
	PhaseInfrastructure

	// PhaseWorkspace (Phase 3) is for the Global Workspace and Leveled Blackboard channels.
	PhaseWorkspace

	// PhaseExecutive (Phase 4) is for Executive Functions, Attention Gate, and Action Gate.
	PhaseExecutive

	// PhaseCognitive (Phase 5) is for primary cognitive abilities (Understanding, Reasoning, Planning, Decision).
	PhaseCognitive

	// PhaseBackground (Phase 6) is for asynchronous and adaptation loops (Reflection, Learning).
	PhaseBackground
)

// Phased defines an optional contract for components or wrappers to declare their boot phase.
// If a Lifecycle component does not implement Phased, the Kernel assigns a safe default phase.
type Phased interface {
	BootPhase() Phase
}
