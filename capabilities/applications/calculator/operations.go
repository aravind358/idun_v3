package calculator

// CalculatorOperation defines strongly-typed constants for permitted operations.
type CalculatorOperation string

const (
	OperationAdd      CalculatorOperation = "add"
	OperationSubtract CalculatorOperation = "subtract"
	OperationMultiply CalculatorOperation = "multiply"
	OperationDivide   CalculatorOperation = "divide"
	OperationModulo   CalculatorOperation = "modulo"
)

// IsValid validates if a string matches a known CalculatorOperation.
func (o CalculatorOperation) IsValid() bool {
	switch o {
	case OperationAdd, OperationSubtract, OperationMultiply, OperationDivide, OperationModulo:
		return true
	}
	return false
}
