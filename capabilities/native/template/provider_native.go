package template

import "context"

type NativeProvider struct{}

func NewNativeProvider() *NativeProvider {
	return &NativeProvider{}
}

// ExecuteExample provides a scaffold structure.
// TODO: Replace with actual platform-specific implementation. (Hint: use build tags if needed)
func (p *NativeProvider) ExecuteExample(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"provider": "native",
	}, nil
}
