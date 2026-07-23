package system

// SystemOperation defines strongly-typed constants for permitted operations.
type SystemOperation string

const (
	OperationSystemInfo SystemOperation = "info"
	OperationEnv        SystemOperation = "env"
	OperationHost       SystemOperation = "host"
	OperationCPU        SystemOperation = "cpu"
	OperationMemory     SystemOperation = "memory"
	OperationDisk       SystemOperation = "disk"

	OperationShutdown   SystemOperation = "shutdown"
	OperationRestart    SystemOperation = "restart"
	OperationSleep      SystemOperation = "sleep"
	OperationLock       SystemOperation = "lock"
)

// IsValid validates if a string matches a known SystemOperation.
func (o SystemOperation) IsValid() bool {
	switch o {
	case OperationSystemInfo, OperationEnv, OperationHost, OperationCPU, OperationMemory, OperationDisk,
		OperationShutdown, OperationRestart, OperationSleep, OperationLock:
		return true
	}
	return false
}
