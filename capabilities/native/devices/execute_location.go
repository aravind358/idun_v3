package devices

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeLocation(ctx context.Context, req capabilities.CapabilityRequest, op DeviceOperation) (map[string]interface{}, error) {
	switch op {
	case OperationGetGPS:
		data, err := c.provider.GetGPS(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"gps": data,
		}, nil
	}
	return nil, nil
}
