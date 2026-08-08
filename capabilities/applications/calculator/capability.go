package calculator

import (
	"context"
	"errors"
	"fmt"
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

	// 1. Binding & Validation
	typedReq, err := BindCalculatorRequest(req.Parameters)
	if err != nil {
		return c.normalizeError(req.RequirementID, start, "Validation", err)
	}

	// 2. Lifecycle Check
	if err := c.checkLifecycle(); err != nil {
		return c.normalizeError(req.RequirementID, start, "Unavailable", err)
	}

	// 3. Execution
	data, execErr := c.executeMath(typedReq)
	if execErr != nil {
		return c.normalizeError(req.RequirementID, start, "Execution", execErr)
	}

	return c.normalizeResult(req.RequirementID, start, typedReq.Intent, data), nil
}

func (c *Capability) checkLifecycle() error {
	state := c.State().Lifecycle
	if state == "DISABLED" || state == "UNLOADED" {
		return errors.New("capability is not currently available for execution: " + string(state))
	}
	return nil
}

func (c *Capability) executeMath(req CalculatorRequest) (map[string]interface{}, error) {
	var result float64

	switch req.Operation {
	case OperationAdd:
		result = req.OperandA + req.OperandB
	case OperationSubtract:
		result = req.OperandA - req.OperandB
	case OperationMultiply:
		result = req.OperandA * req.OperandB
	case OperationDivide:
		if req.OperandB == 0 {
			return nil, errors.New("division by zero")
		}
		result = req.OperandA / req.OperandB
	case OperationModulo:
		if req.OperandB == 0 {
			return nil, errors.New("modulo by zero")
		}
		result = float64(int64(req.OperandA) % int64(req.OperandB)) // Simple modulo for MVP
	default:
		return nil, fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	return map[string]interface{}{
		"result": result,
		"a":      req.OperandA,
		"b":      req.OperandB,
		"op":     string(req.Operation),
	}, nil
}

func (c *Capability) normalizeResult(reqID string, start time.Time, operation string, data map[string]interface{}) capabilities.CapabilityResult {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       true,
		Realization:   capabilities.Deterministic, // Pure math is deterministic
		ResponseType:  "calculator",
		Operation:     operation,
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
