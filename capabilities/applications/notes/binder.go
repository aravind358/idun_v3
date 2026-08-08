package notes

import (
	"errors"
)

// NotesRequest is the strongly-typed internal representation of a notes request.
type NotesRequest struct {
	Operation string
	Title     string
	Content   string
	Intent    string
}

// BindNotesRequest converts generic map parameters into a typed NotesRequest,
// performing centralized validation, normalization, and default assignment.
func BindNotesRequest(params map[string]string) (NotesRequest, error) {
	req := NotesRequest{
		Operation: params["operation"],
		Title:     params["title"],
		Content:   params["content"],
		Intent:    params["intent"],
	}

	// 1. Assign defaults
	if req.Intent == "" {
		req.Intent = "manage_notes"
	}

	// 2. Validate envelope
	if req.Operation == "" {
		return req, errors.New("missing 'operation' parameter")
	}

	op := NotesOperation(req.Operation)
	if !op.IsValid() {
		return req, errors.New("unsupported operation: " + req.Operation)
	}

	// 3. Domain validation
	switch op {
	case OperationCreate, OperationRead, OperationUpdate, OperationDelete:
		if req.Title == "" {
			return req, errors.New("missing 'title' parameter")
		}
	}

	return req, nil
}
