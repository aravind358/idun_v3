package template

import (
	"context"
	"errors"
	"fmt"
	"time"

	"idun/capabilities"
	"idun/capabilities/applications/core"
)

// Capability defines the application template capability.
type Capability struct {
	core.AppCapability
}

// New creates a new instance of the Template Capability.
func New(deps core.AppCapabilityDependencies) *Capability {
	return &Capability{
		AppCapability: core.NewAppCapability("app-tmpl-1", Metadata(), deps.Resolver),
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
	operation := TemplateOperation(opStr)

	// 3. Execution
	var data map[string]interface{}
	var execErr error

	switch operation {
	case OperationExample:
		data, execErr = c.executeExample(req.Parameters)
	default:
		execErr = fmt.Errorf("operation %s not implemented", operation)
	}

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

	op := TemplateOperation(operation)
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

func (c *Capability) executeExample(params map[string]string) (map[string]interface{}, error) {
	// Implement Application logic here
	// Examples:
	// - Perform deterministic calculation (Model 2)
	// - Use c.Resolver.Resolve(...) to invoke Native capabilities (Model 1)
	return map[string]interface{}{
		"status": "success",
	}, nil
}

func (c *Capability) normalizeResult(reqID string, start time.Time, data map[string]interface{}) capabilities.CapabilityResult {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       true,
		Realization:   capabilities.Deterministic, // Or Generative if LLM formatting is needed
		ResponseType:  "template",
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
