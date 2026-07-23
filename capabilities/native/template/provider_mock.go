package template

import "context"

type MockProvider struct {
	ShouldFail bool
}

func NewMockProvider(shouldFail bool) *MockProvider {
	return &MockProvider{ShouldFail: shouldFail}
}

// ExecuteExample provides a scaffold structure.
// TODO: Replace with mock logic matching the capability.
func (p *MockProvider) ExecuteExample(ctx context.Context) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, context.DeadlineExceeded
	}
	return map[string]interface{}{
		"provider": "mock",
	}, nil
}
