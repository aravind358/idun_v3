package devices

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeBattery(ctx context.Context, req capabilities.CapabilityRequest, op DeviceOperation) (map[string]interface{}, error) {
	switch op {
	case OperationBatteryStatus:
		status, err := c.provider.BatteryStatus(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status": status,
		}, nil
	}
	return nil, nil
}
