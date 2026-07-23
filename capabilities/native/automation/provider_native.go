package automation

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

func (p *NativeProvider) MouseMove(ctx context.Context, x, y int) error {
	return errors.New("native mouse movement requires external bindings")
}

func (p *NativeProvider) MouseClick(ctx context.Context, button string, clicks int) error {
	return errors.New("native mouse clicking requires external bindings")
}

func (p *NativeProvider) MouseScroll(ctx context.Context, deltaX, deltaY int) error {
	return errors.New("native mouse scrolling requires external bindings")
}

func (p *NativeProvider) KeyboardPress(ctx context.Context, key string) error {
	return errors.New("native keyboard press requires external bindings")
}

func (p *NativeProvider) KeyboardRelease(ctx context.Context, key string) error {
	return errors.New("native keyboard release requires external bindings")
}

func (p *NativeProvider) KeyboardType(ctx context.Context, text string) error {
	return errors.New("native keyboard typing requires external bindings")
}

func (p *NativeProvider) ClipboardRead(ctx context.Context) (string, error) {
	return "", errors.New("native clipboard reading requires external bindings")
}

func (p *NativeProvider) ClipboardWrite(ctx context.Context, text string) error {
	return errors.New("native clipboard writing requires external bindings")
}

func (p *NativeProvider) CaptureScreen(ctx context.Context, region map[string]int) ([]byte, error) {
	return nil, errors.New("native screen capture requires external bindings")
}

func (p *NativeProvider) ListWindows(ctx context.Context) ([]map[string]interface{}, error) {
	return nil, errors.New("native window enumeration requires external bindings")
}

func (p *NativeProvider) GetWindow(ctx context.Context, handle string) (map[string]interface{}, error) {
	return nil, errors.New("native window lookup requires external bindings")
}

func (p *NativeProvider) FocusWindow(ctx context.Context, handle string) error {
	return errors.New("native window focus requires external bindings")
}

func (p *NativeProvider) MinimizeWindow(ctx context.Context, handle string) error {
	return errors.New("native window minimize requires external bindings")
}

func (p *NativeProvider) MaximizeWindow(ctx context.Context, handle string) error {
	return errors.New("native window maximize requires external bindings")
}

func (p *NativeProvider) RestoreWindow(ctx context.Context, handle string) error {
	return errors.New("native window restore requires external bindings")
}

func (p *NativeProvider) ListProcesses(ctx context.Context) ([]map[string]interface{}, error) {
	return nil, errors.New("native process enumeration requires external bindings")
}

func (p *NativeProvider) GetProcess(ctx context.Context, pid int) (map[string]interface{}, error) {
	return nil, errors.New("native process lookup requires external bindings")
}
