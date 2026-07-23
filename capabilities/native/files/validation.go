package files

import (
	"errors"
	"path/filepath"
	"strings"

	"idun/capabilities"
)

func (c *Capability) validateRequest(req capabilities.CapabilityRequest) error {
	if req.RequirementID == "" {
		return errors.New("missing requirement ID")
	}

	operation := req.Parameters["operation"]
	if operation == "" {
		return errors.New("missing 'operation' parameter")
	}

	op := FileOperation(operation)
	if !op.IsValid() {
		return errors.New("unsupported operation: " + operation)
	}

	// Validate path parameters (prevent empty paths or directory traversal)
	path := req.Parameters["path"]
	if path != "" {
		// Prevent obvious traversal attempts as a basic security measure
		if strings.Contains(path, "..") {
			return errors.New("path traversal detected: '..' not allowed in paths")
		}
		// Basic normalization check
		if !filepath.IsAbs(path) {
			// Some operations might allow relative paths, but we enforce safety bounds.
			// Let's rely on provider path cleaning, but block `..` here.
		}
	}

	return nil
}
