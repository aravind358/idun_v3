package devices

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeSensors(ctx context.Context, req capabilities.CapabilityRequest, op DeviceOperation) (map[string]interface{}, error) {
	switch op {
	case OperationGetAccelerometer:
		data, err := c.provider.GetAccelerometer(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"accelerometer": data,
		}, nil

	case OperationGetGyroscope:
		data, err := c.provider.GetGyroscope(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"gyroscope": data,
		}, nil

	case OperationGetCompass:
		data, err := c.provider.GetCompass(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"compass": data,
		}, nil

	case OperationGetSensor:
		sensorType := req.Parameters["sensor_type"]
		data, err := c.provider.GetSensor(ctx, sensorType)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"sensor": data,
		}, nil
	}
	return nil, nil
}
