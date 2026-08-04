package system

import "context"

type MockProvider struct {
	ShouldFail bool
}

func NewMockProvider(shouldFail bool) *MockProvider {
	return &MockProvider{ShouldFail: shouldFail}
}

func (p *MockProvider) GetSystemInfo(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, context.DeadlineExceeded
	}
	return map[string]interface{}{"os": "mock_os", "time": "2026-07-23T12:00:00Z"}, nil
}

func (p *MockProvider) GetEnvInfo(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, context.DeadlineExceeded
	}
	return map[string]interface{}{"environment": []string{"MOCK_ENV=true"}}, nil
}

func (p *MockProvider) GetHostInfo(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, context.DeadlineExceeded
	}
	return map[string]interface{}{"hostname": "mock_host", "uptime": "100h"}, nil
}

func (p *MockProvider) GetCPUInfo(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, context.DeadlineExceeded
	}
	return map[string]interface{}{"cpu_cores": 8}, nil
}

func (p *MockProvider) GetMemoryInfo(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, context.DeadlineExceeded
	}
	return map[string]interface{}{"memory": "16GB"}, nil
}

func (p *MockProvider) GetDiskInfo(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, context.DeadlineExceeded
	}
	return map[string]interface{}{"disk": "512GB"}, nil
}

func (p *MockProvider) ExecutePower(ctx context.Context, action SystemOperation) error {
	if p.ShouldFail {
		return context.DeadlineExceeded
	}
	return nil
}



func (p *MockProvider) GetBatteryInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"status": "Charging", "percentage": 100}, nil
}
