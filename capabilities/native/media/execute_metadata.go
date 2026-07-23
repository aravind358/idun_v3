package media

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeMetadata(ctx context.Context, req capabilities.CapabilityRequest, op MediaOperation) (map[string]interface{}, error) {
	switch op {
	case OperationGetMetadata:
		path := req.Parameters["path"]
		metadata, err := c.provider.GetMetadata(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"path":     path,
			"metadata": metadata,
		}, nil
	}

	return nil, nil
}
