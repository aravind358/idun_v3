package files

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"idun/capabilities"
	"idun/capabilities/applications/core"
	nativefiles "idun/capabilities/native/files"
)

// Metadata returns the capability definition.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "FilesManager",
		Category:                capabilities.CategoryFiles,
		Description:             "Manages files and directories in the workspace sandbox.",
		Version:                 "1.0.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		Author:                  "IDUN Core",
		Tags:                    []string{"files", "app"},
	}
}

// Capability provides workspace-bound file operations.
type Capability struct {
	core.AppCapability
	workspaceRoot string
}

// New creates a new Files application capability.
func New(resolver core.NativeCapabilityResolver, workspaceRoot string) *Capability {
	return &Capability{
		AppCapability: core.NewAppCapability("app-files-1", Metadata(), resolver),
		workspaceRoot: workspaceRoot,
	}
}

// WorkspaceResolver ensures paths are normalized, canonical, and resolved relative to the workspace.
type WorkspaceResolver struct {
	root string
}

func (r *WorkspaceResolver) Resolve(path string) (string, error) {
	// Must not reject based on authorization, only resolve structure.
	if path == "" {
		return r.root, nil
	}
	
	cleanPath := filepath.Clean(path)
	
	// If it's absolute, return it as is (PermissionPolicy will reject if it's out of bounds)
	if filepath.IsAbs(cleanPath) {
		return filepath.Clean(cleanPath), nil
	}
	
	// Resolve relative to workspace root
	return filepath.Join(r.root, cleanPath), nil
}

// PermissionPolicy evaluates the security policy on a resolved path.
type PermissionPolicy struct {
	workspaceRoot string
}

func (p *PermissionPolicy) IsAllowed(resolvedPath string, op string) error {
	// Normalize workspace root for exact prefix matching
	normRoot := filepath.Clean(p.workspaceRoot)
	normPath := filepath.Clean(resolvedPath)
	
	// Ensure the path is within the workspace bounds
	if !strings.HasPrefix(strings.ToLower(normPath), strings.ToLower(normRoot)) {
		return fmt.Errorf("security policy violation: path %q is outside the workspace root %q", resolvedPath, p.workspaceRoot)
	}

	// Reject dangerous operations
	dangerousOps := []string{"delete everything", "delete all files", "remove all folders", "format drive", "erase disk"}
	for _, d := range dangerousOps {
		if strings.Contains(strings.ToLower(resolvedPath), d) || strings.ToLower(op) == d {
			return fmt.Errorf("security policy violation: dangerous operation %q is blocked", op)
		}
	}
	
	return nil
}

// Execute fulfills the capability execution request.
func (c *Capability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	start := time.Now()

	// 1. Parameter extraction
	operation := req.Parameters["operation"]
	filename := req.Parameters["filename"]
	directory := req.Parameters["directory"]
	source := req.Parameters["source"]
	destination := req.Parameters["destination"]
	pathParam := req.Parameters["path"]

	// Decide target path based on operation semantics
	targetPath := ""
	if pathParam != "" {
		targetPath = pathParam
	} else if filename != "" {
		targetPath = filename
	} else if directory != "" {
		targetPath = directory
	} else if source != "" {
		targetPath = source
	}

	// For dangerous inputs that bypass grammar extraction directly into operation
	dangerousOps := []string{"delete everything", "delete all files", "remove all folders", "format drive", "erase disk"}
	for _, d := range dangerousOps {
		if strings.ToLower(operation) == d || strings.ToLower(targetPath) == d {
			return capabilities.CapabilityResult{
				RequirementID: req.RequirementID,
				Success:       false,
				Error: &capabilities.CapabilityError{
					Code:    "SecurityViolation",
					Message: fmt.Sprintf("security policy violation: dangerous operation %q is blocked", operation),
				},
				Duration: time.Since(start),
			}, nil
		}
	}

	// 2. Workspace Resolution
	resolver := &WorkspaceResolver{root: c.workspaceRoot}
	resolvedPath, err := resolver.Resolve(targetPath)
	if err != nil {
		return capabilities.CapabilityResult{}, err
	}
	
	var resolvedDest string
	if destination != "" {
		resolvedDest, err = resolver.Resolve(destination)
		if err != nil {
			return capabilities.CapabilityResult{}, err
		}
	}

	// 3. Permission Policy Check
	policy := &PermissionPolicy{workspaceRoot: c.workspaceRoot}
	if err := policy.IsAllowed(resolvedPath, operation); err != nil {
		return capabilities.CapabilityResult{
			RequirementID: req.RequirementID,
			Success:       false,
			Error: &capabilities.CapabilityError{
				Code:    "SecurityViolation",
				Message: err.Error(),
			},
			Duration: time.Since(start),
		}, nil
	}
	
	if destination != "" {
		if err := policy.IsAllowed(resolvedDest, operation); err != nil {
			return capabilities.CapabilityResult{
				RequirementID: req.RequirementID,
				Success:       false,
				Error: &capabilities.CapabilityError{
					Code:    "SecurityViolation",
					Message: err.Error(),
				},
				Duration: time.Since(start),
			}, nil
		}
	}

	// 4. Map Cognitive Semantics to Native Operations
	var nativeOp nativefiles.FileOperation
	switch strings.ToLower(operation) {
	case "open", "read":
		nativeOp = nativefiles.OperationReadText
	case "delete", "remove":
		nativeOp = nativefiles.OperationDeleteFile
	case "move":
		nativeOp = nativefiles.OperationMoveFile
	case "copy":
		nativeOp = nativefiles.OperationCopyFile
	case "rename":
		nativeOp = nativefiles.OperationRenameFile
	case "create", "make", "mkdir", "create_dir", "set":
		// Distinguish directory from file creation based on targetPath context if possible
		// Default to create directory if "folder" or "dir" is in the command, or if it has no extension, but simple fallback is OK
		if strings.Contains(strings.ToLower(req.Parameters["command"]), "file") || filepath.Ext(targetPath) != "" {
			nativeOp = nativefiles.OperationCreateFile
		} else {
			nativeOp = nativefiles.OperationCreateDirectory
		}
	case "create_file":
		nativeOp = nativefiles.OperationCreateFile
	case "write", "update":
		nativeOp = nativefiles.OperationWriteFile
	case "append":
		nativeOp = nativefiles.OperationAppendFile
	case "exists", "check", "check if", "does":
		nativeOp = nativefiles.OperationFileExists

	case "list", "show":
		nativeOp = nativefiles.OperationListDirectory
	default:
		// Try mapping directly if it's already a native operation
		nativeOp = nativefiles.FileOperation(operation)
		if !nativeOp.IsValid() {
			return capabilities.CapabilityResult{}, fmt.Errorf("unsupported operation semantics: %s", operation)
		}
	}

	nativeParams := map[string]string{
		"operation": string(nativeOp),
		"path":      resolvedPath,
	}
	if destination != "" {
		nativeParams["destination"] = resolvedDest
	}

	// 5. Native Files Capability Invocation
	fileCap, err := c.Resolver.Resolve(ctx, req.RequirementID, "NativeFilesCapability", nativeParams)
	if err != nil {
		return capabilities.CapabilityResult{}, fmt.Errorf("failed to resolve native files capability: %w", err)
	}

	nativeReq := capabilities.CapabilityRequest{
		RequirementID: req.RequirementID,
		Parameters:    nativeParams,
	}

	res, err := fileCap.Execute(ctx, nativeReq)
	if err != nil {
		return res, err
	}

	// Normalize operation name for the template
	normalizedOp := strings.ToLower(operation)
	if normalizedOp == "check if" || normalizedOp == "check" || normalizedOp == "does" {
		normalizedOp = "exists"
	} else if normalizedOp == "remove" {
		normalizedOp = "delete"
	}

	// Intercept the native result to enrich it with presentation routing and semantic operation.
	// We do NOT modify res.Data to preserve semantic purity.
	res.Realization = capabilities.Deterministic
	res.ResponseType = "files"
	res.Operation = normalizedOp
	
	return res, nil
}
