package files

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) checkPermission(ctx context.Context, reqID string, operation FileOperation) error {
	if c.permManager == nil {
		return nil
	}

	// Currently, all file operations check the "Files" category.
	// A more granular PermissionManager would inspect the ReqID to differentiate
	// read vs write vs delete based on the requirement schema.
	return c.permManager.CheckPermission(ctx, capabilities.CategoryFiles, reqID)
}
