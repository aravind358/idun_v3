package reminder

import (
	"errors"
	"strings"
)

// ReminderRequest is the strongly-typed internal representation of a reminder request.
type ReminderRequest struct {
	Operation    string
	ID           string
	Task         string
	Person       string
	Message      string
	Time         string
	Date         string
	Duration     string
	Priority     string
	TimeZone     string
	Repeat       string
	RepeatUntil  string
	Location     string
	NotifyBefore string
	Category     string
}

// BindReminderRequest converts generic map parameters into a typed ReminderRequest,
// performing centralized validation and normalization.
func BindReminderRequest(params map[string]string) (ReminderRequest, error) {
	req := ReminderRequest{
		Operation:    params["operation"],
		ID:           params["id"],
		Task:         params["task"],
		Person:       params["person"],
		Message:      params["message"],
		Time:         params["time"],
		Date:         params["date"],
		Duration:     params["duration"],
		Priority:     params["priority"],
		TimeZone:     params["timezone"],
		Repeat:       params["repeat"],
		RepeatUntil:  params["repeat_until"],
		Location:     params["location"],
		NotifyBefore: params["notify_before"],
		Category:     params["category"],
	}

	// 1. Validate Operation
	if req.Operation == "" {
		return req, errors.New("missing 'operation' parameter")
	}
	op := ReminderOperation(req.Operation)
	if !op.IsValid() {
		return req, errors.New("unsupported operation: " + req.Operation)
	}

	// 2. Normalize Message if missing
	if req.Message == "" {
		var parts []string
		if req.Task != "" {
			// Capitalize first letter of task (e.g. "call" -> "Call")
			task := req.Task
			if len(task) > 0 {
				task = strings.ToUpper(task[:1]) + task[1:]
			}
			parts = append(parts, task)
		}
		if req.Person != "" {
			parts = append(parts, req.Person)
		}
		
		req.Message = strings.Join(parts, " ")
		
		// If neither task nor person were provided and it's a set operation, it might be invalid
		if req.Operation == string(OperationSet) && req.Message == "" {
			return req, errors.New("missing required reminder content (task, person, or message)")
		}
	}

	// 3. Additional Validation based on operation
	if req.Operation == string(OperationSet) {
		if req.Time == "" && req.Date == "" {
			return req, errors.New("missing required reminder schedule (time or date)")
		}
	}

	return req, nil
}
