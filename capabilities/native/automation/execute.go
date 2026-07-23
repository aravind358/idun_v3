package automation

import (
	"context"
	"fmt"
	"time"

	"idun/capabilities"
)

// Execute fulfills the Capability interface.
func (c *Capability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	start := time.Now()

	// 1. Validation
	if err := c.validateRequest(req); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Validation", err)
	}

	// 2. Lifecycle Check
	if err := c.checkLifecycle(); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Unavailable", err)
	}

	opStr := req.Parameters["operation"]
	operation := AutomationOperation(opStr)

	// 3. Permission Check
	if err := c.checkPermission(ctx, req.RequirementID, operation); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Permission", err)
	}

	// 4. Execution Router
	var data map[string]interface{}
	var execErr error

	switch operation {
	case OperationMouseMove, OperationMouseClick, OperationMouseScroll:
		data, execErr = c.executeMouse(ctx, req, operation)
	case OperationKeyboardPress, OperationKeyboardRelease, OperationKeyboardType:
		data, execErr = c.executeKeyboard(ctx, req, operation)
	case OperationClipboardRead, OperationClipboardWrite:
		data, execErr = c.executeClipboard(ctx, req, operation)
	case OperationCaptureScreen:
		data, execErr = c.executeScreen(ctx, req, operation)
	case OperationListWindows, OperationGetWindow, OperationFocusWindow, OperationMinimizeWindow, OperationMaximizeWindow, OperationRestoreWindow:
		data, execErr = c.executeWindows(ctx, req, operation)
	case OperationListProcesses, OperationGetProcess:
		data, execErr = c.executeProcess(ctx, req, operation)
	default:
		execErr = fmt.Errorf("unknown operation: %s", operation)
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Validation", execErr)
	}

	if execErr != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Execution", execErr)
	}

	c.metrics.RecordSuccess(time.Since(start))
	return c.normalizeResult(req.RequirementID, start, data), nil
}
