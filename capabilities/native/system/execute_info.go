package system

import "context"

func (c *Capability) executeInfo(ctx context.Context) (map[string]interface{}, error) {
	return c.provider.GetSystemInfo(ctx)
}
