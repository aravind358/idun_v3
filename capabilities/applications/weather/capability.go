package weather

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"idun/capabilities"
	"idun/capabilities/applications/core"
)

// Capability defines the weather application capability.
// It is a Model 1 App Capability (Orchestration).
// It MUST invoke the Native Network Capability to perform HTTP requests.
type Capability struct {
	core.AppCapability
}

// New creates a new instance of the Weather Capability.
func New(deps core.AppCapabilityDependencies) *Capability {
	return &Capability{
		AppCapability: core.NewAppCapability("app-weather-1", Metadata(), deps.Resolver),
	}
}

// Execute fulfills the Capability interface.
func (c *Capability) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResult, error) {
	start := time.Now()

	// 1. Validation
	if err := c.validateRequest(req); err != nil {
		return c.normalizeError(req.RequirementID, start, "Validation", err)
	}

	// 2. Lifecycle Check
	if err := c.checkLifecycle(); err != nil {
		return c.normalizeError(req.RequirementID, start, "Unavailable", err)
	}

	opStr := req.Parameters["operation"]
	operation := WeatherOperation(opStr)

	// 3. Execution via Native Capability Orchestration
	data, execErr := c.executeWeatherRequest(ctx, req.RequirementID, operation, req.Parameters)

	if execErr != nil {
		return c.normalizeError(req.RequirementID, start, "Execution", execErr)
	}

	return c.normalizeResult(req.RequirementID, start, data), nil
}

func (c *Capability) validateRequest(req capabilities.CapabilityRequest) error {
	if req.RequirementID == "" {
		return errors.New("missing requirement ID")
	}

	operation := req.Parameters["operation"]
	if operation == "" {
		return errors.New("missing 'operation' parameter")
	}

	op := WeatherOperation(operation)
	if !op.IsValid() {
		return errors.New("unsupported operation: " + operation)
	}

	location := req.Parameters["location"]
	if location == "" {
		return errors.New("missing 'location' parameter")
	}

	return nil
}

func (c *Capability) checkLifecycle() error {
	state := c.State().Lifecycle
	if state == "DISABLED" || state == "UNLOADED" {
		return errors.New("capability is not currently available for execution: " + string(state))
	}
	return nil
}

func (c *Capability) executeWeatherRequest(ctx context.Context, reqID string, operation WeatherOperation, params map[string]string) (map[string]interface{}, error) {
	if c.Resolver == nil {
		return nil, errors.New("native capability resolver is not wired")
	}

	location := params["location"]
	
	// Format URL for wttr.in (example free API)
	// We use the JSON format for programmatic parsing
	apiURL := fmt.Sprintf("https://wttr.in/%s?format=j1", url.QueryEscape(location))

	// Construct parameters for the Native Network Capability
	networkParams := map[string]string{
		"operation": "http_get", // Assuming Native Network has an http_get operation
		"url":       apiURL,
	}

	// Resolve the Native Network Capability
	// Note: We use "sys-net-1" or whatever the canonical ID/Name of the Native Network capability is.
	// We'll use the Name "NetworkCapability" as a placeholder for resolution.
	netCap, err := c.Resolver.Resolve(ctx, reqID, "NetworkCapability", networkParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve native network capability: %w", err)
	}

	// Build the sub-request
	subReq := capabilities.CapabilityRequest{
		RequirementID: reqID + "-sub",
		Parameters:    networkParams,
		ContextID:     reqID, // Pass through context ID if needed
	}

	// Execute the Native Capability
	res, execErr := netCap.Execute(ctx, subReq)
	if execErr != nil {
		return nil, fmt.Errorf("native network execution failed: %w", execErr)
	}
	
	if !res.Success {
		return nil, fmt.Errorf("native network request failed: %v", res.Error.Message)
	}

	// In a complete implementation, we would parse the JSON response from the network capability
	// Here we just wrap the network response for the Generative engine
	
	return map[string]interface{}{
		"location": location,
		"type":     string(operation),
		"raw_data": res.Data,
	}, nil
}

func (c *Capability) normalizeResult(reqID string, start time.Time, data map[string]interface{}) capabilities.CapabilityResult {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       true,
		Realization:   capabilities.Generative, // Weather needs LLM to summarize/format the raw JSON
		ResponseType:  "weather",
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
			Retry:   code == "Unavailable" || code == "Execution", // Network errors can be retried
		},
		Duration:      time.Since(start),
	}, nil
}
