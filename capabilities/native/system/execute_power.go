package system

import "context"

func (c *Capability) executePower(ctx context.Context, action SystemOperation) (map[string]interface{}, error) {
	err := c.provider.ExecutePower(ctx, action)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status": "success",
		"action": string(action),
	}, nil
}
