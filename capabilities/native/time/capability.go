package time

import (
	"context"
	"time"

	"idun/capabilities"
	coretime "idun/core/time"
)

// TimeCapability provides access to the Time Core Service.
// It exposes accurate temporal operations without any cognitive logic.
type TimeCapability struct {
	capabilities.BaseCapability
	timeSvc coretime.TimeService
}

// New creates a new instance of the Time Capability.
func New(timeSvc coretime.TimeService) *TimeCapability {
	meta := capabilities.CapabilityMetadata{
		Name:        "TimeService",
		Category:    capabilities.CategorySystem,
		Description: "Provides accurate system time and date",
		Version:     "1.0.0",
		Author:      "IDUN Core",
		Tags:        []string{"time", "native", "system"},
	}

	return &TimeCapability{
		BaseCapability: capabilities.NewBaseCapability("sys-time-1", meta),
		timeSvc:        timeSvc,
	}
}

// Execute fulfills the Capability interface by calling the Time Core Service.
func (c *TimeCapability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	start := time.Now()

	if c.timeSvc == nil {
		return capabilities.CapabilityResult{
			RequirementID: req.RequirementID,
			Success:       false,
			Error:         &capabilities.CapabilityError{Code: "INTERNAL_ERROR", Message: "TimeService is not wired", Retry: false},
			Duration:      time.Since(start),
		}, nil // Note: execution returns success=false rather than a hard Go error for capability failures
	}

	// For the initial V1 capability, we just return the current time.
	// Future iterations could parse parameters (e.g., req.Parameters["operation"] = "today").
	now := c.timeSvc.Now()
	
	data := map[string]interface{}{
		"CurrentTime": now,
		"Timezone":    now.Location().String(),
	}

	responseType := "time"
	if pi, ok := ctx.Value("planIntent").(string); ok && pi == "query_date" {
		responseType = "date"
	}

	return capabilities.CapabilityResult{
		RequirementID: req.RequirementID,
		Success:       true,
		Realization:   capabilities.Deterministic,
		ResponseType:  responseType,
		Data:          data,
		Duration:      time.Since(start),
	}, nil
}
