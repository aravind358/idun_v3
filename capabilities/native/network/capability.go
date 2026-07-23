package network

import (
	"idun/capabilities"
)

// Capability defines the native network capability.
type Capability struct {
	capabilities.BaseCapability
	permManager capabilities.PermissionManager
	provider    NetworkProvider
	metrics     *CapabilityMetrics
}

// New creates a new instance of the Native Network Capability.
func New(permManager capabilities.PermissionManager, provider NetworkProvider) *Capability {
	return &Capability{
		BaseCapability: capabilities.NewBaseCapability("network-native-1", Metadata()),
		permManager:    permManager,
		provider:       provider,
		metrics:        NewCapabilityMetrics(),
	}
}

// Metrics returns diagnostic execution metrics safely.
func (c *Capability) Metrics() *CapabilityMetrics {
	return c.metrics
}
