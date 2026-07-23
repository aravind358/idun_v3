package communication

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeHistory(ctx context.Context, req capabilities.CapabilityRequest, op CommunicationOperation) (map[string]interface{}, error) {
	switch op {
	case OperationGetHistory:
		threadID := req.Parameters["thread_id"]
		history, err := c.provider.GetHistory(ctx, threadID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"thread_id": threadID,
			"history":   history,
		}, nil

	case OperationDeleteMessage:
		messageID := req.Parameters["message_id"]
		err := c.provider.DeleteMessage(ctx, messageID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":     "deleted",
			"message_id": messageID,
		}, nil

	case OperationMarkRead:
		messageID := req.Parameters["message_id"]
		err := c.provider.MarkRead(ctx, messageID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":     "marked_read",
			"message_id": messageID,
		}, nil

	case OperationMarkUnread:
		messageID := req.Parameters["message_id"]
		err := c.provider.MarkUnread(ctx, messageID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":     "marked_unread",
			"message_id": messageID,
		}, nil
	}

	return nil, nil
}
