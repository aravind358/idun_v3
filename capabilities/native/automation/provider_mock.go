package automation

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

func (p *MockProvider) MouseMove(ctx context.Context, x, y int) error {
	if p.ShouldFail {
		return errors.New("mock mouse move failure")
	}
	return nil
}

func (p *MockProvider) MouseClick(ctx context.Context, button string, clicks int) error {
	if p.ShouldFail {
		return errors.New("mock mouse click failure")
	}
	return nil
}

func (p *MockProvider) MouseScroll(ctx context.Context, deltaX, deltaY int) error {
	if p.ShouldFail {
		return errors.New("mock mouse scroll failure")
	}
	return nil
}

func (p *MockProvider) KeyboardPress(ctx context.Context, key string) error {
	if p.ShouldFail {
		return errors.New("mock keyboard press failure")
	}
	return nil
}

func (p *MockProvider) KeyboardRelease(ctx context.Context, key string) error {
	if p.ShouldFail {
		return errors.New("mock keyboard release failure")
	}
	return nil
}

func (p *MockProvider) KeyboardType(ctx context.Context, text string) error {
	if p.ShouldFail {
		return errors.New("mock keyboard type failure")
	}
	return nil
}

func (p *MockProvider) ClipboardRead(ctx context.Context) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock clipboard read failure")
	}
	return "mock clipboard content", nil
}

func (p *MockProvider) ClipboardWrite(ctx context.Context, text string) error {
	if p.ShouldFail {
		return errors.New("mock clipboard write failure")
	}
	return nil
}

func (p *MockProvider) CaptureScreen(ctx context.Context, region map[string]int) ([]byte, error) {
	if p.ShouldFail {
		return nil, errors.New("mock screen capture failure")
	}
	return []byte("mock image data"), nil
}

func (p *MockProvider) ListWindows(ctx context.Context) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock window list failure")
	}
	return []map[string]interface{}{{"handle": "win-1", "title": "Mock Window"}}, nil
}

func (p *MockProvider) GetWindow(ctx context.Context, handle string) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock window get failure")
	}
	return map[string]interface{}{"handle": handle, "title": "Mock Window"}, nil
}

func (p *MockProvider) FocusWindow(ctx context.Context, handle string) error {
	if p.ShouldFail {
		return errors.New("mock window focus failure")
	}
	return nil
}

func (p *MockProvider) MinimizeWindow(ctx context.Context, handle string) error {
	if p.ShouldFail {
		return errors.New("mock window minimize failure")
	}
	return nil
}

func (p *MockProvider) MaximizeWindow(ctx context.Context, handle string) error {
	if p.ShouldFail {
		return errors.New("mock window maximize failure")
	}
	return nil
}

func (p *MockProvider) RestoreWindow(ctx context.Context, handle string) error {
	if p.ShouldFail {
		return errors.New("mock window restore failure")
	}
	return nil
}

func (p *MockProvider) ListProcesses(ctx context.Context) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock process list failure")
	}
	return []map[string]interface{}{{"pid": 1234, "name": "mock_process.exe"}}, nil
}

func (p *MockProvider) GetProcess(ctx context.Context, pid int) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock process get failure")
	}
	return map[string]interface{}{"pid": pid, "name": "mock_process.exe"}, nil
}
