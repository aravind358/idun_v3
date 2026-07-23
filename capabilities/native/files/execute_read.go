package files

import (
	"context"
	"encoding/base64"

	"idun/capabilities"
)

func (c *Capability) executeRead(ctx context.Context, req capabilities.CapabilityRequest, op FileOperation) (map[string]interface{}, error) {
	path := req.Parameters["path"]

	switch op {
	case OperationReadFile, OperationReadBytes:
		data, err := c.provider.ReadFile(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"path": path,
			"data": base64.StdEncoding.EncodeToString(data), // base64 for bytes
		}, nil

	case OperationReadText:
		text, err := c.provider.ReadText(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"path": path,
			"text": text,
		}, nil

	case OperationFileExists:
		exists, err := c.provider.FileExists(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"path":   path,
			"exists": exists,
		}, nil
	}

	return nil, nil
}
