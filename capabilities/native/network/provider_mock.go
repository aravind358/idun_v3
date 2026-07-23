package network

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

func (p *MockProvider) ResolveDNS(ctx context.Context, hostname string) ([]string, error) {
	if p.ShouldFail {
		return nil, errors.New("mock dns failure")
	}
	return []string{"192.168.1.1"}, nil
}

func (p *MockProvider) LookupIP(ctx context.Context, ip string) ([]string, error) {
	if p.ShouldFail {
		return nil, errors.New("mock ip lookup failure")
	}
	return []string{"mock.local"}, nil
}

func (p *MockProvider) HTTPRequest(ctx context.Context, method, url string, headers map[string]string, body []byte, timeoutMs int) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock http failure")
	}
	return map[string]interface{}{
		"status_code": 200,
		"headers":     map[string]string{"Content-Type": "application/json"},
		"body":        []byte(`{"mock":"response"}`),
	}, nil
}

func (p *MockProvider) Download(ctx context.Context, url, destination string, timeoutMs int) error {
	if p.ShouldFail {
		return errors.New("mock download failure")
	}
	return nil
}

func (p *MockProvider) Upload(ctx context.Context, url, source string, timeoutMs int) error {
	if p.ShouldFail {
		return errors.New("mock upload failure")
	}
	return nil
}

func (p *MockProvider) OpenTCPSocket(ctx context.Context, address string, timeoutMs int) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock tcp open failure")
	}
	return "mock-tcp-1", nil
}

func (p *MockProvider) OpenUDPSocket(ctx context.Context, address string, timeoutMs int) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock udp open failure")
	}
	return "mock-udp-1", nil
}

func (p *MockProvider) CloseSocket(ctx context.Context, socketID string) error {
	if p.ShouldFail {
		return errors.New("mock socket close failure")
	}
	return nil
}

func (p *MockProvider) ListInterfaces(ctx context.Context) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock interface list failure")
	}
	return []map[string]interface{}{{"name": "eth0", "index": 1}}, nil
}

func (p *MockProvider) GetInterface(ctx context.Context, name string) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock interface get failure")
	}
	return map[string]interface{}{"name": "eth0", "index": 1}, nil
}

func (p *MockProvider) ConnectionStatus(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock status failure")
	}
	return map[string]interface{}{"connected": true}, nil
}

func (p *MockProvider) Ping(ctx context.Context, address string, timeoutMs int) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock ping failure")
	}
	return map[string]interface{}{"reachable": true, "latency": 10}, nil
}
