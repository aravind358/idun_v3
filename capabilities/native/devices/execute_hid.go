package devices

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeHID(ctx context.Context, req capabilities.CapabilityRequest, op DeviceOperation) (map[string]interface{}, error) {
	switch op {
	case OperationListHID:
		data, err := c.provider.ListHID(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"devices": data,
		}, nil
	}
	return nil, nil
}
