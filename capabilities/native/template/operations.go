package template

// TemplateOperation defines strongly-typed constants for permitted operations.
type TemplateOperation string

const (
	OperationExample TemplateOperation = "example_operation" // TODO: Replace with real operations
)

// IsValid validates if a string matches a known TemplateOperation.
func (o TemplateOperation) IsValid() bool {
	switch o {
	case OperationExample: // TODO: Add real operations here
		return true
	}
	return false
}
