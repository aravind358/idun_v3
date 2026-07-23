package automation

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeClipboard(ctx context.Context, req capabilities.CapabilityRequest, op AutomationOperation) (map[string]interface{}, error) {
	switch op {
	case OperationClipboardRead:
		text, err := c.provider.ClipboardRead(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "read", "text": text}, nil

	case OperationClipboardWrite:
		text := req.Parameters["text"]
		if err := c.provider.ClipboardWrite(ctx, text); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "written"}, nil
	}
	return nil, nil
}
