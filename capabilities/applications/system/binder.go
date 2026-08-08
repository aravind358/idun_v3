package system

import (
	"errors"
	"time"
)

// SystemRequest is the strongly-typed internal representation of a system request.
type SystemRequest struct {
	Intent        string
	Operation     string
	OriginalInput string

	// Reserved for future growth
	Target  string
	Device  string
	Process string
	Service string

	Force   bool
	Timeout time.Duration
}

// BindSystemRequest converts generic map parameters into a typed SystemRequest,
// performing centralized validation, precedence extraction, and normalization.
func BindSystemRequest(params map[string]string) (SystemRequest, error) {
	req := SystemRequest{
		Intent:        params["intent"],
		Operation:     params["operation"],
		OriginalInput: params["raw_input"],
	}

	// 1. Validation
	if req.Intent == "" {
		return req, errors.New("missing 'intent' parameter")
	}

	// 2. Defaults
	if req.OriginalInput == "" {
		req.OriginalInput = req.Operation
	}

	return req, nil
}
