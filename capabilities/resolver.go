package capabilities

import (
	"context"
	"errors"
	"fmt"
)

// DefaultResolver maps capability requirements to active registry capabilities.
type DefaultResolver struct {
	registry CapabilityRegistry
}

func NewResolver(registry CapabilityRegistry) *DefaultResolver {
	return &DefaultResolver{
		registry: registry,
	}
}

func (r *DefaultResolver) Resolve(ctx context.Context, requirementID, capabilityName string, params map[string]string) (Capability, error) {
	candidates := r.registry.FindByName(capabilityName)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no capability found matching name: %s", capabilityName)
	}

	// In a complete implementation, this would score capabilities by CostEstimate, LatencyMs, Health.
	// For V1, select the first healthy capability.
	for _, cap := range candidates {
		state := cap.State()
		if state.Operational == StatusHealthy && state.Lifecycle == LifecycleHealthy {
			return cap, nil
		}
	}

	// Fallback to the first available if none are explicitly healthy (for early boot / tests)
	if len(candidates) > 0 {
		return candidates[0], nil
	}

	return nil, errors.New("no viable capability available for requirement: " + requirementID)
}
