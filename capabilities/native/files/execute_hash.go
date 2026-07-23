package files

import (
	"context"

	"idun/capabilities"
)

func (c *Capability) executeHash(ctx context.Context, req capabilities.CapabilityRequest, op FileOperation) (map[string]interface{}, error) {
	path := req.Parameters["path"]
	algorithm := req.Parameters["algorithm"]
	if algorithm == "" {
		algorithm = "sha256" // default
	}

	switch op {
	case OperationCalculateHash:
		hash, err := c.provider.CalculateHash(ctx, path, algorithm)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"path":      path,
			"algorithm": algorithm,
			"hash":      hash,
		}, nil
	}

	return nil, nil
}
