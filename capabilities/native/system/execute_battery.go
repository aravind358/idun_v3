package system

import "context"

func (c *Capability) executeBattery(ctx context.Context) (map[string]interface{}, error) {
	return c.provider.GetBatteryInfo(ctx)
}
