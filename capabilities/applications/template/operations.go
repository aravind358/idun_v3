package template

// TemplateOperation defines strongly-typed constants for permitted operations.
type TemplateOperation string

const (
	OperationExample TemplateOperation = "example_operation"
)

// IsValid validates if a string matches a known TemplateOperation.
func (o TemplateOperation) IsValid() bool {
	switch o {
	case OperationExample:
		return true
	}
	return false
}
