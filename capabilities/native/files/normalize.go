package files

import (
	"time"

	"idun/capabilities"
)

func (c *Capability) normalizeResult(reqID string, start time.Time, data map[string]interface{}) capabilities.CapabilityResult {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       true,
		Data:          data,
		Duration:      time.Since(start),
	}
}

func (c *Capability) normalizeError(reqID string, start time.Time, code string, err error) (capabilities.CapabilityResult, error) {
	capErr := &capabilities.CapabilityError{
		Code:    code,
		Message: err.Error(),
		Retry:   false,
	}

	if code == "Timeout" || code == "Unavailable" {
		capErr.Retry = true
	}

	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       false,
		Error:         capErr,
		Duration:      time.Since(start),
	}, nil
}
