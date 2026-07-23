package automation

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) checkPermission(ctx context.Context, reqID string, operation AutomationOperation) error {
	if c.permManager == nil {
		return nil
	}

	// Uses the Automation category for standard capability permissions
	return c.permManager.CheckPermission(ctx, capabilities.CategoryAutomation, reqID)
}
