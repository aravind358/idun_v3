package automation

import (
	"context"
	"encoding/base64"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeScreen(ctx context.Context, req capabilities.CapabilityRequest, op AutomationOperation) (map[string]interface{}, error) {
	switch op {
	case OperationCaptureScreen:
		region := make(map[string]int)
		if x, err := strconv.Atoi(req.Parameters["x"]); err == nil {
			region["x"] = x
		}
		if y, err := strconv.Atoi(req.Parameters["y"]); err == nil {
			region["y"] = y
		}
		if w, err := strconv.Atoi(req.Parameters["width"]); err == nil {
			region["width"] = w
		}
		if h, err := strconv.Atoi(req.Parameters["height"]); err == nil {
			region["height"] = h
		}

		imgData, err := c.provider.CaptureScreen(ctx, region)
		if err != nil {
			return nil, err
		}
		
		b64 := base64.StdEncoding.EncodeToString(imgData)
		return map[string]interface{}{
			"status": "captured",
			"image_base64": b64,
		}, nil
	}
	return nil, nil
}
