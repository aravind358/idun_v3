package communication

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

func (p *MockProvider) SendMessage(ctx context.Context, destination string, payload []byte) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock send error")
	}
	return "msg-12345", nil
}

func (p *MockProvider) ReceiveMessage(ctx context.Context, source string) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock receive error")
	}
	return []map[string]interface{}{{"id": "msg-1", "content": "hello"}}, nil
}

func (p *MockProvider) GetHistory(ctx context.Context, threadID string) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock history error")
	}
	return []map[string]interface{}{{"id": "msg-1", "content": "history"}}, nil
}

func (p *MockProvider) DeleteMessage(ctx context.Context, messageID string) error {
	if p.ShouldFail {
		return errors.New("mock delete error")
	}
	return nil
}

func (p *MockProvider) MarkRead(ctx context.Context, messageID string) error {
	if p.ShouldFail {
		return errors.New("mock mark read error")
	}
	return nil
}

func (p *MockProvider) MarkUnread(ctx context.Context, messageID string) error {
	if p.ShouldFail {
		return errors.New("mock mark unread error")
	}
	return nil
}

func (p *MockProvider) GetStatus(ctx context.Context, destination string) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock status error")
	}
	return "online", nil
}
