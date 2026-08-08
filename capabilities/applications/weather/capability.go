package weather

import (
	"context"
	"encoding/json"
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

	// 1. Validation of envelope
	if req.RequirementID == "" {
		return c.normalizeError(req.RequirementID, start, "Validation", errors.New("missing requirement ID"))
	}

	// 2. Lifecycle Check
	if err := c.checkLifecycle(); err != nil {
		return c.normalizeError(req.RequirementID, start, "Unavailable", err)
	}

	// 3. Binding & Validation
	typedReq, err := BindWeatherRequest(req.Parameters)
	if err != nil {
		return c.normalizeError(req.RequirementID, start, "Validation", err)
	}

	// 4. Execution via Native Capability Orchestration
	data, execErr := c.executeWeatherRequest(ctx, req.RequirementID, typedReq)

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

func (c *Capability) executeWeatherRequest(ctx context.Context, reqID string, req WeatherRequest) (map[string]interface{}, error) {
	if c.Resolver == nil {
		return nil, errors.New("native capability resolver is not wired")
	}

	location := req.Location
	
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
	// Note: We use "NativeNetworkCapability" as it is the canonical Name.
	netCap, err := c.Resolver.Resolve(ctx, reqID, "NativeNetworkCapability", networkParams)
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

	// Parse the JSON response from the network capability to extract semantic facts
	type wttrDesc struct {
		Value string `json:"value"`
	}
	type wttrCondition struct {
		TempC         string       `json:"temp_C"`
		TempF         string       `json:"temp_F"`
		FeelsLikeC    string       `json:"FeelsLikeC"`
		FeelsLikeF    string       `json:"FeelsLikeF"`
		Humidity      string       `json:"humidity"`
		WindspeedKmph string       `json:"windspeedKmph"`
		Winddir16Point string      `json:"winddir16Point"`
		Visibility    string       `json:"visibility"`
		WeatherDesc   []wttrDesc   `json:"weatherDesc"`
	}
	type wttrResponse struct {
		CurrentCondition []wttrCondition `json:"current_condition"`
	}

	var rawBytes []byte
	if b, ok := res.Data["body"].([]byte); ok {
		rawBytes = b
	} else if s, ok := res.Data["body"].(string); ok {
		rawBytes = []byte(s)
	}

	var semanticWeather map[string]interface{}

	if len(rawBytes) > 0 {
		var apiResp wttrResponse
		if err := json.Unmarshal(rawBytes, &apiResp); err == nil && len(apiResp.CurrentCondition) > 0 {
			cond := apiResp.CurrentCondition[0]
			desc := "Unknown"
			if len(cond.WeatherDesc) > 0 {
				desc = cond.WeatherDesc[0].Value
			}
			semanticWeather = map[string]interface{}{
				"TemperatureC":  cond.TempC,
				"TemperatureF":  cond.TempF,
				"FeelsLikeC":    cond.FeelsLikeC,
				"FeelsLikeF":    cond.FeelsLikeF,
				"Condition":     desc,
				"Humidity":      cond.Humidity,
				"WindSpeedKmph": cond.WindspeedKmph,
				"WindDirection": cond.Winddir16Point,
				"VisibilityKm":  cond.Visibility,
			}
		}
	}

	if semanticWeather == nil {
		semanticWeather = map[string]interface{}{"Error": "No weather data retrieved"}
	}
	
	return map[string]interface{}{
		"location": location,
		"weather":  semanticWeather,
	}, nil
}

func (c *Capability) normalizeResult(reqID string, start time.Time, operation string, data map[string]interface{}) capabilities.CapabilityResult {
	return capabilities.CapabilityResult{
		RequirementID: reqID,
		Success:       true,
		Realization:   capabilities.Deterministic, // Weather uses deterministic template now
		ResponseType:  "weather",
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
			Retry:   code == "Unavailable" || code == "Execution", // Network errors can be retried
		},
		Duration:      time.Since(start),
	}, nil
}
