package system

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
	operation := SystemOperation(opStr)

	// 3. Permission Check
	if err := c.checkPermission(ctx, req.RequirementID, operation); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Permission", err)
	}

	// 4. Execution Router
	var data map[string]interface{}
	var execErr error

	switch operation {
	case OperationSystemInfo:
		data, execErr = c.executeInfo(ctx)
	case OperationEnv:
		data, execErr = c.executeEnv(ctx)
	case OperationHost:
		data, execErr = c.executeHost(ctx)
	case OperationCPU:
		data, execErr = c.executeCPU(ctx)
	case OperationMemory:
		data, execErr = c.executeMemory(ctx)
	case OperationBattery:
		data, execErr = c.executeBattery(ctx)
	case OperationDisk:
		data, execErr = c.executeDisk(ctx)
	case OperationShutdown, OperationRestart, OperationSleep, OperationLock:
		data, execErr = c.executePower(ctx, operation)
	case OperationScheduleTask:
		data, execErr = c.executeScheduleTask(ctx, req)
	case OperationCancelTask:
		data, execErr = c.executeCancelTask(ctx, req)
	case OperationListTasks:
		data, execErr = c.executeListTasks(ctx, req)
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
