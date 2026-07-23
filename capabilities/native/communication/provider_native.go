package communication

import (
	"context"
	"errors"
)

type NativeProvider struct{}

func NewNativeProvider() *NativeProvider {
	return &NativeProvider{}
}

func (p *NativeProvider) SendMessage(ctx context.Context, destination string, payload []byte) (string, error) {
	// Native transport stub. In a real system, this would interface with IPC, domain sockets, or named pipes.
	return "", errors.New("native communication transport not configured")
}

func (p *NativeProvider) ReceiveMessage(ctx context.Context, source string) ([]map[string]interface{}, error) {
	return nil, errors.New("native communication transport not configured")
}

func (p *NativeProvider) GetHistory(ctx context.Context, threadID string) ([]map[string]interface{}, error) {
	return nil, errors.New("native communication transport not configured")
}

func (p *NativeProvider) DeleteMessage(ctx context.Context, messageID string) error {
	return errors.New("native communication transport not configured")
}

func (p *NativeProvider) MarkRead(ctx context.Context, messageID string) error {
	return errors.New("native communication transport not configured")
}

func (p *NativeProvider) MarkUnread(ctx context.Context, messageID string) error {
	return errors.New("native communication transport not configured")
}

func (p *NativeProvider) GetStatus(ctx context.Context, destination string) (string, error) {
	return "", errors.New("native communication transport not configured")
}
