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
	OperationBattery    SystemOperation = "battery"

	OperationShutdown   SystemOperation = "shutdown"
	OperationRestart    SystemOperation = "restart"
	OperationSleep      SystemOperation = "sleep"
	OperationLock       SystemOperation = "lock"

	OperationScheduleTask SystemOperation = "schedule_task"
	OperationCancelTask   SystemOperation = "cancel_task"
	OperationListTasks    SystemOperation = "list_tasks"
)

// IsValid validates if a string matches a known SystemOperation.
func (o SystemOperation) IsValid() bool {
	switch o {
	case OperationSystemInfo, OperationEnv, OperationHost, OperationCPU, OperationMemory, OperationDisk, OperationBattery,
		OperationShutdown, OperationRestart, OperationSleep, OperationLock,
		OperationScheduleTask, OperationCancelTask, OperationListTasks:
		return true
	}
	return false
}
