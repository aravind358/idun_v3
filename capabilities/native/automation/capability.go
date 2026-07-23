package automation

import (
	"idun/capabilities"
)

// Capability defines the native automation capability.
type Capability struct {
	capabilities.BaseCapability
	permManager capabilities.PermissionManager
	provider    AutomationProvider
	metrics     *CapabilityMetrics
}

// New creates a new instance of the Native Automation Capability.
func New(permManager capabilities.PermissionManager, provider AutomationProvider) *Capability {
	return &Capability{
		BaseCapability: capabilities.NewBaseCapability("automation-native-1", Metadata()),
		permManager:    permManager,
		provider:       provider,
		metrics:        NewCapabilityMetrics(),
	}
}

// Metrics returns diagnostic execution metrics safely.
func (c *Capability) Metrics() *CapabilityMetrics {
	return c.metrics
}
