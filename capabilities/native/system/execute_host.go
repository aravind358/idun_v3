package system

import "context"

func (c *Capability) executeHost(ctx context.Context) (map[string]interface{}, error) {
	return c.provider.GetHostInfo(ctx)
}
