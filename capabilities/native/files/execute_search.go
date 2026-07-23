package files

import (
	"context"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeSearch(ctx context.Context, req capabilities.CapabilityRequest, op FileOperation) (map[string]interface{}, error) {
	root := req.Parameters["root"]
	if root == "" {
		root = req.Parameters["path"] // fallback if they use "path"
	}
	pattern := req.Parameters["pattern"]
	
	recursive := false
	if b, err := strconv.ParseBool(req.Parameters["recursive"]); err == nil {
		recursive = b
	}
	
	caseSensitive := false
	if b, err := strconv.ParseBool(req.Parameters["case_sensitive"]); err == nil {
		caseSensitive = b
	}

	switch op {
	case OperationSearchFiles:
		matches, err := c.provider.SearchFiles(ctx, root, pattern, recursive, caseSensitive)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"root":    root,
			"pattern": pattern,
			"matches": matches,
		}, nil
	}

	return nil, nil
}
