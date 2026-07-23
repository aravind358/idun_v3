package automation

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeKeyboard(ctx context.Context, req capabilities.CapabilityRequest, op AutomationOperation) (map[string]interface{}, error) {
	switch op {
	case OperationKeyboardPress:
		key := req.Parameters["key"]
		if err := c.provider.KeyboardPress(ctx, key); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "pressed", "key": key}, nil

	case OperationKeyboardRelease:
		key := req.Parameters["key"]
		if err := c.provider.KeyboardRelease(ctx, key); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "released", "key": key}, nil

	case OperationKeyboardType:
		text := req.Parameters["text"]
		if err := c.provider.KeyboardType(ctx, text); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "typed", "text_length": len(text)}, nil
	}
	return nil, nil
}
