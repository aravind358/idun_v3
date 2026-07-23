package network

import (
	"context"
	"encoding/json"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeHTTP(ctx context.Context, req capabilities.CapabilityRequest, op NetworkOperation) (map[string]interface{}, error) {
	url := req.Parameters["url"]
	
	timeout := 30000 // default 30s
	if val, ok := req.Parameters["timeout_ms"]; ok {
		if t, err := strconv.Atoi(val); err == nil && t > 0 {
			timeout = t
		}
	}

	var headers map[string]string
	if val, ok := req.Parameters["headers"]; ok {
		json.Unmarshal([]byte(val), &headers)
	}

	var body []byte
	if val, ok := req.Parameters["body"]; ok {
		body = []byte(val)
	}

	method := "GET"
	switch op {
	case OperationHTTPPost:
		method = "POST"
	case OperationHTTPHead:
		method = "HEAD"
	}

	return c.provider.HTTPRequest(ctx, method, url, headers, body, timeout)
}
