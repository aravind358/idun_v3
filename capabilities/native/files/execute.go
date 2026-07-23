package files

import (
	"context"
	"fmt"
	"time"

	"idun/capabilities"
)

// Execute fulfills the Capability interface.
func (c *Capability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	start := time.Now()

	// 1. Validation
	if err := c.validateRequest(req); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Validation", err)
	}

	// 2. Lifecycle Check
	if err := c.checkLifecycle(); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Unavailable", err)
	}

	opStr := req.Parameters["operation"]
	operation := FileOperation(opStr)

	// 3. Permission Check
	if err := c.checkPermission(ctx, req.RequirementID, operation); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Permission", err)
	}

	// 4. Execution Router
	var data map[string]interface{}
	var execErr error

	switch operation {
	case OperationReadFile, OperationReadBytes, OperationReadText, OperationFileExists:
		data, execErr = c.executeRead(ctx, req, operation)
	case OperationCreateFile, OperationWriteFile, OperationAppendFile, OperationCopyFile, OperationMoveFile, OperationRenameFile, OperationDeleteFile, OperationTemporaryFile:
		data, execErr = c.executeWrite(ctx, req, operation)
	case OperationListDirectory, OperationCreateDirectory, OperationDeleteDirectory, OperationTemporaryDirectory:
		data, execErr = c.executeDirectory(ctx, req, operation)
	case OperationFileMetadata, OperationDirectoryMetadata:
		data, execErr = c.executeMetadata(ctx, req, operation)
	case OperationSearchFiles:
		data, execErr = c.executeSearch(ctx, req, operation)
	case OperationCalculateHash:
		data, execErr = c.executeHash(ctx, req, operation)
	default:
		execErr = fmt.Errorf("unknown operation: %s", operation)
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Validation", execErr)
	}

	if execErr != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Execution", execErr)
	}

	c.metrics.RecordSuccess(time.Since(start))
	return c.normalizeResult(req.RequirementID, start, data), nil
}
