package files

import (
	"idun/capabilities"
)

// Capability defines the native files capability.
type Capability struct {
	capabilities.BaseCapability
	permManager capabilities.PermissionManager
	provider    FileProvider
	metrics     *CapabilityMetrics
}

// New creates a new instance of the Native Files Capability.
func New(permManager capabilities.PermissionManager, provider FileProvider) *Capability {
	return &Capability{
		BaseCapability: capabilities.NewBaseCapability("files-native-1", Metadata()),
		permManager:    permManager,
		provider:       provider,
		metrics:        NewCapabilityMetrics(),
	}
}

// Metrics returns diagnostic execution metrics safely.
func (c *Capability) Metrics() *CapabilityMetrics {
	return c.metrics
}
