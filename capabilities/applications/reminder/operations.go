package reminder

// ReminderOperation defines strongly-typed constants for permitted operations.
type ReminderOperation string

const (
	OperationSet    ReminderOperation = "set"
	OperationGet    ReminderOperation = "get"
	OperationList   ReminderOperation = "list"
	OperationCancel ReminderOperation = "cancel"
)

// IsValid validates if a string matches a known ReminderOperation.
func (o ReminderOperation) IsValid() bool {
	switch o {
	case OperationSet, OperationGet, OperationList, OperationCancel:
		return true
	}
	return false
}
