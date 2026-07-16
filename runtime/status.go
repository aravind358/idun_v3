package runtime

// RuntimeStatus represents the operational state of the IDUN cognitive runtime.
// Owned by RuntimeHost and Kernel.
type RuntimeStatus string

const (
	// StatusStopped indicates the runtime has not started or has completely shut down.
	StatusStopped RuntimeStatus = "STOPPED"
	// StatusStarting indicates the runtime is currently executing topological startup phases.
	StatusStarting RuntimeStatus = "STARTING"
	// StatusRunning indicates all enabled subsystems booted successfully and the runtime is active.
	StatusRunning RuntimeStatus = "RUNNING"
	// StatusStopping indicates the runtime is currently executing reverse-topological shutdown.
	StatusStopping RuntimeStatus = "STOPPING"
	// StatusFailed indicates startup encountered a terminal error and performed rollback.
	StatusFailed RuntimeStatus = "FAILED"
)
