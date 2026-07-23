package devices

import (
	"context"
	"fmt"
	"time"

	"idun/capabilities"
)

// Execute fulfills the Capability interface.
func (c *Capability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	start := time.Now()

	// 1. Validation
	if err := c.validateRequest(req); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Validation", err)
	}

	// 2. Lifecycle Check
	if err := c.checkLifecycle(); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Unavailable", err)
	}

	opStr := req.Parameters["operation"]
	operation := DeviceOperation(opStr)

	// 3. Permission Check
	if err := c.checkPermission(ctx, req.RequirementID, operation); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Permission", err)
	}

	// 4. Execution Router
	var data map[string]interface{}
	var execErr error

	switch operation {
	case OperationListUSBDevices, OperationGetUSBDevice:
		data, execErr = c.executeUSB(ctx, req, operation)
	case OperationListBluetoothDevices, OperationPairBluetooth, OperationUnpairBluetooth:
		data, execErr = c.executeBluetooth(ctx, req, operation)
	case OperationBatteryStatus:
		data, execErr = c.executeBattery(ctx, req, operation)
	case OperationPowerStatus:
		data, execErr = c.executePower(ctx, req, operation)
	case OperationGetGPS:
		data, execErr = c.executeLocation(ctx, req, operation)
	case OperationGetAccelerometer, OperationGetGyroscope, OperationGetCompass, OperationGetSensor:
		data, execErr = c.executeSensors(ctx, req, operation)
	case OperationListHID:
		data, execErr = c.executeHID(ctx, req, operation)
	case OperationDeviceMetadata:
		data, execErr = c.executeMetadata(ctx, req, operation)
	default:
		execErr = fmt.Errorf("unknown operation: %s", operation)
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Validation", execErr)
	}

	if execErr != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Execution", execErr)
	}

	c.metrics.RecordSuccess(time.Since(start))
	return c.normalizeResult(req.RequirementID, start, data), nil
}
