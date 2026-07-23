package automation

import (
	"context"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeMouse(ctx context.Context, req capabilities.CapabilityRequest, op AutomationOperation) (map[string]interface{}, error) {
	switch op {
	case OperationMouseMove:
		x, _ := strconv.Atoi(req.Parameters["x"])
		y, _ := strconv.Atoi(req.Parameters["y"])
		if err := c.provider.MouseMove(ctx, x, y); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "moved", "x": x, "y": y}, nil

	case OperationMouseClick:
		button := req.Parameters["button"]
		if button == "" {
			button = "left"
		}
		clicks := 1
		if cStr, ok := req.Parameters["clicks"]; ok {
			if parsed, err := strconv.Atoi(cStr); err == nil && parsed > 0 {
				clicks = parsed
			}
		}
		if err := c.provider.MouseClick(ctx, button, clicks); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "clicked", "button": button, "clicks": clicks}, nil

	case OperationMouseScroll:
		dx, _ := strconv.Atoi(req.Parameters["delta_x"])
		dy, _ := strconv.Atoi(req.Parameters["delta_y"])
		if err := c.provider.MouseScroll(ctx, dx, dy); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "scrolled", "delta_x": dx, "delta_y": dy}, nil
	}
	return nil, nil
}
