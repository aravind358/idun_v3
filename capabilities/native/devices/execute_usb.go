package devices

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeUSB(ctx context.Context, req capabilities.CapabilityRequest, op DeviceOperation) (map[string]interface{}, error) {
	switch op {
	case OperationListUSBDevices:
		devices, err := c.provider.ListUSBDevices(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"devices": devices,
		}, nil

	case OperationGetUSBDevice:
		deviceID := req.Parameters["device_id"]
		device, err := c.provider.GetUSBDevice(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"device": device,
		}, nil
	}
	return nil, nil
}
