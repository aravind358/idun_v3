package capabilities

import (
	"errors"
	"fmt"
	"sync"
)

// DefaultLifecycleManager enforces strict state transitions.
type DefaultLifecycleManager struct {
	mu       sync.RWMutex
	registry CapabilityRegistry
}

func NewLifecycleManager(registry CapabilityRegistry) *DefaultLifecycleManager {
	return &DefaultLifecycleManager{
		registry: registry,
	}
}

// Transition changes the lifecycle state if the transition is valid.
func (m *DefaultLifecycleManager) Transition(capabilityID string, targetState LifecycleState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cap, exists := m.registry.Get(capabilityID)
	if !exists {
		return fmt.Errorf("capability not found: %s", capabilityID)
	}

	currentState := cap.State().Lifecycle
	if !isValidTransition(currentState, targetState) {
		return fmt.Errorf("invalid capability lifecycle transition: %s -> %s", currentState, targetState)
	}

	// In a real implementation, we would mutate the capability state here.
	// However, our Capability interface returns state by value.
	// For V1, we assume the capability itself manages its internal state via a SetState method,
	// or we define a StateMutator interface.
	if mut, ok := cap.(StateMutator); ok {
		mut.SetLifecycleState(targetState)
		return nil
	}

	return errors.New("capability does not support state mutation")
}

func isValidTransition(from, to LifecycleState) bool {
	// A basic state machine definition.
	switch from {
	case LifecycleDiscovered:
		return to == LifecycleRegistered || to == LifecycleDisabled
	case LifecycleRegistered:
		return to == LifecycleLoaded || to == LifecycleDisabled
	case LifecycleLoaded:
		return to == LifecycleInitialized || to == LifecycleUnloaded
	case LifecycleInitialized:
		return to == LifecycleHealthy || to == LifecycleUnloaded
	case LifecycleHealthy:
		return to == LifecycleExecuting || to == LifecycleIdle || to == LifecycleDisabled
	case LifecycleExecuting:
		return to == LifecycleHealthy || to == LifecycleIdle || to == LifecycleDisabled
	case LifecycleIdle:
		return to == LifecycleHealthy || to == LifecycleDisabled || to == LifecycleExecuting
	case LifecycleDisabled:
		return to == LifecycleLoaded || to == LifecycleUnloaded || to == LifecycleRegistered
	case LifecycleUnloaded:
		return to == LifecycleDiscovered
	case "": // Initial creation
		return to == LifecycleDiscovered
	default:
		return false
	}
}

// StateMutator allows the lifecycle manager to inject state changes.
type StateMutator interface {
	SetLifecycleState(state LifecycleState)
	SetOperationalStatus(status OperationalStatus)
}
