package network

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeDNS(ctx context.Context, req capabilities.CapabilityRequest, op NetworkOperation) (map[string]interface{}, error) {
	switch op {
	case OperationResolveDNS:
		hostname := req.Parameters["hostname"]
		ips, err := c.provider.ResolveDNS(ctx, hostname)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"hostname": hostname,
			"ips":      ips,
		}, nil

	case OperationLookupIP:
		ip := req.Parameters["ip"]
		hostnames, err := c.provider.LookupIP(ctx, ip)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"ip":        ip,
			"hostnames": hostnames,
		}, nil
	}
	return nil, nil
}
