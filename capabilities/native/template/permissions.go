package template

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) checkPermission(ctx context.Context, reqID string, operation TemplateOperation) error {
	if c.permManager == nil {
		return nil
	}

	// TODO: Map operations to proper permission categories and requests
	if operation == OperationExample {
		return c.permManager.CheckPermission(ctx, capabilities.CategorySystem, reqID)
	}

	return c.permManager.CheckPermission(ctx, capabilities.CategorySystem, reqID)
}
