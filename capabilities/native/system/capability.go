package system

import (
	"idun/capabilities"
	"idun/core/scheduler"
)

// Capability defines the native operating-system capability.
type Capability struct {
	capabilities.BaseCapability
	permManager capabilities.PermissionManager
	provider    SystemProvider
	scheduler   *scheduler.SchedulerService
	metrics     *CapabilityMetrics
}

// New creates a new instance of the Native System Capability.
func New(permManager capabilities.PermissionManager, provider SystemProvider, sched *scheduler.SchedulerService) *Capability {
	return &Capability{
		BaseCapability: capabilities.NewBaseCapability("sys-native-1", Metadata()),
		permManager:    permManager,
		provider:       provider,
		scheduler:      sched,
		metrics:        NewCapabilityMetrics(),
	}
}

// Metrics returns diagnostic execution metrics safely.
func (c *Capability) Metrics() *CapabilityMetrics {
	return c.metrics
}
