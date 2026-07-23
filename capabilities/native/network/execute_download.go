package network

import (
	"context"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeDownload(ctx context.Context, req capabilities.CapabilityRequest, op NetworkOperation) (map[string]interface{}, error) {
	url := req.Parameters["url"]
	timeout := 60000 // default 60s
	if val, ok := req.Parameters["timeout_ms"]; ok {
		if t, err := strconv.Atoi(val); err == nil && t > 0 {
			timeout = t
		}
	}

	switch op {
	case OperationDownload:
		destination := req.Parameters["destination"]
		err := c.provider.Download(ctx, url, destination, timeout)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":      "downloaded",
			"destination": destination,
		}, nil

	case OperationUpload:
		source := req.Parameters["source"]
		err := c.provider.Upload(ctx, url, source, timeout)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status": "uploaded",
			"source": source,
		}, nil
	}

	return nil, nil
}
