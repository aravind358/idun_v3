package capabilities

import (
	"context"
)

// ============================================================================
// Core Capability Interfaces
// ============================================================================

// Capability represents the atomic, stateless bridge between Intelligence and World.
type Capability interface {
	// ID returns the unique identifier for this capability instance.
	ID() string

	// Metadata returns immutable information describing the capability.
	Metadata() CapabilityMetadata

	// State returns the current lifecycle and health state.
	State() CapabilityState

	// Execute performs the mechanical action and returns a normalized result.
	Execute(ctx context.Context, req CapabilityRequest) (CapabilityResult, error)
}

// CapabilityRegistry manages the catalog of discovered and available capabilities.
type CapabilityRegistry interface {
	Register(cap Capability) error
	Deregister(capabilityID string) error
	Get(capabilityID string) (Capability, bool)
	List() []Capability
	FindByName(name string) []Capability
}

// CapabilityResolver maps an abstract requirement to a concrete execution resource.
type CapabilityResolver interface {
	// Resolve selects the best capability for the given abstract requirement.
	Resolve(ctx context.Context, requirementID, capabilityName string, params map[string]string) (Capability, error)
}

// LifecycleManager handles the state transitions of capabilities.
type LifecycleManager interface {
	Transition(capabilityID string, targetState LifecycleState) error
}

// CapabilityManager is the primary facade for the capability framework.
type CapabilityManager interface {
	Registry() CapabilityRegistry
	Resolver() CapabilityResolver
	Lifecycle() LifecycleManager
	Start() error
	Stop() error
}

// ResultNormalizer standardizes raw implementation outputs.
type ResultNormalizer interface {
	Normalize(rawResponse interface{}) (CapabilityResult, error)
}
