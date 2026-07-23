package network

import (
	"context"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeInterfaces(ctx context.Context, req capabilities.CapabilityRequest, op NetworkOperation) (map[string]interface{}, error) {
	switch op {
	case OperationListInterfaces:
		ifaces, err := c.provider.ListInterfaces(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"interfaces": ifaces,
		}, nil

	case OperationGetInterface:
		name := req.Parameters["name"]
		iface, err := c.provider.GetInterface(ctx, name)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"interface": iface,
		}, nil

	case OperationConnectionStatus:
		status, err := c.provider.ConnectionStatus(ctx)
		if err != nil {
			return nil, err
		}
		return status, nil

	case OperationPing:
		address := req.Parameters["address"]
		timeout := 5000 // default 5s
		if val, ok := req.Parameters["timeout_ms"]; ok {
			if t, err := strconv.Atoi(val); err == nil && t > 0 {
				timeout = t
			}
		}
		return c.provider.Ping(ctx, address, timeout)
	}

	return nil, nil
}
