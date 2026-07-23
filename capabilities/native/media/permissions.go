package media

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) checkPermission(ctx context.Context, reqID string, operation MediaOperation) error {
	if c.permManager == nil {
		return nil
	}

	// Uses the Media category for standard capability permissions
	return c.permManager.CheckPermission(ctx, capabilities.CategoryMedia, reqID)
}
