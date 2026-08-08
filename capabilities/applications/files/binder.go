package files

import (
	"errors"
	"strings"
)

// FileRequest is the strongly-typed internal representation of a file capability request.
type FileRequest struct {
	Operation   string
	Target      string
	Destination string
	Context     string
	Intent      string
}

// BindFileRequest converts generic map parameters into a typed FileRequest,
// performing centralized validation, precedence extraction, and semantic normalization.
func BindFileRequest(params map[string]string) (FileRequest, error) {
	req := FileRequest{
		Operation:   params["operation"],
		Destination: params["destination"],
		Context:     params["command"], // Mapped from raw command context
		Intent:      params["intent"],
	}

	// 1. Validation of required inputs
	if req.Operation == "" {
		return req, errors.New("missing 'operation' parameter")
	}

	// 2. Precedence logic for target extraction
	if path := params["path"]; path != "" {
		req.Target = path
	} else if filename := params["filename"]; filename != "" {
		req.Target = filename
	} else if directory := params["directory"]; directory != "" {
		req.Target = directory
	} else if source := params["source"]; source != "" {
		req.Target = source
	}

	// 3. Normalization
	normalizedOp := strings.ToLower(req.Operation)
	if normalizedOp == "check if" || normalizedOp == "check" || normalizedOp == "does" {
		normalizedOp = "exists"
	} else if normalizedOp == "remove" {
		normalizedOp = "delete"
	}
	req.Operation = normalizedOp

	if req.Intent == "" {
		req.Intent = "manage_files"
	}

	return req, nil
}
