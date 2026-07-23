package system

import "context"

func (c *Capability) executeCPU(ctx context.Context) (map[string]interface{}, error) {
	return c.provider.GetCPUInfo(ctx)
}
