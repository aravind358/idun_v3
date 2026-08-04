package capabilities

import (
	"context"
	"time"
)

// BaseCapability provides a reusable foundation for implementing capabilities.
type BaseCapability struct {
	id       string
	metadata CapabilityMetadata
	state    CapabilityState
}

func NewBaseCapability(id string, metadata CapabilityMetadata) BaseCapability {
	return BaseCapability{
		id:       id,
		metadata: metadata,
		state: CapabilityState{
			Lifecycle:   LifecycleDiscovered,
			Operational: StatusHealthy,
			LastUpdated: time.Now(),
		},
	}
}

func (b *BaseCapability) ID() string {
	return b.id
}

func (b *BaseCapability) Metadata() CapabilityMetadata {
	return b.metadata
}

func (b *BaseCapability) State() CapabilityState {
	return b.state
}

func (b *BaseCapability) SetLifecycleState(state LifecycleState) {
	b.state.Lifecycle = state
	b.state.LastUpdated = time.Now()
}

func (b *BaseCapability) SetOperationalStatus(status OperationalStatus) {
	b.state.Operational = status
	b.state.LastUpdated = time.Now()
}

// Execute is a placeholder. Embedders must override this.
func (b *BaseCapability) Execute(ctx context.Context, req CapabilityRequest) (CapabilityResult, error) {
	return CapabilityResult{Realization: Generative}, nil
}
