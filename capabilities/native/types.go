package native

import (
	"idun/core/memory"
	"idun/core/scheduler"
	"idun/core/storage"
	coretime "idun/core/time"
	"idun/intelligence/workspace"
)

// NativeCapabilityDependencies is a generic dependency container for all native capabilities.
// This allows the Capability Framework to instantiate capabilities with required core services
// without cluttering the initialization function signatures.
type NativeCapabilityDependencies struct {
	Time      coretime.TimeService
	Memory    memory.Memory
	Storage   *storage.Storage
	Scheduler *scheduler.SchedulerService
	Workspace workspace.Workspace
}
