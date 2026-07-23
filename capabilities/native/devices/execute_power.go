package devices

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executePower(ctx context.Context, req capabilities.CapabilityRequest, op DeviceOperation) (map[string]interface{}, error) {
	switch op {
	case OperationPowerStatus:
		status, err := c.provider.PowerStatus(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status": status,
		}, nil
	}
	return nil, nil
}
