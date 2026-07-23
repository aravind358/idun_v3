package automation

import (
	"errors"

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

	op := AutomationOperation(operation)
	if !op.IsValid() {
		return errors.New("unsupported operation: " + operation)
	}

	return nil
}
