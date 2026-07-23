package automation

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeWindows(ctx context.Context, req capabilities.CapabilityRequest, op AutomationOperation) (map[string]interface{}, error) {
	switch op {
	case OperationListWindows:
		windows, err := c.provider.ListWindows(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"windows": windows}, nil

	case OperationGetWindow:
		handle := req.Parameters["handle"]
		window, err := c.provider.GetWindow(ctx, handle)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"window": window}, nil

	case OperationFocusWindow:
		handle := req.Parameters["handle"]
		if err := c.provider.FocusWindow(ctx, handle); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "focused", "handle": handle}, nil

	case OperationMinimizeWindow:
		handle := req.Parameters["handle"]
		if err := c.provider.MinimizeWindow(ctx, handle); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "minimized", "handle": handle}, nil

	case OperationMaximizeWindow:
		handle := req.Parameters["handle"]
		if err := c.provider.MaximizeWindow(ctx, handle); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "maximized", "handle": handle}, nil

	case OperationRestoreWindow:
		handle := req.Parameters["handle"]
		if err := c.provider.RestoreWindow(ctx, handle); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "restored", "handle": handle}, nil
	}
	return nil, nil
}
