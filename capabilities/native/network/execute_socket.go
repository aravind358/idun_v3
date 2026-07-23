package network

import (
	"context"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeSocket(ctx context.Context, req capabilities.CapabilityRequest, op NetworkOperation) (map[string]interface{}, error) {
	switch op {
	case OperationOpenTCPSocket, OperationOpenUDPSocket:
		address := req.Parameters["address"]
		timeout := 5000 // default 5s
		if val, ok := req.Parameters["timeout_ms"]; ok {
			if t, err := strconv.Atoi(val); err == nil && t > 0 {
				timeout = t
			}
		}

		var id string
		var err error
		if op == OperationOpenTCPSocket {
			id, err = c.provider.OpenTCPSocket(ctx, address, timeout)
		} else {
			id, err = c.provider.OpenUDPSocket(ctx, address, timeout)
		}

		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"socket_id": id,
		}, nil

	case OperationCloseSocket:
		socketID := req.Parameters["socket_id"]
		err := c.provider.CloseSocket(ctx, socketID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":    "closed",
			"socket_id": socketID,
		}, nil
	}

	return nil, nil
}
