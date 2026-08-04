package weather

import (
	"idun/capabilities"
)

// Metadata returns the immutable definition of the Weather Capability.
func Metadata() capabilities.CapabilityMetadata {
	return capabilities.CapabilityMetadata{
		Name:                    "WeatherApp",
		Category:                capabilities.CategoryExternalServices,
		Description:             "Application capability for fetching weather forecasts.",
		Version:                 "1.0.0",
		APIVersion:              "1.0.0",
		MinimumFrameworkVersion: "1.0.0",
		SupportedPlatforms:      []string{"windows", "linux", "darwin"},
		PermissionCategory:      "Weather",
		ImplementationType:      "Application",
		Author:                  "IDUN Core",
		Tags:                    []string{"application", "weather", "network"},
	}
}
