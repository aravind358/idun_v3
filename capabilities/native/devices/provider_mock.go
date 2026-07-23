package devices

import (
	"context"
	"errors"
)

type MockProvider struct {
	ShouldFail bool
}

func NewMockProvider(shouldFail bool) *MockProvider {
	return &MockProvider{ShouldFail: shouldFail}
}

func (p *MockProvider) ListUSBDevices(ctx context.Context) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock usb enumeration failure")
	}
	return []map[string]interface{}{{"id": "usb-1", "name": "Mock USB Drive"}}, nil
}

func (p *MockProvider) GetUSBDevice(ctx context.Context, deviceID string) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock usb get failure")
	}
	return map[string]interface{}{"id": deviceID, "name": "Mock USB Device"}, nil
}

func (p *MockProvider) ListBluetoothDevices(ctx context.Context) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock bluetooth enumeration failure")
	}
	return []map[string]interface{}{{"id": "bt-1", "name": "Mock Bluetooth Speaker"}}, nil
}

func (p *MockProvider) PairBluetooth(ctx context.Context, deviceID string) error {
	if p.ShouldFail {
		return errors.New("mock bluetooth pairing failure")
	}
	return nil
}

func (p *MockProvider) UnpairBluetooth(ctx context.Context, deviceID string) error {
	if p.ShouldFail {
		return errors.New("mock bluetooth unpairing failure")
	}
	return nil
}

func (p *MockProvider) BatteryStatus(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock battery status failure")
	}
	return map[string]interface{}{"percentage": 85, "charging": true}, nil
}

func (p *MockProvider) PowerStatus(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock power status failure")
	}
	return map[string]interface{}{"ac_power": true}, nil
}

func (p *MockProvider) GetGPS(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock gps failure")
	}
	return map[string]interface{}{"latitude": 37.7749, "longitude": -122.4194}, nil
}

func (p *MockProvider) GetAccelerometer(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock accelerometer failure")
	}
	return map[string]interface{}{"x": 0.0, "y": 9.8, "z": 0.0}, nil
}

func (p *MockProvider) GetGyroscope(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock gyroscope failure")
	}
	return map[string]interface{}{"x": 0.1, "y": 0.0, "z": -0.1}, nil
}

func (p *MockProvider) GetCompass(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock compass failure")
	}
	return map[string]interface{}{"heading": 180.0}, nil
}

func (p *MockProvider) GetSensor(ctx context.Context, sensorType string) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock sensor failure")
	}
	return map[string]interface{}{"type": sensorType, "value": 42}, nil
}

func (p *MockProvider) ListHID(ctx context.Context) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock hid enumeration failure")
	}
	return []map[string]interface{}{{"id": "hid-1", "name": "Mock Keyboard"}}, nil
}

func (p *MockProvider) DeviceMetadata(ctx context.Context, deviceID string) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock metadata failure")
	}
	return map[string]interface{}{"id": deviceID, "manufacturer": "Mock Corp"}, nil
}
