package media

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
	operation := MediaOperation(opStr)

	// 3. Permission Check
	if err := c.checkPermission(ctx, req.RequirementID, operation); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Permission", err)
	}

	// 4. Execution Router
	var data map[string]interface{}
	var execErr error

	switch operation {
	case OperationPlayAudio, OperationStopAudio, OperationPauseAudio, OperationResumeAudio, OperationRecordAudio:
		data, execErr = c.executeAudio(ctx, req, operation)
	case OperationPlayVideo, OperationPauseVideo, OperationResumeVideo, OperationStopVideo, OperationRecordVideo:
		data, execErr = c.executeVideo(ctx, req, operation)
	case OperationCaptureImage, OperationLoadImage, OperationSaveImage:
		data, execErr = c.executeImage(ctx, req, operation)
	case OperationGetMetadata:
		data, execErr = c.executeMetadata(ctx, req, operation)
	case OperationListMediaDevices, OperationGetDevice, OperationListCodecs:
		data, execErr = c.executeDevices(ctx, req, operation)
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
