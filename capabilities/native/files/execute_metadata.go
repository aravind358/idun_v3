package files

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeMetadata(ctx context.Context, req capabilities.CapabilityRequest, op FileOperation) (map[string]interface{}, error) {
	path := req.Parameters["path"]

	switch op {
	case OperationFileMetadata, OperationDirectoryMetadata:
		meta, err := c.provider.GetMetadata(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"path":     path,
			"metadata": meta,
		}, nil
	}

	return nil, nil
}
