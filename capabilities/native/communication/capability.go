package communication

import (
	"idun/capabilities"
)

// Capability defines the native communication capability.
type Capability struct {
	capabilities.BaseCapability
	permManager capabilities.PermissionManager
	provider    CommunicationProvider
	metrics     *CapabilityMetrics
}

// New creates a new instance of the Native Communication Capability.
func New(permManager capabilities.PermissionManager, provider CommunicationProvider) *Capability {
	return &Capability{
		BaseCapability: capabilities.NewBaseCapability("comm-native-1", Metadata()),
		permManager:    permManager,
		provider:       provider,
		metrics:        NewCapabilityMetrics(),
	}
}

// Metrics returns diagnostic execution metrics safely.
func (c *Capability) Metrics() *CapabilityMetrics {
	return c.metrics
}
