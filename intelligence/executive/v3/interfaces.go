package v3

import (
	"context"
	"idun/core/foundation"
	"idun/intelligence/planning/v3"
)

// CapabilityExecutor represents a physical backend driver capable of executing
// a requested action and returning a payload.
type CapabilityExecutor interface {
	Execute(ctx context.Context, params map[string]any) (payload []byte, err error)
}

// CapabilityRegistry maps abstract capability URIs (e.g., "urn:capability:device.turn_on")
// to concrete physical CapabilityExecutor drivers.
type CapabilityRegistry interface {
	Resolve(capabilityID string) (CapabilityExecutor, error)
}

// PlanProvider is an abstract interface (typically fulfilled by Memory)
// for retrieving the underlying ExecutionPlan authorized by a DecisionRecord.
type PlanProvider interface {
	GetPlan(ctx context.Context, planID foundation.ArtifactID) (*v3.ExecutionPlan, error)
}

// MemoryProvider abstracts the Content-Addressed Storage (CAS) interactions.
// It allows the Executive to store physical capability payloads and retrieve an OutputRef URI.
type MemoryProvider interface {
	StorePayload(ctx context.Context, payload []byte) (outputRef string, err error)
}
