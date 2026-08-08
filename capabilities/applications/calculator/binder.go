package calculator

import (
	"errors"
	"strconv"
)

// CalculatorRequest is the strongly-typed internal representation of a calculator request.
type CalculatorRequest struct {
	Intent    string
	Operation CalculatorOperation
	OperandA  float64
	OperandB  float64

	// Reserved for future growth
	Expression string
	Precision  int
	Format     string
}

// BindCalculatorRequest converts generic map parameters into a typed CalculatorRequest,
// performing centralized validation, syntax parsing, and normalization.
func BindCalculatorRequest(params map[string]string) (CalculatorRequest, error) {
	req := CalculatorRequest{
		Intent: params["intent"],
	}

	// 1. Validation and Normalization
	if req.Intent == "" {
		req.Intent = "calculate"
	}

	opStr := params["operation"]
	if opStr == "" {
		return req, errors.New("missing 'operation' parameter")
	}

	op := CalculatorOperation(opStr)
	if !op.IsValid() {
		return req, errors.New("unsupported operation: " + opStr)
	}
	req.Operation = op

	// 2. Syntax Validation and Parsing
	aStr, ok1 := params["a"]
	bStr, ok2 := params["b"]
	
	if !ok1 || !ok2 {
		return req, errors.New("missing operands 'a' or 'b'")
	}

	a, err1 := strconv.ParseFloat(aStr, 64)
	b, err2 := strconv.ParseFloat(bStr, 64)

	if err1 != nil || err2 != nil {
		return req, errors.New("operands 'a' and 'b' must be numeric")
	}

	req.OperandA = a
	req.OperandB = b

	return req, nil
}
