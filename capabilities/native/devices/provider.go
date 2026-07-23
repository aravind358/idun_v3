package devices

import "context"

// DevicesProvider abstracts native hardware and sensor operations.
// This isolates the capability from host APIs and enables seamless mock testing.
type DevicesProvider interface {
	// USB
	ListUSBDevices(ctx context.Context) ([]map[string]interface{}, error)
	GetUSBDevice(ctx context.Context, deviceID string) (map[string]interface{}, error)

	// Bluetooth
	ListBluetoothDevices(ctx context.Context) ([]map[string]interface{}, error)
	PairBluetooth(ctx context.Context, deviceID string) error
	UnpairBluetooth(ctx context.Context, deviceID string) error

	// Battery and Power
	BatteryStatus(ctx context.Context) (map[string]interface{}, error)
	PowerStatus(ctx context.Context) (map[string]interface{}, error)

	// Location and Sensors
	GetGPS(ctx context.Context) (map[string]interface{}, error)
	GetAccelerometer(ctx context.Context) (map[string]interface{}, error)
	GetGyroscope(ctx context.Context) (map[string]interface{}, error)
	GetCompass(ctx context.Context) (map[string]interface{}, error)
	GetSensor(ctx context.Context, sensorType string) (map[string]interface{}, error)

	// HID
	ListHID(ctx context.Context) ([]map[string]interface{}, error)

	// Metadata
	DeviceMetadata(ctx context.Context, deviceID string) (map[string]interface{}, error)
}
