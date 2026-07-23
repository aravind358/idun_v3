package devices

import (
	"context"
	"errors"
	"sync"
)

type NativeProvider struct {
	mu sync.Mutex
}

func NewNativeProvider() *NativeProvider {
	return &NativeProvider{}
}

func (p *NativeProvider) ListUSBDevices(ctx context.Context) ([]map[string]interface{}, error) {
	return nil, errors.New("native usb enumeration requires external bindings")
}

func (p *NativeProvider) GetUSBDevice(ctx context.Context, deviceID string) (map[string]interface{}, error) {
	return nil, errors.New("native usb device lookup requires external bindings")
}

func (p *NativeProvider) ListBluetoothDevices(ctx context.Context) ([]map[string]interface{}, error) {
	return nil, errors.New("native bluetooth enumeration requires external bindings")
}

func (p *NativeProvider) PairBluetooth(ctx context.Context, deviceID string) error {
	return errors.New("native bluetooth pairing requires external bindings")
}

func (p *NativeProvider) UnpairBluetooth(ctx context.Context, deviceID string) error {
	return errors.New("native bluetooth unpairing requires external bindings")
}

func (p *NativeProvider) BatteryStatus(ctx context.Context) (map[string]interface{}, error) {
	return nil, errors.New("native battery status requires external bindings")
}

func (p *NativeProvider) PowerStatus(ctx context.Context) (map[string]interface{}, error) {
	return nil, errors.New("native power status requires external bindings")
}

func (p *NativeProvider) GetGPS(ctx context.Context) (map[string]interface{}, error) {
	return nil, errors.New("native gps access requires external bindings")
}

func (p *NativeProvider) GetAccelerometer(ctx context.Context) (map[string]interface{}, error) {
	return nil, errors.New("native accelerometer access requires external bindings")
}

func (p *NativeProvider) GetGyroscope(ctx context.Context) (map[string]interface{}, error) {
	return nil, errors.New("native gyroscope access requires external bindings")
}

func (p *NativeProvider) GetCompass(ctx context.Context) (map[string]interface{}, error) {
	return nil, errors.New("native compass access requires external bindings")
}

func (p *NativeProvider) GetSensor(ctx context.Context, sensorType string) (map[string]interface{}, error) {
	return nil, errors.New("native generic sensor access requires external bindings")
}

func (p *NativeProvider) ListHID(ctx context.Context) ([]map[string]interface{}, error) {
	return nil, errors.New("native hid enumeration requires external bindings")
}

func (p *NativeProvider) DeviceMetadata(ctx context.Context, deviceID string) (map[string]interface{}, error) {
	return nil, errors.New("native device metadata access requires external bindings")
}
