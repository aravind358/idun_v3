package network

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
	operation := NetworkOperation(opStr)

	// 3. Permission Check
	if err := c.checkPermission(ctx, req.RequirementID, operation); err != nil {
		c.metrics.RecordFailure(time.Since(start))
		return c.normalizeError(req.RequirementID, start, "Permission", err)
	}

	// 4. Execution Router
	var data map[string]interface{}
	var execErr error

	switch operation {
	case OperationResolveDNS, OperationLookupIP:
		data, execErr = c.executeDNS(ctx, req, operation)
	case OperationHTTPGet, OperationHTTPPost, OperationHTTPHead:
		data, execErr = c.executeHTTP(ctx, req, operation)
	case OperationOpenTCPSocket, OperationOpenUDPSocket, OperationCloseSocket:
		data, execErr = c.executeSocket(ctx, req, operation)
	case OperationListInterfaces, OperationGetInterface, OperationConnectionStatus, OperationPing:
		data, execErr = c.executeInterfaces(ctx, req, operation)
	case OperationDownload, OperationUpload:
		data, execErr = c.executeDownload(ctx, req, operation)
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
