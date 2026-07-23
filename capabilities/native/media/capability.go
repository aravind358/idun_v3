package media

import (
	"idun/capabilities"
)

// Capability defines the native media capability.
type Capability struct {
	capabilities.BaseCapability
	permManager capabilities.PermissionManager
	provider    MediaProvider
	metrics     *CapabilityMetrics
}

// New creates a new instance of the Native Media Capability.
func New(permManager capabilities.PermissionManager, provider MediaProvider) *Capability {
	return &Capability{
		BaseCapability: capabilities.NewBaseCapability("media-native-1", Metadata()),
		permManager:    permManager,
		provider:       provider,
		metrics:        NewCapabilityMetrics(),
	}
}

// Metrics returns diagnostic execution metrics safely.
func (c *Capability) Metrics() *CapabilityMetrics {
	return c.metrics
}
