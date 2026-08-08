package weather

import (
	"errors"
)

// WeatherRequest is the strongly-typed internal representation of a weather request.
type WeatherRequest struct {
	Operation string
	Location  string
	Intent    string
}

// BindWeatherRequest converts generic map parameters into a typed WeatherRequest,
// performing centralized validation, normalization, and default assignment.
func BindWeatherRequest(params map[string]string) (WeatherRequest, error) {
	req := WeatherRequest{
		Operation: params["operation"],
		Location:  params["location"],
		Intent:    params["intent"],
	}

	// 1. Assign defaults
	if req.Operation == "" {
		req.Operation = string(OperationCurrent)
	}
	if req.Location == "" {
		req.Location = "Local"
	}
	if req.Intent == "" {
		req.Intent = "query_weather"
	}

	// 2. Validate
	op := WeatherOperation(req.Operation)
	if !op.IsValid() {
		return req, errors.New("unsupported operation: " + req.Operation)
	}

	return req, nil
}
