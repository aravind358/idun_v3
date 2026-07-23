package system

import (
	"idun/capabilities"
)

// Capability defines the native operating-system capability.
type Capability struct {
	capabilities.BaseCapability
	permManager capabilities.PermissionManager
	provider    SystemProvider
	metrics     *CapabilityMetrics
}

// New creates a new instance of the Native System Capability.
func New(permManager capabilities.PermissionManager, provider SystemProvider) *Capability {
	return &Capability{
		BaseCapability: capabilities.NewBaseCapability("sys-native-1", Metadata()),
		permManager:    permManager,
		provider:       provider,
		metrics:        NewCapabilityMetrics(),
	}
}

// Metrics returns diagnostic execution metrics safely.
func (c *Capability) Metrics() *CapabilityMetrics {
	return c.metrics
}
