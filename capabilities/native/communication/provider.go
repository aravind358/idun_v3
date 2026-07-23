package communication

import "context"

// CommunicationProvider abstracts native communication operations.
// This isolates the capability from host APIs and enables seamless mock testing.
type CommunicationProvider interface {
	SendMessage(ctx context.Context, destination string, payload []byte) (string, error)
	ReceiveMessage(ctx context.Context, source string) ([]map[string]interface{}, error)
	GetHistory(ctx context.Context, threadID string) ([]map[string]interface{}, error)
	DeleteMessage(ctx context.Context, messageID string) error
	MarkRead(ctx context.Context, messageID string) error
	MarkUnread(ctx context.Context, messageID string) error
	GetStatus(ctx context.Context, destination string) (string, error)
}
