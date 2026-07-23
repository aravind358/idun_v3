package devices

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeMetadata(ctx context.Context, req capabilities.CapabilityRequest, op DeviceOperation) (map[string]interface{}, error) {
	switch op {
	case OperationDeviceMetadata:
		deviceID := req.Parameters["device_id"]
		metadata, err := c.provider.DeviceMetadata(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"device_id": deviceID,
			"metadata":  metadata,
		}, nil
	}
	return nil, nil
}
