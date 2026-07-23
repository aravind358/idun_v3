package media

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeDevices(ctx context.Context, req capabilities.CapabilityRequest, op MediaOperation) (map[string]interface{}, error) {
	switch op {
	case OperationListMediaDevices:
		deviceType := req.Parameters["device_type"]
		devices, err := c.provider.ListMediaDevices(ctx, deviceType)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"device_type": deviceType,
			"devices":     devices,
		}, nil

	case OperationGetDevice:
		deviceID := req.Parameters["device_id"]
		device, err := c.provider.GetDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"device": device,
		}, nil

	case OperationListCodecs:
		codecs, err := c.provider.ListCodecs(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"codecs": codecs,
		}, nil
	}

	return nil, nil
}
