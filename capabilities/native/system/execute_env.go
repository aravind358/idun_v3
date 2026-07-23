package system

import "context"

func (c *Capability) executeEnv(ctx context.Context) (map[string]interface{}, error) {
	return c.provider.GetEnvInfo(ctx)
}
