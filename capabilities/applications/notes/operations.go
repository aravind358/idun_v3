package notes

// NotesOperation defines strongly-typed constants for permitted operations.
type NotesOperation string

const (
	OperationCreate NotesOperation = "create"
	OperationRead   NotesOperation = "read"
	OperationUpdate NotesOperation = "update"
	OperationDelete NotesOperation = "delete"
	OperationList   NotesOperation = "list"
)

// IsValid validates if a string matches a known NotesOperation.
func (o NotesOperation) IsValid() bool {
	switch o {
	case OperationCreate, OperationRead, OperationUpdate, OperationDelete, OperationList:
		return true
	}
	return false
}
