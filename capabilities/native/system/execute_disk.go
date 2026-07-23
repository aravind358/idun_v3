package system

import "context"

func (c *Capability) executeDisk(ctx context.Context) (map[string]interface{}, error) {
	return c.provider.GetDiskInfo(ctx)
}
