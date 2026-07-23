package system

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) checkPermission(ctx context.Context, reqID string, operation SystemOperation) error {
	if c.permManager == nil {
		return nil
	}

	// Read operations
	if operation == OperationSystemInfo || operation == OperationEnv || operation == OperationHost || operation == OperationCPU || operation == OperationMemory || operation == OperationDisk {
		return c.permManager.CheckPermission(ctx, capabilities.CategorySystem, reqID)
	}

	// Power operations require elevated execution permissions
	return c.permManager.CheckPermission(ctx, capabilities.CategorySystem, reqID)
}
