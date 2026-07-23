package media

import (
	"context"
	"encoding/base64"

	"idun/capabilities"
)

func (c *Capability) executeImage(ctx context.Context, req capabilities.CapabilityRequest, op MediaOperation) (map[string]interface{}, error) {
	switch op {
	case OperationCaptureImage:
		deviceID := req.Parameters["device_id"]
		destination := req.Parameters["destination"]
		err := c.provider.CaptureImage(ctx, deviceID, destination)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":      "captured",
			"destination": destination,
		}, nil

	case OperationLoadImage:
		path := req.Parameters["path"]
		info, err := c.provider.LoadImage(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status": "loaded",
			"path":   path,
			"info":   info,
		}, nil

	case OperationSaveImage:
		destination := req.Parameters["destination"]
		var payload []byte
		if b64, ok := req.Parameters["data_bytes_base64"]; ok {
			var err error
			payload, err = base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, err
			}
		}

		err := c.provider.SaveImage(ctx, destination, payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":      "saved",
			"destination": destination,
		}, nil
	}

	return nil, nil
}
