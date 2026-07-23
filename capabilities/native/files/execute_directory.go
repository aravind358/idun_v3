package files

import (
	"context"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeDirectory(ctx context.Context, req capabilities.CapabilityRequest, op FileOperation) (map[string]interface{}, error) {
	path := req.Parameters["path"]

	switch op {
	case OperationListDirectory:
		recursive := false
		if b, err := strconv.ParseBool(req.Parameters["recursive"]); err == nil {
			recursive = b
		}
		list, err := c.provider.ListDirectory(ctx, path, recursive)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"path":  path,
			"items": list,
		}, nil

	case OperationCreateDirectory:
		err := c.provider.CreateDirectory(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success", "path": path}, nil

	case OperationDeleteDirectory:
		err := c.provider.DeleteDirectory(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success", "path": path}, nil
		
	case OperationTemporaryDirectory:
		prefix := req.Parameters["prefix"]
		tempPath, err := c.provider.CreateTemporaryDirectory(ctx, prefix)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success", "path": tempPath}, nil
	}

	return nil, nil
}
