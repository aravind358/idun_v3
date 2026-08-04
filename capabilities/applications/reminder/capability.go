package reminder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"idun/capabilities"
	"idun/capabilities/applications/core"
)

// Capability defines the reminder application capability.
// It is a Model 1 App Capability (Orchestration).
type Capability struct {
	core.AppCapability
}

// New creates a new instance of the Reminder Capability.
func New(deps core.AppCapabilityDependencies) *Capability {
	return &Capability{
		AppCapability: core.NewAppCapability("app-rem-1", Metadata(), deps.Resolver),
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
	operation := ReminderOperation(opStr)

	// 3. Execution
	var data map[string]interface{}
	var execErr error

	switch operation {
	case OperationSet:
		data, execErr = c.executeSet(ctx, req.RequirementID, req.Parameters)
	case OperationCancel:
		data, execErr = c.executeCancel(ctx, req.RequirementID, req.Parameters)
	case OperationList:
		data, execErr = c.executeList(ctx, req.RequirementID)
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

	op := ReminderOperation(operation)
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

func (c *Capability) executeSet(ctx context.Context, reqID string, params map[string]string) (map[string]interface{}, error) {
	if c.Resolver == nil {
		return nil, errors.New("native capability resolver is not wired")
	}

	sysParams := map[string]string{
		"operation": "schedule_task",
		"message":   params["message"],
		"time":      params["time"],
	}

	sysCap, err := c.Resolver.Resolve(ctx, reqID, "sys-native-1", sysParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve native system capability: %w", err)
	}

	res, execErr := sysCap.Execute(ctx, capabilities.CapabilityRequest{
		RequirementID: reqID + "-sub",
		Parameters:    sysParams,
		ContextID:     reqID,
	})
	if execErr != nil {
		return nil, fmt.Errorf("native system execution failed: %w", execErr)
	}
	if !res.Success {
		return nil, fmt.Errorf("native system request failed: %v", res.Error.Message)
	}

	return res.Data, nil
}

func (c *Capability) executeCancel(ctx context.Context, reqID string, params map[string]string) (map[string]interface{}, error) {
	if c.Resolver == nil {
		return nil, errors.New("native capability resolver is not wired")
	}

	sysParams := map[string]string{
		"operation": "cancel_task",
		"id":        params["id"],
	}

	sysCap, err := c.Resolver.Resolve(ctx, reqID, "sys-native-1", sysParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve native system capability: %w", err)
	}

	res, execErr := sysCap.Execute(ctx, capabilities.CapabilityRequest{
		RequirementID: reqID + "-sub",
		Parameters:    sysParams,
		ContextID:     reqID,
	})
	if execErr != nil {
		return nil, fmt.Errorf("native system execution failed: %w", execErr)
	}
	if !res.Success {
		return nil, fmt.Errorf("native system request failed: %v", res.Error.Message)
	}

	return res.Data, nil
}

func (c *Capability) executeList(ctx context.Context, reqID string) (map[string]interface{}, error) {
	if c.Resolver == nil {
		return nil, errors.New("native capability resolver is not wired")
	}

	sysParams := map[string]string{
		"operation": "list_tasks",
	}

	sysCap, err := c.Resolver.Resolve(ctx, reqID, "sys-native-1", sysParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve native system capability: %w", err)
	}

	res, execErr := sysCap.Execute(ctx, capabilities.CapabilityRequest{
		RequirementID: reqID + "-sub",
		Parameters:    sysParams,
		ContextID:     reqID,
	})
	if execErr != nil {
		return nil, fmt.Errorf("native system execution failed: %w", execErr)
	}
	if !res.Success {
		return nil, fmt.Errorf("native system request failed: %v", res.Error.Message)
	}

	return res.Data, nil
}

func (c *Capability) normalizeResult(reqID string, start time.Time, data map[string]interface{}) capabilities.CapabilityResult {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       true,
		Realization:   capabilities.Deterministic,
		ResponseType:  "reminder",
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
