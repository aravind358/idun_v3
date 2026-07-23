package devices

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) checkPermission(ctx context.Context, reqID string, operation DeviceOperation) error {
	if c.permManager == nil {
		return nil
	}

	// Uses the DevicesSensors category for standard capability permissions
	return c.permManager.CheckPermission(ctx, capabilities.CategoryDevicesSensors, reqID)
}
