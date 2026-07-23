package devices

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeBluetooth(ctx context.Context, req capabilities.CapabilityRequest, op DeviceOperation) (map[string]interface{}, error) {
	switch op {
	case OperationListBluetoothDevices:
		devices, err := c.provider.ListBluetoothDevices(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"devices": devices,
		}, nil

	case OperationPairBluetooth:
		deviceID := req.Parameters["device_id"]
		err := c.provider.PairBluetooth(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":    "paired",
			"device_id": deviceID,
		}, nil

	case OperationUnpairBluetooth:
		deviceID := req.Parameters["device_id"]
		err := c.provider.UnpairBluetooth(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":    "unpaired",
			"device_id": deviceID,
		}, nil
	}
	return nil, nil
}
