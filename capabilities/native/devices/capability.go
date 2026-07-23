package devices

import (
	"idun/capabilities"
)

// Capability defines the native devices capability.
type Capability struct {
	capabilities.BaseCapability
	permManager capabilities.PermissionManager
	provider    DevicesProvider
	metrics     *CapabilityMetrics
}

// New creates a new instance of the Native Devices Capability.
func New(permManager capabilities.PermissionManager, provider DevicesProvider) *Capability {
	return &Capability{
		BaseCapability: capabilities.NewBaseCapability("devices-native-1", Metadata()),
		permManager:    permManager,
		provider:       provider,
		metrics:        NewCapabilityMetrics(),
	}
}

// Metrics returns diagnostic execution metrics safely.
func (c *Capability) Metrics() *CapabilityMetrics {
	return c.metrics
}
