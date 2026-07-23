package network

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) checkPermission(ctx context.Context, reqID string, operation NetworkOperation) error {
	if c.permManager == nil {
		return nil
	}

	// Uses the Network category for standard capability permissions
	return c.permManager.CheckPermission(ctx, capabilities.CategoryNetwork, reqID)
}
