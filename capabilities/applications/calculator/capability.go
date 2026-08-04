package calculator

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"idun/capabilities"
	"idun/capabilities/applications/core"
)

// Capability defines the calculator application capability.
// It is a Model 2 App Capability (Pure Deterministic Execution).
type Capability struct {
	core.AppCapability
}

// New creates a new instance of the Calculator Capability.
func New(deps core.AppCapabilityDependencies) *Capability {
	return &Capability{
		AppCapability: core.NewAppCapability("app-calc-1", Metadata(), deps.Resolver),
	}
}

// Execute fulfills the Capability interface.
func (c *Capability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	start := time.Now()

	// 1. Validation
	if err := c.validateRequest(req); err != nil {
		return c.normalizeError(req.RequirementID, start, "Validation", err)
	}

	// 2. Lifecycle Check
	if err := c.checkLifecycle(); err != nil {
		return c.normalizeError(req.RequirementID, start, "Unavailable", err)
	}

	opStr := req.Parameters["operation"]
	operation := CalculatorOperation(opStr)

	// 3. Execution
	data, execErr := c.executeMath(operation, req.Parameters)

	if execErr != nil {
		return c.normalizeError(req.RequirementID, start, "Execution", execErr)
	}

	return c.normalizeResult(req.RequirementID, start, data), nil
}

func (c *Capability) validateRequest(req capabilities.CapabilityRequest) error {
	if req.RequirementID == "" {
		return errors.New("missing requirement ID")
	}

	operation := req.Parameters["operation"]
	if operation == "" {
		return errors.New("missing 'operation' parameter")
	}

	op := CalculatorOperation(operation)
	if !op.IsValid() {
		return errors.New("unsupported operation: " + operation)
	}

	return nil
}

func (c *Capability) checkLifecycle() error {
	state := c.State().Lifecycle
	if state == "DISABLED" || state == "UNLOADED" {
		return errors.New("capability is not currently available for execution: " + string(state))
	}
	return nil
}

func (c *Capability) executeMath(operation CalculatorOperation, params map[string]string) (map[string]interface{}, error) {
	aStr, ok1 := params["a"]
	bStr, ok2 := params["b"]
	
	if !ok1 || !ok2 {
		return nil, errors.New("missing operands 'a' or 'b'")
	}

	a, err1 := strconv.ParseFloat(aStr, 64)
	b, err2 := strconv.ParseFloat(bStr, 64)

	if err1 != nil || err2 != nil {
		return nil, errors.New("operands 'a' and 'b' must be numeric")
	}

	var result float64

	switch operation {
	case OperationAdd:
		result = a + b
	case OperationSubtract:
		result = a - b
	case OperationMultiply:
		result = a * b
	case OperationDivide:
		if b == 0 {
			return nil, errors.New("division by zero")
		}
		result = a / b
	case OperationModulo:
		if b == 0 {
			return nil, errors.New("modulo by zero")
		}
		result = float64(int64(a) % int64(b)) // Simple modulo for MVP
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	return map[string]interface{}{
		"result": result,
		"a":      a,
		"b":      b,
		"op":     string(operation),
	}, nil
}

func (c *Capability) normalizeResult(reqID string, start time.Time, data map[string]interface{}) capabilities.CapabilityResult {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       true,
		Realization:   capabilities.Deterministic, // Pure math is deterministic
		ResponseType:  "calculator",
		Data:          data,
		Duration:      time.Since(start),
	}
}

func (c *Capability) normalizeError(reqID string, start time.Time, code string, err error) (capabilities.CapabilityResult, error) {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       false,
		Error: &capabilities.CapabilityError{
			Code:    code,
			Message: err.Error(),
			Retry:   false,
		},
		Duration:      time.Since(start),
	}, nil
}
