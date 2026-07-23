package devices

// DeviceOperation defines strongly-typed constants for permitted operations.
type DeviceOperation string

const (
	// USB
	OperationListUSBDevices DeviceOperation = "list_usb_devices"
	OperationGetUSBDevice   DeviceOperation = "get_usb_device"

	// Bluetooth
	OperationListBluetoothDevices DeviceOperation = "list_bluetooth_devices"
	OperationPairBluetooth        DeviceOperation = "pair_bluetooth"
	OperationUnpairBluetooth      DeviceOperation = "unpair_bluetooth"

	// Battery
	OperationBatteryStatus DeviceOperation = "battery_status"

	// Power
	OperationPowerStatus DeviceOperation = "power_status"

	// Location and Sensors
	OperationGetGPS           DeviceOperation = "get_gps"
	OperationGetAccelerometer DeviceOperation = "get_accelerometer"
	OperationGetGyroscope     DeviceOperation = "get_gyroscope"
	OperationGetCompass       DeviceOperation = "get_compass"
	OperationGetSensor        DeviceOperation = "get_sensor"

	// HID
	OperationListHID DeviceOperation = "list_hid"

	// Metadata
	OperationDeviceMetadata DeviceOperation = "device_metadata"
)

// IsValid validates if a string matches a known DeviceOperation.
func (o DeviceOperation) IsValid() bool {
	switch o {
	case OperationListUSBDevices, OperationGetUSBDevice,
		OperationListBluetoothDevices, OperationPairBluetooth, OperationUnpairBluetooth,
		OperationBatteryStatus, OperationPowerStatus,
		OperationGetGPS, OperationGetAccelerometer, OperationGetGyroscope, OperationGetCompass, OperationGetSensor,
		OperationListHID, OperationDeviceMetadata:
		return true
	}
	return false
}
