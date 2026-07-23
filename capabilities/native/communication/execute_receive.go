package communication

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeReceive(ctx context.Context, req capabilities.CapabilityRequest, op CommunicationOperation) (map[string]interface{}, error) {
	switch op {
	case OperationReceiveMessage:
		source := req.Parameters["source"]
		msgs, err := c.provider.ReceiveMessage(ctx, source)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"messages": msgs,
		}, nil

	case OperationGetStatus:
		destination := req.Parameters["destination"]
		status, err := c.provider.GetStatus(ctx, destination)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"destination": destination,
			"status":      status,
		}, nil
	}

	return nil, nil
}
