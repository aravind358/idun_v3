package notes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"idun/capabilities"
	"idun/capabilities/applications/core"
)

// Capability defines the notes application capability.
// It is a Model 1 App Capability (Orchestration).
type Capability struct {
	core.AppCapability
}

// New creates a new instance of the Notes Capability.
func New(deps core.AppCapabilityDependencies) *Capability {
	return &Capability{
		AppCapability: core.NewAppCapability("app-notes-1", Metadata(), deps.Resolver),
	}
}

// Execute fulfills the Capability interface.
func (c *Capability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	start := time.Now()

	// 1. Validation of envelope
	if req.RequirementID == "" {
		return c.normalizeError(req.RequirementID, start, "Validation", errors.New("missing requirement ID"))
	}

	// 2. Lifecycle Check
	if err := c.checkLifecycle(); err != nil {
		return c.normalizeError(req.RequirementID, start, "Unavailable", err)
	}

	// 3. Binding & Validation
	typedReq, err := BindNotesRequest(req.Parameters)
	if err != nil {
		return c.normalizeError(req.RequirementID, start, "Validation", err)
	}

	// 4. Execution
	var data map[string]interface{}
	var execErr error

	op := NotesOperation(typedReq.Operation)
	switch op {
	case OperationCreate:
		data, execErr = c.executeCreate(ctx, req.RequirementID, typedReq)
	case OperationRead:
		data, execErr = c.executeRead(ctx, req.RequirementID, typedReq)
	case OperationUpdate:
		data, execErr = c.executeUpdate(ctx, req.RequirementID, typedReq)
	case OperationDelete:
		data, execErr = c.executeDelete(ctx, req.RequirementID, typedReq)
	case OperationList:
		data, execErr = c.executeList(ctx, req.RequirementID)
	default:
		execErr = fmt.Errorf("operation %s not implemented", typedReq.Operation)
	}

	if execErr != nil {
		return c.normalizeError(req.RequirementID, start, "Execution", execErr)
	}

	return c.normalizeResult(req.RequirementID, start, typedReq.Intent, data), nil
}



func (c *Capability) checkLifecycle() error {
	state := c.State().Lifecycle
	if state == "DISABLED" || state == "UNLOADED" {
		return errors.New("capability is not currently available for execution: " + string(state))
	}
	return nil
}

func (c *Capability) executeCreate(ctx context.Context, reqID string, req NotesRequest) (map[string]interface{}, error) {
	if c.Resolver == nil {
		return nil, errors.New("native capability resolver is not wired")
	}

	// Check if note already exists
	existsParams := map[string]string{
		"operation": "file_exists",
		"filename":  req.Title, // Let Native Files determine format/path logic
	}
	
	existsCap, err := c.Resolver.Resolve(ctx, reqID, "NativeFilesCapability", existsParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve native files capability: %w", err)
	}
	
	existsRes, _ := existsCap.Execute(ctx, capabilities.CapabilityRequest{
		RequirementID: reqID + "-exists",
		Parameters:    existsParams,
		ContextID:     reqID,
	})
	
	if existsRes.Success && existsRes.Data["exists"] == true {
		return nil, errors.New("note already exists and cannot be overwritten")
	}

	fileParams := map[string]string{
		"operation": "write_file",
		"filename":  req.Title,
		"data_text": req.Content,
		"append":    "false",
	}

	fileCap, err := c.Resolver.Resolve(ctx, reqID, "NativeFilesCapability", fileParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve native files capability: %w", err)
	}

	res, execErr := fileCap.Execute(ctx, capabilities.CapabilityRequest{
		RequirementID: reqID + "-sub",
		Parameters:    fileParams,
		ContextID:     reqID,
	})
	if execErr != nil {
		return nil, fmt.Errorf("native files execution failed: %w", execErr)
	}
	if !res.Success {
		return nil, fmt.Errorf("native files request failed: %v", res.Error.Message)
	}

	return map[string]interface{}{
		"status": "created",
		"title":  req.Title,
	}, nil
}

func (c *Capability) executeRead(ctx context.Context, reqID string, req NotesRequest) (map[string]interface{}, error) {
	if c.Resolver == nil {
		return nil, errors.New("native capability resolver is not wired")
	}

	fileParams := map[string]string{
		"operation": "read_text",
		"filename":  req.Title,
	}

	fileCap, err := c.Resolver.Resolve(ctx, reqID, "NativeFilesCapability", fileParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve native files capability: %w", err)
	}

	res, execErr := fileCap.Execute(ctx, capabilities.CapabilityRequest{
		RequirementID: reqID + "-sub",
		Parameters:    fileParams,
		ContextID:     reqID,
	})
	if execErr != nil {
		return nil, fmt.Errorf("native files execution failed: %w", execErr)
	}
	if !res.Success {
		return nil, fmt.Errorf("native files request failed: %v", res.Error.Message)
	}

	return map[string]interface{}{
		"status":  "found",
		"title":   req.Title,
		"content": res.Data["text"],
	}, nil
}

func (c *Capability) executeUpdate(ctx context.Context, reqID string, req NotesRequest) (map[string]interface{}, error) {
	// Re-use executeCreate since it overwrites
	res, err := c.executeCreate(ctx, reqID, req)
	if err != nil {
		return nil, err
	}
	res["status"] = "updated"
	return res, nil
}

func (c *Capability) executeDelete(ctx context.Context, reqID string, req NotesRequest) (map[string]interface{}, error) {
	if c.Resolver == nil {
		return nil, errors.New("native capability resolver is not wired")
	}

	fileParams := map[string]string{
		"operation": "delete_file",
		"filename":  req.Title,
	}

	fileCap, err := c.Resolver.Resolve(ctx, reqID, "NativeFilesCapability", fileParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve native files capability: %w", err)
	}

	res, execErr := fileCap.Execute(ctx, capabilities.CapabilityRequest{
		RequirementID: reqID + "-sub",
		Parameters:    fileParams,
		ContextID:     reqID,
	})
	if execErr != nil {
		return nil, fmt.Errorf("native files execution failed: %w", execErr)
	}
	if !res.Success {
		return nil, fmt.Errorf("native files request failed: %v", res.Error.Message)
	}

	return map[string]interface{}{
		"status": "deleted",
		"title":  req.Title,
	}, nil
}

func (c *Capability) executeList(ctx context.Context, reqID string) (map[string]interface{}, error) {
	if c.Resolver == nil {
		return nil, errors.New("native capability resolver is not wired")
	}

	fileParams := map[string]string{
		"operation": "list_directory",
		"path":      ".", // Current directory since it's an implementation detail
	}

	fileCap, err := c.Resolver.Resolve(ctx, reqID, "NativeFilesCapability", fileParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve native files capability: %w", err)
	}

	res, execErr := fileCap.Execute(ctx, capabilities.CapabilityRequest{
		RequirementID: reqID + "-sub",
		Parameters:    fileParams,
		ContextID:     reqID,
	})
	if execErr != nil {
		return nil, fmt.Errorf("native files execution failed: %w", execErr)
	}
	if !res.Success {
		return nil, fmt.Errorf("native files request failed: %v", res.Error.Message)
	}

	// In a real implementation we would parse res.Data["items"] to extract ".txt" files.
	// For V1, we just return the raw directory list or a mapped version.
	return map[string]interface{}{
		"notes": res.Data["items"],
	}, nil
}

func (c *Capability) normalizeResult(reqID string, start time.Time, operation string, data map[string]interface{}) capabilities.CapabilityResult {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       true,
		Realization:   capabilities.Deterministic,
		ResponseType:  "notes",
		Operation:     operation,
		Data:          data,
		Duration:      time.Since(start),
	}
}

func (c *Capability) normalizeError(reqID string, start time.Time, code string, err error) (capabilities.CapabilityResult, error) {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       false,
		Error: &capabilities.CapabilityError{
			Code:    code,
			Message: err.Error(),
			Retry:   false,
		},
		Duration:      time.Since(start),
	}, nil
}
