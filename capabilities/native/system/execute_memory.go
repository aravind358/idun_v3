package system

import "context"

func (c *Capability) executeMemory(ctx context.Context) (map[string]interface{}, error) {
	return c.provider.GetMemoryInfo(ctx)
}
