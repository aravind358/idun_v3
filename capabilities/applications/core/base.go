package core

import (
	"context"
	"idun/capabilities"
)

// NativeCapabilityResolver provides Application capabilities with a focused interface
// to invoke registered Native capabilities without exposing Registry or LifecycleManager.
type NativeCapabilityResolver interface {
	Resolve(ctx context.Context, requirementID, capabilityName string, params map[string]string) (capabilities.Capability, error)
}

// CapabilityManagerResolver adapts a full CapabilityManager to the focused NativeCapabilityResolver interface.
type CapabilityManagerResolver struct {
	manager capabilities.CapabilityManager
}

func NewCapabilityManagerResolver(manager capabilities.CapabilityManager) *CapabilityManagerResolver {
	return &CapabilityManagerResolver{manager: manager}
}

func (r *CapabilityManagerResolver) Resolve(ctx context.Context, requirementID, capabilityName string, params map[string]string) (capabilities.Capability, error) {
	return r.manager.Resolver().Resolve(ctx, requirementID, capabilityName, params)
}

// AppCapability is the base struct for all Application capabilities.
// Embed this instead of capabilities.BaseCapability directly.
type AppCapability struct {
	capabilities.BaseCapability
	Resolver NativeCapabilityResolver
}

// NewAppCapability creates a new AppCapability base.
func NewAppCapability(id string, meta capabilities.CapabilityMetadata, resolver NativeCapabilityResolver) AppCapability {
	return AppCapability{
		BaseCapability: capabilities.NewBaseCapability(id, meta),
		Resolver:       resolver,
	}
}
