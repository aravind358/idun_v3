package communication

import (
	"context"
	"encoding/base64"

	"idun/capabilities"
)

func (c *Capability) executeSend(ctx context.Context, req capabilities.CapabilityRequest) (map[string]interface{}, error) {
	destination := req.Parameters["destination"]
	
	var payload []byte
	if text, ok := req.Parameters["payload_text"]; ok {
		payload = []byte(text)
	} else if b64, ok := req.Parameters["payload_bytes"]; ok {
		var err error
		payload, err = base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, err
		}
	}

	msgID, err := c.provider.SendMessage(ctx, destination, payload)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":     "sent",
		"message_id": msgID,
	}, nil
}
