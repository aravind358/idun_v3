package communication

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) checkPermission(ctx context.Context, reqID string, operation CommunicationOperation) error {
	if c.permManager == nil {
		return nil
	}

	// Uses the Communication category for standard capability permissions
	return c.permManager.CheckPermission(ctx, capabilities.CategoryCommunication, reqID)
}
